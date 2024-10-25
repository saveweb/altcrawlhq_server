package altcrawlhqserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"git.archive.org/wb/gocrawlhq"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jellydator/ttlcache/v3"
)

var onlineClientsStats = ttlcache.New[string, gocrawlhq.IdentifyMessage](
	ttlcache.WithTTL[string, gocrawlhq.IdentifyMessage](time.Minute*1),
	ttlcache.WithDisableTouchOnHit[string, gocrawlhq.IdentifyMessage](),
)

func init() {
	fmt.Println("Initializing onlineClientsStats...")
	go onlineClientsStats.Start()
	fmt.Println("Initialized onlineClientsStats!")
}

func onlineClientsHandler(c *gin.Context) {
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	clients := make([]gocrawlhq.IdentifyMessage, 0)
	for _, client := range onlineClientsStats.Items() {
		clients = append(clients, client.Value())
	}
	c.JSON(http.StatusOK, clients)
}

func websocketHandler(c *gin.Context) {
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	upGrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}

	defer func() {
		closeSocketErr := ws.Close()
		if closeSocketErr != nil {
			panic(err)
		}
	}()

	for {
		wsMsgType, wsMsg, err := ws.ReadMessage()
		if err != nil {
			panic(err)
		}
		fmt.Printf("Message Type: %d, Message: %s\n", wsMsgType, string(wsMsg))

		if wsMsgType != websocket.TextMessage {
			panic("Message type is not text")
		}

		// {"type":"identify","payload":`+string(marshalled)+`}`
		msgType := struct {
			Type string `json:"type"`
		}{}
		if err := json.Unmarshal(wsMsg, &msgType); err != nil {
			panic(err)
		}

		if msgType.Type != "identify" {
			panic("Message type is not identify")
		}

		identifyMessage := struct {
			Payload gocrawlhq.IdentifyMessage `json:"payload"`
		}{}
		if err := json.Unmarshal(wsMsg, &identifyMessage); err != nil {
			panic(err)
		}

		onlineClientsStats.Set(identifyMessage.Payload.Identifier, identifyMessage.Payload, ttlcache.DefaultTTL)

		fmt.Printf("Identify Message: %+v\n", identifyMessage)

		err = ws.WriteJSON(struct {
			Reply string `json:"reply"`
		}{
			Reply: "Echo...",
		})
		if err != nil {
			panic(err)
		}
	}
}
