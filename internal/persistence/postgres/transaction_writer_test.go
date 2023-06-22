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

package postgres

import (
	"context"
	"database/sql/driver"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/hyperledger/firefly-common/pkg/fftypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/apitypes"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
	"github.com/stretchr/testify/assert"
)

func TestExecuteBatchOpsInsertBadOp(t *testing.T) {
	ctx, p, _, done := newMockSQLPersistence(t)
	defer done()

	txOp := &transactionOperation{
		txID: "1",
		txInsert: &apitypes.ManagedTX{
			TransactionHeaders: ffcapi.TransactionHeaders{From: "" /* missing */},
		},
		done: make(chan error, 1),
	}
	p.writer.queue(ctx, txOp)
	err := txOp.flush(ctx)
	assert.Regexp(t, "FF21086", err)
}

func TestExecuteBatchOpsInsertTXFailWrapped(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectQuery("SELECT.*").WillReturnRows(sqlmock.NewRows([]string{"seq"}))
	mdb.ExpectExec("INSERT.*").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	p.writer.runBatch(ctx, &transactionWriterBatch{
		ops: []*transactionOperation{
			{
				txID: "1",
				txInsert: &apitypes.ManagedTX{
					TransactionHeaders: ffcapi.TransactionHeaders{From: "0x12345"},
				},
				nextNonceCB: func(ctx context.Context, signer string) (uint64, error) { return 0, nil },
				done:        make(chan error, 1)},
		},
	})

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsInsertCacheExpiredTXNextNonceFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	oldTime := fftypes.FFTime(time.Now().Add(-10000 * time.Hour))
	p.writer.nextNonceCache.Add("0x12345", &nonceCacheEntry{
		cachedTime: &oldTime,
	})

	mdb.ExpectBegin()
	mdb.ExpectRollback()

	called := make(chan struct{})
	op := &transactionOperation{
		txID: "1",
		txInsert: &apitypes.ManagedTX{
			TransactionHeaders: ffcapi.TransactionHeaders{From: "0x12345"},
		},
		nextNonceCB: func(ctx context.Context, signer string) (uint64, error) {
			close(called)
			return 0, fmt.Errorf("pop")
		},
		done: make(chan error, 1),
	}
	p.writer.runBatch(ctx, &transactionWriterBatch{
		ops: []*transactionOperation{op},
	})

	err := op.flush(ctx)
	assert.Regexp(t, "FF21084", err)
	<-called

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsInsertTXFailQueryExistingNonce(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectQuery("SELECT.*").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	p.writer.runBatch(ctx, &transactionWriterBatch{
		ops: []*transactionOperation{
			{
				txID: "1",
				txInsert: &apitypes.ManagedTX{
					TransactionHeaders: ffcapi.TransactionHeaders{From: "0x12345"},
				},
				nextNonceCB: func(ctx context.Context, signer string) (uint64, error) { return 0, nil },
				done:        make(chan error, 1)},
		},
	})

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsInsertTXFailOverrideNonceBelowTx(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectQuery("SELECT.*").WillReturnRows(newTXRow(p))
	mdb.ExpectExec("INSERT.*").WillReturnResult(driver.ResultNoRows)
	mdb.ExpectCommit()

	tx := &apitypes.ManagedTX{
		TransactionHeaders: ffcapi.TransactionHeaders{From: "0x12345"},
	}
	p.writer.runBatch(ctx, &transactionWriterBatch{
		ops: []*transactionOperation{
			{
				txID:     "1",
				txInsert: tx,
				nextNonceCB: func(ctx context.Context, signer string) (uint64, error) {
					return 1 /* below nonce 11111 in the row queried back */, nil
				},
				done: make(chan error, 1)},
		},
	})
	assert.Equal(t, uint64(0x11112), tx.Nonce.Uint64())

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsInsertTXFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectQuery("SELECT.*transactions").WillReturnRows(sqlmock.NewRows([]string{}))
	mdb.ExpectExec("INSERT.*").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	err := p.db.RunAsGroup(ctx, func(ctx context.Context) error {
		return p.writer.executeBatchOps(ctx, &transactionWriterBatch{
			txInsertsByFrom: map[string][]*transactionOperation{
				"0x12345": {{
					txID:        "111222333",
					txInsert:    &apitypes.ManagedTX{},
					nextNonceCB: func(ctx context.Context, signer string) (uint64, error) { return 1, nil },
				}},
			},
		})
	})
	assert.Regexp(t, "FF00177", err)

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsIdempotencyPreCheckFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	p.writer.txMetaCache.Add("111222333", &txCacheEntry{})

	mdb.ExpectBegin()
	mdb.ExpectQuery("SELECT.*transactions").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	err := p.db.RunAsGroup(ctx, func(ctx context.Context) error {
		return p.writer.executeBatchOps(ctx, &transactionWriterBatch{
			txInsertsByFrom: map[string][]*transactionOperation{
				"0x12345": {{
					txID:        "111222333",
					txInsert:    &apitypes.ManagedTX{},
					nextNonceCB: func(ctx context.Context, signer string) (uint64, error) { return 1, nil },
				}},
			},
		})
	})
	assert.Regexp(t, "FF00176.*pop", err)

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsUpdateTXFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectExec("UPDATE.*").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	err := p.writer.executeBatchOps(ctx, &transactionWriterBatch{
		txUpdates: []*transactionOperation{{
			txUpdate: &apitypes.TXUpdates{},
		}},
	})
	assert.Regexp(t, "FF00178", err)

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsUpsertReceiptFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectExec("INSERT.*").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	err := p.db.RunAsGroup(ctx, func(ctx context.Context) error {
		return p.writer.executeBatchOps(ctx, &transactionWriterBatch{
			receiptInserts: map[string]*apitypes.ReceiptRecord{
				"tx1": {TransactionReceiptResponse: &ffcapi.TransactionReceiptResponse{}},
			},
		})
	})
	assert.Regexp(t, "FF00176", err)

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsInsertTXHistoryFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectExec("INSERT.*").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	err := p.writer.executeBatchOps(ctx, &transactionWriterBatch{
		historyInserts: []*apitypes.TXHistoryRecord{{}},
	})
	assert.Regexp(t, "FF00177", err)

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsDeleteConfirmationsFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectExec("DELETE.*").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	err := p.writer.executeBatchOps(ctx, &transactionWriterBatch{
		confirmationResets: map[string]bool{"1": true},
	})
	assert.Regexp(t, "FF00179", err)

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsInsertConfirmationFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectExec("INSERT.*").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	err := p.writer.executeBatchOps(ctx, &transactionWriterBatch{
		confirmationInserts: []*apitypes.ConfirmationRecord{{
			Confirmation: &apitypes.Confirmation{},
		}},
	})
	assert.Regexp(t, "FF00177", err)

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsDeleteTXFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectExec("DELETE.*receipts").WillReturnResult(driver.RowsAffected(0))
	mdb.ExpectExec("DELETE.*confirmations").WillReturnResult(driver.RowsAffected(0))
	mdb.ExpectExec("DELETE.*txhistory").WillReturnResult(driver.RowsAffected(0))
	mdb.ExpectExec("DELETE.*transactions").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	err := p.db.RunAsGroup(ctx, func(ctx context.Context) error {
		return p.writer.executeBatchOps(ctx, &transactionWriterBatch{
			txDeletes: []string{"1"},
		})
	})
	assert.Regexp(t, "FF00179", err)

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsDeleteTXHistoryFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectExec("DELETE.*receipts").WillReturnResult(driver.RowsAffected(0))
	mdb.ExpectExec("DELETE.*confirmations").WillReturnResult(driver.RowsAffected(0))
	mdb.ExpectExec("DELETE.*txhistory").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	err := p.db.RunAsGroup(ctx, func(ctx context.Context) error {
		return p.writer.executeBatchOps(ctx, &transactionWriterBatch{
			txDeletes: []string{"1"},
		})
	})
	assert.Regexp(t, "FF00179", err)

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsDeleteTXConfirmationsFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectExec("DELETE.*receipts").WillReturnResult(driver.RowsAffected(0))
	mdb.ExpectExec("DELETE.*confirmations").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	err := p.db.RunAsGroup(ctx, func(ctx context.Context) error {
		return p.writer.executeBatchOps(ctx, &transactionWriterBatch{
			txDeletes: []string{"1"},
		})
	})
	assert.Regexp(t, "FF00179", err)

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestExecuteBatchOpsDeleteTXReceiptFail(t *testing.T) {
	ctx, p, mdb, done := newMockSQLPersistence(t)
	defer done()

	mdb.ExpectBegin()
	mdb.ExpectExec("DELETE.*receipts").WillReturnError(fmt.Errorf("pop"))
	mdb.ExpectRollback()

	err := p.db.RunAsGroup(ctx, func(ctx context.Context) error {
		return p.writer.executeBatchOps(ctx, &transactionWriterBatch{
			txDeletes: []string{"1"},
		})
	})
	assert.Regexp(t, "FF00179", err)

	assert.NoError(t, mdb.ExpectationsWereMet())
}

func TestFlushOpClosedContext(t *testing.T) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	cancelCtx()
	err := newTransactionOperation("tx1").flush(ctx)
	assert.Regexp(t, "FF00154", err)
}

func TestQueueClosedBGContext(t *testing.T) {
	_, p, _, done := newMockSQLPersistence(t)
	done()
	p.writer.workQueues = []chan *transactionOperation{make(chan *transactionOperation)}
	p.writer.workerCount = 1

	op := newTransactionOperation("tx1")
	p.writer.queue(context.Background(), op)
	err := op.flush(context.Background())
	assert.Regexp(t, "FF21083", err)

}

func TestQueueClosedContext(t *testing.T) {
	_, p, _, done := newMockSQLPersistence(t)
	done()
	p.writer.workQueues = []chan *transactionOperation{make(chan *transactionOperation)}
	p.writer.workerCount = 1
	p.writer.bgCtx = context.Background()

	closedCtx, cancelCtx := context.WithCancel(context.Background())
	cancelCtx()
	p.writer.queue(closedCtx, newTransactionOperation("tx1"))

}
