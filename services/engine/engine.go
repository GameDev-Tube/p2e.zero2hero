package engine

import (
	"fmt"
	"github.com/GameDev-Tube/famemma.PRE2E/services/db"
	"time"
)

func (e *Engine) Run() {
	for {
		err := e.tick()
		if err != nil {
			fmt.Printf("Engine: failed to run tick: %s\n", err.Error())
		}
	}
}

func (e *Engine) tick() error {
	const batchSize = 50

	state, err := e.db.GetGameState()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}
	if state.GamePhase != db.GamePhaseActive {
		time.Sleep(time.Second * 15)
		return nil
	}

	return e.db.Transact(func(db *db.Db) error {
		for i := int(0); ; i += batchSize {
			predictions, err := db.GetPredictions(batchSize, i, state.GameEdition)
			if err != nil {
				return err
			}
			if len(predictions) == 0 {
				break
			}

			for i := range predictions {
				updatePredictionScore(state, &predictions[i])
				err = db.UpdatePrediction(&predictions[i])
				if err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func updatePredictionScore(state *db.GameState, prediction *db.Prediction) {
	score := uint(0)
	score += calcScore(prediction.ChoiceA, state.OutcomeA)
	score += calcScore(prediction.ChoiceB, state.OutcomeB)
	score += calcScore(prediction.ChoiceC, state.OutcomeC)
	score += calcScore(prediction.ChoiceD, state.OutcomeD)
	score += calcScore(prediction.ChoiceE, state.OutcomeE)
	score += calcScore(prediction.ChoiceF, state.OutcomeF)
	prediction.Score = score
}

const maxDiffForPoints = 1000
const spotOnBonus = 4000
const maxPointsInScaling = 2500

func calcScore(predicted, actual uint64) uint {
	// the score counting is *very* much proof of concept
	if predicted == actual {
		return spotOnBonus + maxPointsInScaling
	}
	var diff uint64
	if predicted > actual {
		diff = predicted - actual
	} else {
		diff = actual - predicted
	}
	if diff > maxDiffForPoints {
		return 0
	}
	return uint(float64(maxDiffForPoints-diff) / float64(maxDiffForPoints) * float64(maxPointsInScaling))
	//return uint(float64((maxDiffForPoints-diff)/maxDiffForPoints) * maxPointsInScaling)
}
