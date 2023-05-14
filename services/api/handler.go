package api

import (
	"encoding/json"
	"fmt"
	"github.com/GameDev-Tube/famemma.PRE2E/services/db"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/labstack/echo/v4"
	"math/big"
	"net/http"
	"strconv"
	"strings"
)

type LeaderboardItem struct {
	NftId uint64 `json:"nft_id"`
	Score uint64 `json:"score"`
}

type LeaderboardResponse struct {
	Items []LeaderboardItem `json:"items"`
}

type LeaderboardPos struct {
	Position uint64 `json:"position"`
	ChoiceA  uint64 `json:"choice_a"`
	ChoiceB  uint64 `json:"choice_b"`
	ChoiceC  uint64 `json:"choice_c"`
	ChoiceD  uint64 `json:"choice_d"`
	ChoiceE  uint64 `json:"choice_e"`
	ChoiceF  uint64 `json:"choice_f"`
	NftId    uint64 `json:"nft_id"`
	Score    uint   `json:"score"`
	Owner    string `json:"owner"`
}

// /leaderboards/:limit/:offset

func (h *handler) GetLeaderboard(c echo.Context) error {
	limit, err := strconv.ParseInt(c.Param("limit"), 10, 64)
	if err != nil || int(limit) <= 0 {
		return c.JSON(http.StatusBadRequest, "bad request")
	}
	offset, err := strconv.ParseInt(c.Param("offset"), 10, 64)
	if err != nil || int(limit) <= 0 {
		return c.JSON(http.StatusBadRequest, "bad request")
	}

	lb, err := h.db.Leaderboards(int(offset), int(limit))
	if err != nil {
		fmt.Println("GetLeaderboards error: ", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}
	out := make([]LeaderboardPos, len(lb))
	for i := range lb {
		ownr, err := h.binding.OwnerOf(&bind.CallOpts{}, big.NewInt(0).SetUint64(lb[i].NftId))
		if err != nil && strings.Contains(err.Error(), "invalid token ID") {
			// token was burned, contract ensures this will happen when allowed.
			ownr = common.HexToAddress("0x0000000000000000000000000000000000000000")
			err = nil
		}
		if err != nil {
			fmt.Println("GetLeaderboards error: ", err)
			return c.JSON(http.StatusInternalServerError, "internal server error")
		}
		out[i] = LeaderboardPos{
			Position: uint64(offset) + uint64(i) + 1,
			ChoiceA:  lb[i].ChoiceA,
			ChoiceB:  lb[i].ChoiceB,
			ChoiceC:  lb[i].ChoiceC,
			ChoiceD:  lb[i].ChoiceD,
			ChoiceE:  lb[i].ChoiceE,
			ChoiceF:  lb[i].ChoiceF,
			NftId:    lb[i].NftId,
			Score:    lb[i].Score,
			Owner:    ownr.String(),
		}
	}
	return c.JSON(http.StatusOK, out)
}

type PredictionRequest struct {
	ChoiceA     uint64 `json:"choice_a"`
	ChoiceB     uint64 `json:"choice_b"`
	ChoiceC     uint64 `json:"choice_c"`
	ChoiceD     uint64 `json:"choice_d"`
	ChoiceE     uint64 `json:"choice_e"`
	ChoiceF     uint64 `json:"choice_f"`
	GameEdition uint   `json:"game_edition"`
}

type PredictionResponse struct {
	Hash string `json:"hash"`
}

func (h *handler) GetPredictionPayload(c echo.Context) error {
	req := &PredictionRequest{}
	err := c.Bind(req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "bad request")
	}
	state, err := h.db.GetGameState()
	if err != nil {
		fmt.Println("GetPredictionPayload error: ", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}
	if state.GamePhase != db.GamePhasePredicting {
		return c.JSON(http.StatusConflict, "not in predicting phase")
	}
	if state.GameEdition != req.GameEdition {
		return c.JSON(http.StatusConflict, fmt.Sprintf("Invalid game edition, expected %s", state.GameEdition))
	}

	hash, err := predictionHash(req)
	if err != nil {
		fmt.Println("GetPredictionPayload error: ", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}

	resp := &PredictionResponse{
		Hash: hexutil.Encode(hash[:]),
	}

	err = h.db.CreatePrediction(
		req.ChoiceA,
		req.ChoiceB,
		req.ChoiceC,
		req.ChoiceD,
		req.ChoiceE,
		req.ChoiceF,
		req.GameEdition,
		hexutil.Encode(hash[:]))

	if err != nil {
		fmt.Println("GetPredictionPayload error: ", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}

	return c.JSON(http.StatusOK, resp)
}

type TokenWhitelistResponse struct {
	Whitelist []string `json:"whitelist"`
}

func (h *handler) GetSwapWhitelist(c echo.Context) error {
	out, err := h.db.GetSwappableTokens()
	if err != nil {
		fmt.Println("GetSwapWhitelist error: ", err)
		return c.JSON(http.StatusInternalServerError, "internal server error")
	}
	resp := &TokenWhitelistResponse{
		Whitelist: make([]string, len(out)),
	}
	for i := range out {
		resp.Whitelist[i] = out[i].String()
	}
	return c.JSON(http.StatusOK, resp)
}

func predictionHash(val *PredictionRequest) (common.Hash, error) {
	data, err := json.Marshal(val)
	if err != nil {
		return common.Hash{}, err
	}
	return crypto.Keccak256Hash(data), nil
}
