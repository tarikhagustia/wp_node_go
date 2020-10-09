package app

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"net/http"
)

const KEY = "8YIU3-00937-IEW8-33319"
const SECRET = "Hakuna Matata"

type Credential struct {
}

func (app *Application) Login(c *gin.Context) {
	user, password, _ := c.Request.BasicAuth()

	if password != KEY {
		ErrorResponse(c, 413, "Unauthenticated")
		return
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"device": user,
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(SECRET))
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Success",
		"data": gin.H{
			"token": tokenString,
		},
	})
}
