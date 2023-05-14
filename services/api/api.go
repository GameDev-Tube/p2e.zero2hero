package api

import (
	"context"
	"fmt"
	"github.com/GameDev-Tube/famemma.PRE2E/bindings"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"strings"
)

func (a *Api) Run() error {
	e := echo.New()
	e.Use(middleware.Recover())

	dir, err := a.frontend.ReadDir("fe")
	if err != nil {
		return err
	}
	for _, itm := range dir {
		if itm.IsDir() || itm.Name() == "config.js" {
			continue
		}
		name := itm.Name()
		if strings.HasSuffix(itm.Name(), ".html") {
			name = strings.TrimSuffix(itm.Name(), ".html")
			if name == "index" {
				name = ""
			}
		}
		path := "fe/" + itm.Name()
		fmt.Printf("registering route: route=%s, file=%s\n", name, path)
		e.FileFS(name, path, a.frontend)
	}

	chainId, err := a.h.rpc.ChainID(context.Background())
	if err != nil {
		return err
	}
	chainIdHex := "0x" + strings.ToUpper(fmt.Sprintf("%x", chainId))

	ws := "ws"
	httpStr := "http"
	if a.externalIsHttps {
		ws += "s"
		httpStr += "s"
	}
	e.GET("config.js", func(c echo.Context) error {
		state, err := a.db.GetGameState()
		if err != nil {
			fmt.Println(err.Error())
			return c.JSON(http.StatusInternalServerError, "internal server error")
		}
		config := fmt.Sprintf(`const wsUrl = "%s://%s/api/ws"; const baseUrl = "%s://%s/"; `,
			ws, a.externalUrl, httpStr, strings.TrimSuffix(a.externalUrl, "/"))
		config = fmt.Sprintf(`%s const P2EAddr = "%s" ; const BEP20Addr = "%s" ; `,
			config, a.p2eAddress, a.bepAddress)
		config = fmt.Sprintf("%s const phase = %d; const edition = %d; const chainId = '%s';",
			config, state.GamePhase, state.GameEdition, chainIdHex)
		config = fmt.Sprintf("%s const p2eAbi = '%s'; const IBEP20Abi = '%s' ; ",
			config, bindings.P2EContractMetaData.ABI, bindings.IBEP20MetaData.ABI)

		return c.Blob(http.StatusOK, "application/javascript", []byte(config))
	})

	e.GET("api/ws", a.h.AdminWS)
	e.GET("api/swap-tokens", a.h.GetSwapWhitelist)
	e.POST("api/predict", a.h.GetPredictionPayload)
	e.GET("api/leaderboards/:limit/:offset", a.h.GetLeaderboard)

	return e.Start(a.sockBind)
}
