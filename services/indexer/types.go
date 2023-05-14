package indexer

import (
	"errors"
	"fmt"
	"github.com/GameDev-Tube/famemma.PRE2E/bindings"
	"github.com/GameDev-Tube/famemma.PRE2E/services/db"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"time"
)

const (
	eventKeyNewPrediction      = "NewPrediction"
	eventKeyStateChangeMinting = "PhaseMintingStart"
	eventKeyStateChangeGame    = "PhaseGameStart"
	eventKeyStateChangeIdle    = "PhaseIdleStart"
	eventSwapListChange        = "TokenWhitelistChange"
)

type events struct {
	NewPred        common.Hash
	MintingStart   common.Hash
	GameStart      common.Hash
	IdleStart      common.Hash
	SwapListChange common.Hash
}

type Indexer struct {
	db             *db.Db
	rpc            *ethclient.Client
	binding        *bindings.P2EContract
	events         events
	addr           common.Address
	blk            *uint64
	catchingUp     bool
	lastCatchupMsg time.Time
}

func New(db *db.Db, p2eAddr, rpcUrl string) (*Indexer, *ethclient.Client, *bindings.P2EContract, error) {
	rpc, err := ethclient.Dial(rpcUrl)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to open RPC connection: %w", err)
	}

	addr := common.HexToAddress(p2eAddr)
	bind, err := bindings.NewP2EContract(addr, rpc)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load P2E binding: %w", err)
	}

	abi, err := bindings.P2EContractMetaData.GetAbi()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load P2E ABI: %w", err)
	}

	eventNewPred, ok1 := abi.Events[eventKeyNewPrediction]
	eventMintingStart, ok2 := abi.Events[eventKeyStateChangeMinting]
	eventGameStart, ok3 := abi.Events[eventKeyStateChangeGame]
	eventIdleStart, ok4 := abi.Events[eventKeyStateChangeIdle]
	eventSwapList, ok5 := abi.Events[eventSwapListChange]
	if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 {
		return nil, nil, nil, errors.New("failed to preload event id, some event does not exist in binding")
	}

	return &Indexer{
		db:      db,
		rpc:     rpc,
		binding: bind,
		addr:    addr,
		events: events{
			NewPred:        eventNewPred.ID,
			MintingStart:   eventMintingStart.ID,
			GameStart:      eventGameStart.ID,
			IdleStart:      eventIdleStart.ID,
			SwapListChange: eventSwapList.ID,
		},
		catchingUp:     true,
		lastCatchupMsg: time.Now(),
	}, rpc, bind, nil
}
