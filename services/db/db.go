package db

import (
	"errors"
	"fmt"
	"github.com/GameDev-Tube/famemma.PRE2E/utils"
	"github.com/ethereum/go-ethereum/common"
	"gorm.io/gorm"
)

func (d *Db) Transact(callback func(*Db) error) error {
	return d.orm.Transaction(func(tx *gorm.DB) error {
		return callback(&Db{
			orm:        tx,
			startBlock: d.startBlock,
		})
	})
}

func (d *Db) GetIndexerBlock() (uint64, error) {
	m, err := d.getIndexerState()
	return m.CurrentBlock, err
}

func (d *Db) getIndexerState() (*IndexerStatus, error) {
	m := &IndexerStatus{}
	err := d.orm.Where(&IndexerStatus{Model: gorm.Model{ID: 1}}).Find(m).Error
	if err == nil && m.CurrentBlock > 0 {
		return m, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || m.CurrentBlock == 0 {
		m.CurrentBlock = d.startBlock
	}
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}
	err = d.orm.Create(m).Error
	if err != nil {
		return nil, fmt.Errorf("failed to init state: %w", err)
	}
	return m, nil
}

func (d *Db) UpdateIndexerBlock(blk uint64) error {
	m, err := d.getIndexerState()
	if err != nil {
		return err
	}
	m.CurrentBlock = blk
	return d.orm.Save(m).Error
}

func (d *Db) ConfirmHash(hash string, nftID uint64) error {
	state, err := d.GetGameState()
	if err != nil {
		return err
	}
	if state.GamePhase != GamePhasePredicting {
		return errors.New("cannot create prediction - not predicting game state")
	}
	m := &Prediction{Hash: hash, Game: state.GameEdition}
	err = d.orm.Where(&m).Find(&m).Error
	if err == nil && m.Model.ID > 0 {
		fmt.Printf("confirming prediction id %d\n", m.Model.ID)
		m.Confirmed = true
		m.NftId = nftID
		err = d.orm.Save(m).Error
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // ignore unknown. Someone minted without registering prediction
		}
	}
	return err
}

// Leaderboards will select newest game that can have leaderboards, which is current ongoing game or finished game, or
// if current round is in predicting state, previous round
func (d *Db) Leaderboards(offset, limit int) ([]Prediction, error) {
	state, err := d.GetGameState()
	if err != nil {
		return nil, err
	}
	if state.GamePhase == GamePhasePre || state.GamePhase == GamePhasePredicting {
		if state.GameEdition == 0 {
			return make([]Prediction, 0), nil
		}
		state, err = d.getGameState(utils.Ptr(state.GameEdition - 1))
	}
	if err != nil && err == gorm.ErrRecordNotFound {
		return make([]Prediction, 0), nil // no games happened since system deploy, leaderboards are simply empty.
	}
	if err != nil {
		return nil, err
	}
	predictions := make([]Prediction, 0)
	err = d.orm.Model(&predictions).Where(&Prediction{
		Confirmed: true,
		Game:      state.GameEdition,
	}).Order("score DESC").Limit(limit).Offset(offset).Find(&predictions).Error

	if err != nil && err == gorm.ErrRecordNotFound {
		return make([]Prediction, 0), nil
	}
	return predictions, err
}

func (d *Db) CreatePrediction(pa, pb, pc, pd, pe, pf uint64, round uint, hash string) error {
	state, err := d.GetGameState()
	if err != nil {
		return err
	}
	if state.GamePhase != GamePhasePredicting {
		return errors.New("cannot create prediction - not predicting game state")
	}
	p := &Prediction{
		Model:   gorm.Model{},
		ChoiceA: pa,
		ChoiceB: pb,
		ChoiceC: pc,
		ChoiceD: pd,
		ChoiceE: pe,
		ChoiceF: pf,
		Hash:    hash,

		Confirmed: false,
		Score:     0,
		NftId:     0,
		Game:      round,
	}

	return d.orm.Create(p).Error
}

// GetPredictions is meant for processing, not presentation, it doesn't order by score but by ID, it doesn't select game
// edition automatically, etc.
func (d *Db) GetPredictions(limit, offset int, game uint) ([]Prediction, error) {
	predictions := make([]Prediction, 0)
	err := d.orm.Model(&predictions).Where(&Prediction{
		Confirmed: true,
		Game:      game,
	}).Order("id ASC").Limit(limit).Offset(offset).Find(&predictions).Error

	if err != nil && err == gorm.ErrRecordNotFound {
		return make([]Prediction, 0), nil
	}
	return predictions, err
}

func (d *Db) UpdatePrediction(p *Prediction) error {
	return d.orm.Save(p).Error
}

// BumpGameEdition should be ran using transaction
func (d *Db) BumpGameEdition(state *GameState) error {
	state.GamePhase = GamePhaseFinished
	err := d.orm.Save(state).Error
	if err != nil {
		return err
	}
	return d.orm.Create(&GameState{
		GamePhase:   GamePhasePredicting,
		GameEdition: state.GameEdition + 1,
		OutcomeA:    0,
		OutcomeB:    0,
		OutcomeC:    0,
		OutcomeD:    0,
		OutcomeE:    0,
		OutcomeF:    0,
	}).Error
}

func (d *Db) UpdateGameState(state *GameState) error {
	return d.orm.Save(state).Error
}

func (d *Db) GetGameState() (*GameState, error) {
	return d.getGameState(nil)
}
func (d *Db) getGameState(edition *uint) (*GameState, error) {
	state := &GameState{}
	if edition != nil {
		state.GameEdition = *edition
	}
	err := d.orm.Where(state).Order("game_edition DESC").First(state).Error
	if err == gorm.ErrRecordNotFound {
		state = d.defaultGameState()
		err = d.orm.Create(state).Error
		if err != nil {
			return nil, err
		}
	}
	return state, err
}

func (d *Db) defaultGameState() *GameState {
	return &GameState{
		GamePhase:   GamePhasePre,
		GameEdition: 0,
		OutcomeA:    0,
		OutcomeB:    0,
		OutcomeC:    0,
		OutcomeD:    0,
		OutcomeE:    0,
		OutcomeF:    0,
	}
}

func (d *Db) GetSwappableTokens() ([]common.Address, error) {
	models := make([]SwapList, 0)
	err := d.orm.Where(&SwapList{IsWhitelisted: utils.Ptr(true)}).Find(&models).Error
	if (err != nil && err == gorm.ErrRecordNotFound) || len(models) == 0 {
		return make([]common.Address, 0), nil
	}
	if err != nil {
		return nil, err
	}
	out := make([]common.Address, len(models))
	for i := range models {
		out[i] = common.HexToAddress(models[i].TokenAddress)
	}
	return out, nil
}

func (d *Db) SetTokenSwapPermitted(token common.Address, allowed bool) error {
	o := &SwapList{}
	err := d.orm.Where(&SwapList{TokenAddress: token.String()}).First(o).Error
	if (err != nil && err == gorm.ErrRecordNotFound) || o.ID == 0 {
		err = d.orm.Create(&SwapList{
			TokenAddress:  token.String(),
			IsWhitelisted: &allowed,
		}).Error
	} else if err == nil {
		o.IsWhitelisted = &allowed
		err = d.orm.Save(o).Error
	}
	return err
}

func (d *Db) Drop() error {
	// IndexerStatus Prediction GameState
	err := d.orm.Exec("drop table indexer_statuses").Error
	if err == nil {
		err = d.orm.Exec("drop table predictions").Error
	}
	if err == nil {
		err = d.orm.Exec("drop table game_states").Error
	}
	if err == nil {
		err = d.orm.Exec("drop table swap_lists").Error
	}
	return err
}
