// Copyright © 2022 Kaleido, Inc.
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

package events

import (
	"github.com/hyperledger/firefly-common/pkg/log"
	"github.com/hyperledger/firefly-transaction-manager/pkg/ffcapi"
)

// blockListener ensures it always pulls blocks from the channel passed to the connector
// for new block events, regardless of whether the downstream confirmations update queue
// is full blocked (likely because the event stream is blocked).
// This is critical to avoid the situation where one blocked stream, stops another stream
// from receiving block events.
// We use the same "GapPotential" flag that the connector can mark on a reconnect, to mark
// when we've had to discard events for a blocked event listener (event listeners could stay
// blocked indefinitely, so we can't leak memory by storing up an indefinite number of new
// block events).
func (es *eventStream) blockListener(startedState *startedStreamState) {
	defer close(startedState.blockListenerDone)
	var blockedUpdate *ffcapi.BlockHashEvent
	for {
		if blockedUpdate != nil {
			select {
			case blockUpdate := <-startedState.blocks:
				// Have to discard this
				blockedUpdate.GapPotential = true // there is a gap for sure at this point
				log.L(startedState.ctx).Debugf("Blocked event stream missed new block event: %v", blockUpdate.BlockHashes)
			case es.confirmations.NewBlockHashes() <- blockedUpdate:
				// We're not blocked any more
				log.L(startedState.ctx).Infof("Event stream block-listener unblocked")
				blockedUpdate = nil
			case <-startedState.ctx.Done():
				log.L(startedState.ctx).Debugf("Block listener exiting (previously blocked)")
				return
			}
		} else {
			select {
			case blockUpdate := <-startedState.blocks:
				log.L(startedState.ctx).Debugf("Received block event: %v", blockUpdate.BlockHashes)
				// Nothing to do unless we have confirmations turned on
				if es.confirmations != nil {
					select {
					case es.confirmations.NewBlockHashes() <- blockUpdate:
						// all good, we passed it on
					default:
						// we can't deliver it immediately, we switch to blocked mode
						log.L(startedState.ctx).Infof("Event stream block-listener became blocked")
						// Take a copy of the block update, so we can modify (to mark a gap) without affecting other streams
						var bu = *blockUpdate
						blockedUpdate = &bu
					}
				}
			case <-startedState.ctx.Done():
				log.L(startedState.ctx).Debugf("Block listener exiting")
				return
			}
		}
	}
}
