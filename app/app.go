package app

import (
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	"github.com/gin-gonic/gin"
	"time"
)

type Application struct {
	JwtToken    string
	JwtSecret   string
	Conn        *whatsapp.Conn
	IsConnected bool
}

func (app *Application) Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func (app *Application) Restore() (string, error) {
	conn, err := whatsapp.NewConn(60 * time.Second)
	app.Conn = conn
	app.Conn.SetClientVersion(2, 2041, 6)
	app.Conn.AddHandler(&waHandler{app.Conn, uint64(time.Now().Unix())})
	session, err := readSession()
	if err == nil {
		session, err = app.Conn.RestoreWithSession(session)
		if err != nil {
			return "", fmt.Errorf("restoring session failed: %v", err)
		}
	} else {
		return "nil", fmt.Errorf("Error while reading session")
	}
	app.IsConnected = true
	return "", nil
}
