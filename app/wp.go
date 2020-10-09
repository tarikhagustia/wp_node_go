package app

import (
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	"github.com/gin-gonic/gin"
	qrcode "github.com/skip2/go-qrcode"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func (app *Application) WpSendMessage(c *gin.Context) {
	msisdn := c.Request.FormValue("msisdn")
	message := c.Request.FormValue("message")

	text := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: msisdn + "@s.whatsapp.net",
		},
		Text: message,
	}

	response, err := app.Conn.Send(text)
	if err != nil {
		c.JSON(500, gin.H{
			"code":    http.StatusInternalServerError,
			"message": err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"code":    http.StatusOK,
		"data": gin.H{
			"id": response,
		},
	})
	c.Abort()
	return
}

func (app *Application) WpLogin(c *gin.Context) {
	output := c.Request.FormValue("output")
	if app.IsConnected == false {
		timeoutString := c.Request.FormValue("timeout")
		timeout, _ := strconv.Atoi(timeoutString)
		conn, err := whatsapp.NewConn(time.Duration(timeout) * time.Second)
		if err != nil {
			log.Println("Error")
		}
		app.Conn = conn
		app.Conn.SetClientVersion(2, 2041, 6)
		app.Conn.AddHandler(&waHandler{app.Conn, uint64(time.Now().Unix())})
	}

	qr, err := login(app.Conn)
	if err != nil {
		log.Println(err.Error())
		c.JSON(200, gin.H{
			"message": err.Error(),
		})
		c.Abort()
		return

	}
	png, _ := qrcode.Encode(qr, qrcode.Medium, 256)
	qr = base64.StdEncoding.EncodeToString(png)
	qrString := "data:image/png;base64," + qr
	if output == "html" {
		response := `
        <html>
          <head>
            <title>WhatsApp Login</title>
          </head>
          <body>
            <img src="` + qrString + `" />
            <p>
              <b>QR Code Scan</b>
              <br/>
            </p>
          </body>
        </html>
      `
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(response))
	} else {
		c.JSON(200, gin.H{
			"message": "success",
			"data": gin.H{
				"qr": qrString,
			},
		})
	}

}

func login(wac *whatsapp.Conn) (string, error) {
	session, err := readSession()
	if err == nil {
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return "", fmt.Errorf("restoring session failed: %v", err)
		}
	} else {
		qr := make(chan string)
		go func() {
			session, err := wac.Login(qr)
			if err = writeSession(session); err != nil {
				fmt.Errorf("error saving session: %v", err)
			}
		}()
		message := <-qr
		return message, nil
	}
	return "", nil
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open("./storage/whatsappSession.gob")
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	file, err := os.Create("./storage/whatsappSession.gob")
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}

type waHandler struct {
	c         *whatsapp.Conn
	startTime uint64
}

//HandleError needs to be implemented to be a valid WhatsApp handler
func (h *waHandler) HandleError(err error) {

	log.Println(reflect.TypeOf(err).String())
	if strings.Contains(err.Error(), "server closed connection") {
		log.Printf("Connection failed, underlying error: %v", err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		err := h.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	}
	if e, ok := err.(*whatsapp.ErrConnectionFailed); ok {
		log.Printf("Connection failed, underlying error: %v", e.Err)
		log.Println("Waiting 30sec...")
		<-time.After(30 * time.Second)
		log.Println("Reconnecting...")
		err := h.c.Restore()
		if err != nil {
			log.Fatalf("Restore failed: %v", err)
		}
	} else {
		log.Printf("error occoured: %v\n", err)
	}

	log.Println(err)
}

//Optional to be implemented. Implement HandleXXXMessage for the types you need.
func (wh *waHandler) HandleTextMessage(message whatsapp.TextMessage) {
	if message.Info.FromMe || message.Info.Timestamp < wh.startTime {
		return
	}
	fmt.Printf("%v %v %v %v\n\t%v\n", message.Info.Timestamp, message.Info.Id, message.Info.RemoteJid, message.ContextInfo.QuotedMessageID, message.Text)
}

//Optional
func (*waHandler) HandleJsonMessage(message string) {
	log.Println(message)
}

/*//Example for media handling. Video, Audio, Document are also possible in the same way
func (h *waHandler) HandleImageMessage(message whatsapp.ImageMessage) {
	data, err := message.Download()
	if err != nil {
		if err != whatsapp.ErrMediaDownloadFailedWith410 && err != whatsapp.ErrMediaDownloadFailedWith404 {
			return
		}
		if _, err = h.c.LoadMediaInfo(message.Info.RemoteJid, message.Info.Id, strconv.FormatBool(message.Info.FromMe)); err == nil {
			data, err = message.Download()
			if err != nil {
				return
			}
		}
	}
	filename := fmt.Sprintf("%v/%v.%v", os.TempDir(), message.Info.Id, strings.Split(message.Type, "/")[1])
	file, err := os.Create(filename)
	defer file.Close()
	if err != nil {
		return
	}
	_, err = file.Write(data)
	if err != nil {
		return
	}
	log.Printf("%v %v\n\timage received, saved at:%v\n", message.Info.Timestamp, message.Info.RemoteJid, filename)
}*/
