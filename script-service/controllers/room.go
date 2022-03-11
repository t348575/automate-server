package controllers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/automate/automate-server/script-service/config"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"go.uber.org/fx"
)

type RoomController struct {
	fx.In

	Redis *redis.Client
}

type User struct{}

type Conn struct {
	hb      time.Time
	hbCount uint16
	script  int64
	user    int64
	ws      *websocket.Conn
	active  bool
}

type AuthOptions struct {
	Script int64 `json:"script"`
}

var users map[uint64]User
var timeout = time.Second * 5
var hbInterval = time.Second * 1
var cfg config.Config

func RegisterRoomController(app *fiber.App, c RoomController) {
	app.Get("/room", websocket.New(c.JoinRoom))
}

func (r *RoomController) JoinRoom(c *websocket.Conn) {
	conn := Conn{
		hb:      time.Now(),
		hbCount: 0,
		script:  0,
		user:    0,
		ws:      c,
		active:  true,
	}

	var (
		mt  int
		msg []byte
		err error
	)

	// start heartbeat
	go conn.heartbeat()

	for conn.active {
		if mt, msg, err = c.ReadMessage(); err != nil {
			break
		}

		if mt != websocket.TextMessage {
			break
		}

		temp := strings.SplitN(string(msg), ",", 2)
		if len(temp) != 2 {
			break
		}

		switch temp[0] {
		case "auth":
			authOptions := new(AuthOptions)
			if err := json.Unmarshal([]byte(temp[1]), authOptions); err != nil {
				break
			}
			conn.Authenticate(authOptions)

		case "msg":
			// TODO redis stuff
		}

	}
}

func (c *Conn) Close(code int, text string) {
	c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, text))
}

func (r *Conn) heartbeat() {
	if time.Now().Sub(r.hb) > timeout {
		r.active = false
		r.Close(websocket.CloseNoStatusReceived, "Ping timeout")
	}
	r.hbCount++

	if r.hbCount > 5 && r.user == 0 {
		r.active = false
		r.Close(websocket.ClosePolicyViolation, "Not authenticated")
	}
}

func (r *Conn) Authenticate(options *AuthOptions) {
	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	res := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(res)

	a.Reuse()
	req := a.Request()
	req.Header.SetMethod(fiber.MethodPost)

	req.SetRequestURI(cfg.InternalServices)
	req.Header.Set("Content-Type", "application/json")

	// req.SetBody(c.Body())
	// if err := a.Parse(); err != nil {
	// 	return utils.StandardInternalError(c, err)
	// }

	// code, body, errArr := a.SetResponse(res).Timeout(5 * time.Second).Bytes()
	// if errArr != nil || len(errArr) != 0 {
	// 	return utils.StandardInternalError(c, errArr[0])
	// }

	// c.Set("Content-Type", "application/json")
	// return c.Status(code).Send(body)
}
