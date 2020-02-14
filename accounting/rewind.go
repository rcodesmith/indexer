// Copyright (C) 2019-2020 Algorand, Inc.
// This file is part of the Algorand Indexer
//
// Algorand Indexer is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// Algorand Indexer is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Algorand Indexer.  If not, see <https://www.gnu.org/licenses/>.

package accounting

import (
	"context"
	"fmt"

	"github.com/algorand/go-algorand-sdk/client/algod/models"
	"github.com/algorand/go-algorand-sdk/encoding/msgpack"
	atypes "github.com/algorand/go-algorand-sdk/types"

	"github.com/algorand/indexer/idb"
	"github.com/algorand/indexer/types"
)

func assetUpdate(account *models.Account, assetid uint64, amount int64) {
	av := account.Assets[assetid]
	av.Amount = uint64(int64(av.Amount) + amount)
	account.Assets[assetid] = av
}

func AccountAtRound(account models.Account, round uint64, db idb.IndexerDb) (acct models.Account, err error) {
	acct = account
	addr, err := atypes.DecodeAddress(account.Address)
	if err != nil {
		return
	}
	tf := idb.TransactionFilter{
		Address:  addr[:],
		MinRound: round + 1,
		MaxRound: account.Round,
	}
	txns := db.Transactions(context.Background(), tf)
	for txnrow := range txns {
		var stxn types.SignedTxnInBlock
		err = msgpack.Decode(txnrow.TxnBytes, &stxn)
		if err != nil {
			return
		}
		if addr == stxn.Txn.Sender {
			acct.AmountWithoutPendingRewards += uint64(stxn.Txn.Fee)
			acct.AmountWithoutPendingRewards -= uint64(stxn.SenderRewards)
		}
		switch stxn.Txn.Type {
		case atypes.PaymentTx:
			if addr == stxn.Txn.Sender {
				acct.AmountWithoutPendingRewards += uint64(stxn.Txn.Amount)
			}
			if addr == stxn.Txn.Receiver {
				acct.AmountWithoutPendingRewards -= uint64(stxn.Txn.Amount)
				acct.AmountWithoutPendingRewards -= uint64(stxn.ReceiverRewards)
			}
			if addr == stxn.Txn.CloseRemainderTo {
				acct.AmountWithoutPendingRewards += uint64(stxn.ClosingAmount)
				acct.AmountWithoutPendingRewards -= uint64(stxn.CloseRewards)
			}
		case atypes.KeyRegistrationTx:
		case atypes.AssetConfigTx:
			if stxn.Txn.ConfigAsset == 0 {
				// create asset, unwind the application of the value
				// TODO: fetch block_header for txnrow.Round, .TxnCounter find `assetId = txnCounter + uint64(intra) + 1`
				err = fmt.Errorf("%d:%d %s TODO: handle acfg", txnrow.Round, txnrow.Intra, account.Address)
				return
			}
		case atypes.AssetTransferTx:
			if addr == stxn.Txn.AssetSender || addr == stxn.Txn.Sender {
				assetUpdate(&acct, uint64(stxn.Txn.XferAsset), int64(stxn.Txn.AssetAmount))
			}
			if addr == stxn.Txn.AssetReceiver {
				assetUpdate(&acct, uint64(stxn.Txn.XferAsset), -int64(stxn.Txn.AssetAmount))
			}
		case atypes.AssetFreezeTx:
		default:
			panic("unknown txn type")
		}
	}

	// TODO: fetch MaxRound: round, Limit: 1

	tf.MaxRound = round
	tf.MinRound = 0
	tf.Limit = 1

	acct.Round = round
	return
}
