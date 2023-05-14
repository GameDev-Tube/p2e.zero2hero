package api

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"golang.org/x/net/websocket"
	"io"
	"strconv"
)

const (
	wsMessageErr        = "error"
	wsMessageGameUpdate = "update-game-outcomes"
	wsMessageOk         = "ok"
)

type WSMessage struct {
	Event string            `json:"event"`
	Data  map[string]string `json:"data"`
}

type WSMessageAuth struct {
	Key string `json:"key"`
}

func (h *handler) AdminWS(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer func() {
			_ = ws.Close()
			fmt.Printf("websocket connection closed \n")
		}()
		for {
			auth := &WSMessageAuth{}
			err := websocket.JSON.Receive(ws, auth)
			if err != nil {
				if err == io.EOF { // no data
					continue
				}
				fmt.Printf("admin auth failure: %v\n", err)
				return
			}
			if auth.Key != h.adminkey {
				fmt.Println("admin auth failure: key missmatch")
				err = websocket.JSON.Send(ws, WSMessage{
					Event: wsMessageErr,
					Data:  map[string]string{"msg": "bad key"},
				})
				if err != nil {
					return
				}
			} else {
				stat, err := h.db.GetGameState()
				if err != nil {
					return
				}

				err = websocket.JSON.Send(ws, WSMessage{
					Event: wsMessageOk,
					Data: map[string]string{
						"a": fmt.Sprintf("%d", stat.OutcomeA),
						"b": fmt.Sprintf("%d", stat.OutcomeB),
						"c": fmt.Sprintf("%d", stat.OutcomeC),
						"d": fmt.Sprintf("%d", stat.OutcomeD),
						"e": fmt.Sprintf("%d", stat.OutcomeE),
						"f": fmt.Sprintf("%d", stat.OutcomeF),
					},
				})
				if err != nil {
					return
				}
				break
			}
		}

		buff := &WSMessage{}
		for {
			err := websocket.JSON.Receive(ws, buff)
			if err != nil {
				if err == io.EOF { // no data
					continue
				}
				fmt.Printf("fail reading message: %v\n", err)
				return
			}
			fmt.Printf("got message %+v", buff)
			switch buff.Event {
			case wsMessageGameUpdate:
				msgOk, err := h.handleUpdate(buff.Data)
				if !msgOk {
					// if msgOk is false, error is always nil
					err = websocket.JSON.Send(ws, WSMessage{
						Event: wsMessageErr,
						Data:  map[string]string{"msg": "Failed to parse data"},
					})
					if err != nil {
						fmt.Println(err.Error())
						return
					}
				}
				if err != nil {
					err = websocket.JSON.Send(ws, WSMessage{
						Event: wsMessageErr,
						Data: map[string]string{
							"msg": "failed to process data",
						},
					})
				}
				err = websocket.JSON.Send(ws, WSMessage{
					Event: wsMessageOk,
				})
			default:
				err = websocket.JSON.Send(ws, WSMessage{
					Event: wsMessageErr,
					Data:  map[string]string{"msg": "Unrecognized event"},
				})
			}
			/*
				// Read
				msg := ""
				err = websocket.Message.Receive(ws, &msg)
				if err != nil {
					c.Logger().Error(err)
				}
				fmt.Printf("%s\n", msg)*/
		}
	}).ServeHTTP(c.Response(), c.Request())
	return nil
}

func (h *handler) handleUpdate(data map[string]string) (bool, error) {
	aStr, oka := data["a"]
	bStr, okb := data["b"]
	cStr, okc := data["c"]
	dStr, okd := data["d"]
	eStr, oke := data["e"]
	fStr, okf := data["f"]
	if !(oka && okb && okc && okd && oke && okf) {
		return false, nil
	}
	a, erra := strconv.ParseUint(aStr, 10, 64)
	b, errb := strconv.ParseUint(bStr, 10, 64)
	c, errc := strconv.ParseUint(cStr, 10, 64)
	d, errd := strconv.ParseUint(dStr, 10, 64)
	e, erre := strconv.ParseUint(eStr, 10, 64)
	f, errf := strconv.ParseUint(fStr, 10, 64)

	if erra != nil || errb != nil || errc != nil || errd != nil || erre != nil || errf != nil {
		return false, nil
	}

	stat, err := h.db.GetGameState()
	if err != nil {
		return true, err
	}
	stat.OutcomeA = a
	stat.OutcomeB = b
	stat.OutcomeC = c
	stat.OutcomeD = d
	stat.OutcomeE = e
	stat.OutcomeF = f
	return true, h.db.UpdateGameState(stat)
}
