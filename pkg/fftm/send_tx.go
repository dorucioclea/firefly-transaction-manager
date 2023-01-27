// Copyright © 2023 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package fftm

import (
	"context"

	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger/firefly-transaction-manager/pkg/apitypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

func (m *manager) sendManagedTransaction(ctx context.Context, request *apitypes.TransactionRequest) (*apitypes.ManagedTX, error) {

	// Prepare the transaction, which will mean we have a transaction that should be submittable.
	// If we fail at this stage, we don't need to write any state as we are sure we haven't submitted
	// anything to the blockchain itself.
	prepared, _, err := m.connector.TransactionPrepare(ctx, &ffcapi.TransactionPrepareRequest{
		TransactionInput: request.TransactionInput,
	})
	if err != nil {
		return nil, err
	}

	return m.submitPreparedTX(ctx, request.Headers.ID, &request.TransactionHeaders, prepared.Gas, prepared.TransactionData)
}

func (m *manager) sendManagedContractDeployment(ctx context.Context, request *apitypes.ContractDeployRequest) (*apitypes.ManagedTX, error) {

	// Prepare the transaction, which will mean we have a transaction that should be submittable.
	// If we fail at this stage, we don't need to write any state as we are sure we haven't submitted
	// anything to the blockchain itself.
	prepared, _, err := m.connector.DeployContractPrepare(ctx, &request.ContractDeployPrepareRequest)
	if err != nil {
		return nil, err
	}

	return m.submitPreparedTX(ctx, request.Headers.ID, &request.TransactionHeaders, prepared.Gas, prepared.TransactionData)
}

func (m *manager) submitPreparedTX(ctx context.Context, txID string, txHeaders *ffcapi.TransactionHeaders, gas *fftypes.FFBigInt, transactionData string) (*apitypes.ManagedTX, error) {

	// The request ID is the primary ID, and should be supplied by the user for idempotence
	if txID == "" {
		txID = fftypes.NewUUID().String()
	}

	// First job is to assign the next nonce to this request.
	// We block any further sends on this nonce until we've got this one successfully into the node, or
	// fail deterministically in a way that allows us to return it.
	lockedNonce, err := m.assignAndLockNonce(ctx, txID, txHeaders.From)
	if err != nil {
		return nil, err
	}
	// We will call markSpent() once we reach the point the nonce has been used
	defer lockedNonce.complete(ctx)

	// Sequencing ID is always generated by us - so we have a deterministic order of transactions
	// Note: We must allocate this within the nonce lock, to ensure that the nonce sequence and the
	//       global transaction sequence line up.
	seqID := apitypes.NewULID()

	// Next we update FireFly core with the pre-submitted record pending record, with the allocated nonce.
	// From this point on, we will guide this transaction through to submission.
	// We return an "ack" at this point, and dispatch the work of getting the transaction submitted
	// to the background worker.
	now := fftypes.Now()
	mtx := &apitypes.ManagedTX{
		ID:                 txID, // on input the request ID must be the namespaced operation ID
		Created:            now,
		Updated:            now,
		SequenceID:         seqID,
		Nonce:              fftypes.NewFFBigInt(int64(lockedNonce.nonce)),
		Gas:                gas,
		TransactionHeaders: *txHeaders,
		TransactionData:    transactionData,
		Status:             apitypes.TxStatusPending,
	}

	m.txhistory.SetSubStatus(ctx, mtx, apitypes.TxSubStatusReceived)
	m.txhistory.AddSubStatusAction(ctx, mtx, apitypes.TxActionAssignNonce, fftypes.JSONAnyPtr(`{"nonce":"`+mtx.Nonce.String()+`"}`), nil)

	if err = m.persistence.WriteTransaction(m.ctx, mtx, true); err != nil {
		return nil, err
	}
	log.L(m.ctx).Infof("Tracking transaction %s at nonce %s / %d", mtx.ID, mtx.TransactionHeaders.From, mtx.Nonce.Int64())
	m.markInflightStale()

	// Ok - we've spent it. The rest of the processing will be triggered off of lockedNonce
	// completion adding this transaction to the pool (and/or the change event that comes in from
	// FireFly core from the update to the transaction)
	lockedNonce.spent = mtx
	return mtx, nil
}
