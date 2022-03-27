package controllers

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/automate/automate-server/models"
	"github.com/automate/automate-server/script-service/channel"
	"github.com/automate/automate-server/script-service/config"
	"github.com/automate/automate-server/utils-go"
	"github.com/go-playground/validator/v10"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"github.com/pkg/errors"
	"go.uber.org/fx"
)

var redisChannel *channel.RedisChannel

type RoomController struct {
	fx.In
}

type ScriptData struct{}

type Conn struct {
	hb          time.Time
	hbCount     uint16
	script      int64
	user        int64
	node        string
	ws          *websocket.Conn
	active      bool
	recvChannel <-chan *redis.Message
	sendClient  *redis.Client
}

type AuthOptions struct {
	Script  int64    `json:"script" validate:"required,number,min=1"`
	Token   string   `json:"token" validate:"required,ascii,min=1,max=1024"`
	Actions []string `json:"actions" validate:"required,len=0"`
}

type Permissions struct {
	Approved []string `json:"approved"`
	Denied   []string `json:"denied"`
	User     int64    `json:"user"`
}

var timeout = time.Second * 5
var hbInterval = time.Second * 1
var cfg config.Config

func RegisterRoomController(app *fiber.App, r *channel.RedisChannel, c RoomController) {
	redisChannel = r
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

	ticker := time.NewTicker(500 * time.Millisecond)
	go conn.heartbeat(ticker)

	for conn.active {
		if mt, msg, err = c.ReadMessage(); err != nil {
			break
		}

		if mt != websocket.TextMessage {
			break
		}

		before, after, found := strings.Cut(string(msg), ",")
		if !found {
			break
		}

		switch before {
		case "auth":
			authOptions := new(AuthOptions)
			if err := json.Unmarshal([]byte(after), authOptions); err != nil {
				break
			}
			if err := utils.ValidateStruct(validator.New().Struct(*authOptions)); err != nil {
				res, _ := json.Marshal(err)
				conn.Close(websocket.ClosePolicyViolation, string(res))
			}

			authOptions.Actions = []string{"CREATE", "READ", "UPDATE"}
			res, err := conn.Authenticate(*authOptions)
			if err != nil {
				conn.Close(websocket.ClosePolicyViolation, err.Error())
			}

			if len(res.Approved) == 0 {
				conn.Close(websocket.ClosePolicyViolation, "No permissions")
			}

			conn.script = authOptions.Script
			conn.user = res.User

			conn.node, err = conn.CreateRoom(models.NewScriptRoom{
				ScriptId: conn.script,
				User:     conn.user,
			})
			if err != nil {
				conn.Close(websocket.ClosePolicyViolation, err.Error())
			}

			conn.recvChannel, conn.sendClient = redisChannel.SubscribeToScript(conn.script, conn.user, conn.node)
		case "msg":
			// TODO redis stuff
		}

	}
}

func (c *Conn) Close(code int, text string) {
	go redisChannel.CloseChannel(c.script, c.node)
	c.ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(code, text))
}

func (r *Conn) heartbeat(ticker *time.Ticker) {
	for {
		_ = <-ticker.C
		if time.Now().Sub(r.hb) > timeout {
			r.active = false
			ticker.Stop()
			r.Close(websocket.CloseNoStatusReceived, "Ping timeout")
		}
		r.hbCount++

		if r.hbCount > 5 && r.user == 0 {
			r.active = false
			ticker.Stop()
			r.Close(websocket.ClosePolicyViolation, "Not authenticated")
		}
	}
}

func (r *Conn) Authenticate(options AuthOptions) (*Permissions, error) {
	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	res := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(res)

	a.Reuse()
	req := a.Request()
	req.Header.SetMethod(fiber.MethodPost)

	req.SetRequestURI(cfg.InternalService + "/scripts/stream")
	req.Header.Set("Content-Type", "application/json")

	data, err := json.Marshal(options)
	if err != nil {
		return nil, err
	}

	req.SetBody([]byte(data))
	if err := a.Parse(); err != nil {
		return nil, err
	}

	code, body, errArr := a.SetResponse(res).Timeout(5 * time.Second).Bytes()
	if errArr != nil || len(errArr) != 0 {
		return nil, err
	}

	if code != 200 {
		return nil, errors.New("Permission denied")
	}

	permissions := new(Permissions)
	if err := json.Unmarshal(body, permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}

func (r *Conn) CreateRoom(options models.NewScriptRoom) (string, error) {
	a := fiber.AcquireAgent()
	defer fiber.ReleaseAgent(a)

	res := fiber.AcquireResponse()
	defer fiber.ReleaseResponse(res)

	a.Reuse()
	req := a.Request()
	req.Header.SetMethod(fiber.MethodPost)

	req.SetRequestURI(cfg.InternalService + "/scripts/stream")
	req.Header.Set("Content-Type", "application/json")

	data, err := json.Marshal(options)
	if err != nil {
		return "", err
	}

	req.SetBody([]byte(data))
	if err := a.Parse(); err != nil {
		return "", err
	}

	code, body, errArr := a.SetResponse(res).Timeout(5 * time.Second).Bytes()
	if errArr != nil || len(errArr) != 0 {
		return "", errArr[0]
	}

	parsedData := make(map[string]string, 0)
	if err = json.Unmarshal(body, &parsedData); err != nil {
		return "", err
	}

	if code != 200 {
		if parsedData["error"] != "" {
			return "", errors.New(parsedData["error"])
		}

		return "", errors.New("Unknown error")
	}

	return parsedData["node"], nil
}
