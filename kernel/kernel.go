package kernel

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	app "github.com/tarikhagustia/wp_node_go/app"
	"log"
	"net/http"
	"strings"
)

const KEY = "8YIU3-00937-IEW8-33319"
const SECRET = "Hakuna Matata"

// AuthRequired : ..
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := VerifyToken(c.Request)
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("device", claims["device"])
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			c.Abort()
			return
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
			c.Abort()
			return
		}
		// log.Println(token)
	}
}

// ExtractToken : ..
func ExtractToken(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	//normally Authorization the_token_xxx
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
func VerifyToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractToken(r)
	if tokenString == "" {
		return nil, fmt.Errorf("Token not found")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(SECRET), nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}
func Initialize() error {

	application := app.Application{
		JwtSecret:   "Hakuna Matata",
		JwtToken:    "8YIU3-00937-IEW8-33319",
		IsConnected: false,
	}

	// Do restore session
	application.Restore()

	r := gin.Default()
	r.GET("/ping", application.Ping)

	// API Version 1 Router Group
	// Simple group: v1
	r.POST("/api/v1/whatsapp/auth", application.Login)
	r.Use(AuthRequired())
	{
		v1 := r.Group("/api/v1/whatsapp")
		{
			v1.POST("/login", application.WpLogin)
			v1.POST("/send/text", application.WpSendMessage)
		}
	}

	err := r.Run(":3000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}
