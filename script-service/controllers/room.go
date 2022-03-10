package controllers

import (
	"log"
	"time"

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
	c       *websocket.Conn
}

var users map[uint64]User
var timeout = time.Second * 5
var hbInterval = time.Second * 1

func RegisterRoomController(app *fiber.App, c RoomController) {
	app.Get("/room", websocket.New(c.JoinRoom))
}

func (r *RoomController) JoinRoom(c *websocket.Conn) {
	var (
		mt  int
		msg []byte
		err error
	)
	for {
		if mt, msg, err = c.ReadMessage(); err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s, %d", msg, mt)

		if err = c.WriteMessage(mt, msg); err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func (r *Conn) heartbeat() {
	if time.Now().Sub(r.hb) > timeout {
		// r.c.WriteMessage()
	}
	r.hbCount++
}
