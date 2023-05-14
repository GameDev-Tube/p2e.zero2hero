package db

import "gorm.io/gorm"

type IndexerStatus struct {
	gorm.Model

	CurrentBlock uint64
}

type Prediction struct {
	gorm.Model

	ChoiceA uint64
	ChoiceB uint64
	ChoiceC uint64
	ChoiceD uint64
	ChoiceE uint64
	ChoiceF uint64
	Hash    string

	NftId     uint64
	Confirmed bool
	Score     uint
	Game      uint
}

const (
	// GamePhasePre corresponds to idle state, but is "zero" state after contract deployment.
	GamePhasePre = uint(iota)
	// GamePhasePredicting corresponds to minting state
	GamePhasePredicting
	// GamePhaseActive corresponds to game started event
	GamePhaseActive
	// GamePhaseFinished corresponds to idle state, and indicates next game edition
	GamePhaseFinished
)

type GameState struct {
	gorm.Model

	GamePhase   uint
	GameEdition uint

	OutcomeA uint64
	OutcomeB uint64
	OutcomeC uint64
	OutcomeD uint64
	OutcomeE uint64
	OutcomeF uint64
}

type SwapList struct {
	gorm.Model

	TokenAddress  string
	IsWhitelisted *bool `gorm:"not null"` // ptr so can easily explicitly query for true false or any.
}
