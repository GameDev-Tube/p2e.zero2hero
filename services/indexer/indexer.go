package indexer

import (
	"context"
	"fmt"
	"github.com/GameDev-Tube/famemma.PRE2E/services/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"strings"
	"sync"
	"time"
)

// keep N blocks behind to avoid problems with reorgs etc.
const blocksBehind = 1
const sleepTimeWhenCoughtUpMS = 1500

func (i *Indexer) Run() {
	for {
		if !i.catchingUp {
			time.Sleep(sleepTimeWhenCoughtUpMS * time.Millisecond)
		}
		i.tick()
	}
}

func (i *Indexer) tick() {
	if i.blk == nil {
		b, err := i.db.GetIndexerBlock()
		if err != nil {
			fmt.Printf("failed to run tick: %s\n", err.Error())
			return
		}
		i.blk = &b
	}
	b := *i.blk
	var currentBlk uint64
	var err error
	tout(func(c context.Context) {
		currentBlk, err = i.rpc.BlockNumber(c)
	})
	if err != nil {
		fmt.Printf("failed to run tick: %s\n", err.Error())
		return
	}
	var blk *types.Block
	tout(func(c context.Context) {
		blk, err = i.rpc.BlockByNumber(c, big.NewInt(0).SetUint64(b))
	})
	if err != nil {
		fmt.Printf("failed to run tick: %s\n", err.Error())
		return
	}

	if currentBlk-blocksBehind <= b {
		return
	}

	relevantTransactions := make([]common.Hash, 0)
	for _, tx := range blk.Transactions() {
		if tx.To() == nil {
			continue //contract creation
		}
		if strings.EqualFold(tx.To().String(), i.addr.String()) {
			relevantTransactions = append(relevantTransactions, tx.Hash())
		}
	}

	err = i.db.Transact(func(d *db.Db) error {
		if len(relevantTransactions) > 0 {
			err = i.processTxns(relevantTransactions, d)
		}
		b++
		err = i.db.UpdateIndexerBlock(b)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		fmt.Printf("failed to run tick: %s\n", err.Error())
		return
	}

	if currentBlk-(blocksBehind+5) > b {
		if time.Now().Add(-time.Second * 30).After(i.lastCatchupMsg) {
			i.lastCatchupMsg = time.Now()
			fmt.Printf("catching up, %d blocks behind (at %d out of %d)...\n", currentBlk-b-blocksBehind, b, currentBlk-blocksBehind)
		}
	}

	i.blk = &b
}

// / processTxns takes hashes and parses logs of contract. It is used to detect owner's operations for game phase, and
// / mints done by users.
func (i *Indexer) processTxns(txhs []common.Hash, dbTx *db.Db) error {
	receipts := make([]*types.Receipt, len(txhs))
	errs := make([]error, len(txhs))
	wg := sync.WaitGroup{}
	for ii := range txhs {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			tout(func(c context.Context) {
				receipts[idx], errs[idx] = i.rpc.TransactionReceipt(c, txhs[idx])
			})
		}(ii)
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	for _, rcpt := range receipts {
		fmt.Printf("processing tx %s\n", rcpt.TxHash.String())
		for _, log := range rcpt.Logs {
			if log == nil {
				continue
			}
			state, err := i.db.GetGameState()
			if err != nil {
				return err
			}
			switch log.Topics[0] {
			case i.events.NewPred:
				fmt.Println("processing new prediction event")
				if state.GamePhase != db.GamePhasePredicting {
					continue
				}
				event, err := i.binding.ParseNewPrediction(*log)
				if err != nil {
					return err
				}
				err = dbTx.ConfirmHash(hexutil.Encode(event.Hash[:]), event.NftId.Uint64())
				if err != nil {
					return err
				}
			case i.events.GameStart:
				fmt.Println("processing game start event")
				state.GamePhase = db.GamePhaseActive
				err = dbTx.UpdateGameState(state)
				if err != nil {
					return err
				}
			case i.events.MintingStart:
				fmt.Println("processing minting start event")
				if state.GameEdition == 0 && state.GamePhase == db.GamePhasePre {
					state.GamePhase = db.GamePhasePredicting
					err = dbTx.UpdateGameState(state)
				} else {
					// not first edition
					err = i.db.BumpGameEdition(state)
				}

				//state.GameEdition == 0 && state.GamePhase == GamePhasePre
				if err != nil {
					return err
				}
			case i.events.IdleStart:
				fmt.Println("processing idle start event")
				state.GamePhase = db.GamePhaseFinished
				err = dbTx.UpdateGameState(state)
				if err != nil {
					return err
				}
			case i.events.SwapListChange:
				fmt.Println("processing swap list change")
				event, err := i.binding.ParseTokenWhitelistChange(*log)
				if err != nil {
					return err
				}
				err = i.db.SetTokenSwapPermitted(event.Token, event.Permitted)
				if err != nil {
					return err
				}

			default:
				continue
			}
		}
	}
	return nil
}

func tout(cb func(c context.Context)) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*4)
	defer cancel()
	cb(ctx)
}
