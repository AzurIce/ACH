package core

import (
	"log"
	"net/http"
	"strings"
	"time"

	"ach/core/middlewares"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/gorilla/websocket"
)

func (ach *ACHCore) initRouter() {
	ach.router = gin.Default()
	config := cors.DefaultConfig()
	config.ExposeHeaders = []string{"Authorization"}
	config.AllowCredentials = true
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	ach.router.Use(cors.New(config))
	ach.router.Use(middlewares.Frontend(ach.fs), middlewares.JWTAuth())
	// ach.router.Static("/", "./assets")
	ach.router.POST("/api/login", ach.loginHandler)
	ach.router.GET("/api/console", ach.handler)
}

func createToken(user string) (string, error) {

	t := jwt.New(jwt.GetSigningMethod("HS256"))

	t.Claims = &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
	}

	return t.SignedString([]byte("azurcraft"))
}

func (ach *ACHCore) loginHandler(c *gin.Context) {
	password := c.PostForm("password")
	// log.Println(c.Get("password"))
	// log.Println(password)
	if password == "daguoguonb" {
		token, err := createToken("foo")
		if err != nil {
			log.Println(err)
		}
		c.Writer.Write([]byte(token))
	} else {
		c.Status(400)
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ach *ACHCore) handler(c *gin.Context) {
	log.Println("consoleHandler")
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	ws.WriteMessage(
		websocket.TextMessage,
		[]byte(strings.Join(ach.outputBuffer[ach.outputCursor:], "")+strings.Join(ach.outputBuffer[:ach.outputCursor], "")),
	)
	ach.wsList = append(ach.wsList, ws)
	defer func() {
		ws.Close()
		ws = nil
	}()
	for {
		_, str, err := ws.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		ach.processInput(str)
	}
}

func sendMessage(ws *websocket.Conn, str string) {
	err := ws.WriteMessage(websocket.TextMessage, []byte(str))
	if err != nil {
		// Not established.
	}
}
