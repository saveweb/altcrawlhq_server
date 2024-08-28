package altcrawlhqserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"git.archive.org/wb/gocrawlhq"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func isAuthorized(c *gin.Context) bool {
	authKey := c.GetHeader("X-Auth-Key")
	authSecret := c.GetHeader("X-Auth-Secret")
	identifier := c.GetHeader("X-Identifier")

	if authKey == "" || authSecret == "" {
		return false
	}

	if identifier == "" {
		return false
	}

	if authKey == "saveweb_key" && authSecret == "saveweb_sec" {
		return true
	}

	return false
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

type FeedRequest struct {
	Size     int    `json:"size"`
	Strategy string `json:"strategy"`
}

var MONGODB_URI string = os.Getenv("MONGODB_URI")

var mongoClient *mongo.Client

func connect_to_mongodb() {
	fmt.Println("Connecting to MongoDB...")
	fmt.Println("MONGODB_URI: len=", len(MONGODB_URI))
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(MONGODB_URI).SetServerAPIOptions(serverAPI).SetAppName("altcrawlhq").SetCompressors([]string{"zstd", "zlib", "snappy"})
	fmt.Println("AppName: ", *opts.AppName)
	fmt.Println("Compressors: ", opts.Compressors)

	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		panic(err)
	}
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		panic(err)
	}
	mongoClient = client
	fmt.Println("Connected to MongoDB!")
}

func init() {
	connect_to_mongodb()
}

func ServeHTTP() {
	g := gin.New()
	// g.Use(gin.Recovery())
	// err := g.SetTrustedProxies(nil)
	// if err != nil {
	// 	panic(err)
	// }
	g.GET("/", func(c *gin.Context) {
		time.Sleep(1 * time.Second)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
		})
	})

	apiGroup := g.Group("/api")
	{
		projectGroup := apiGroup.Group("/project/:project/")
		{
			projectGroup.GET("/feed", feedHandler)
			projectGroup.POST("/finished", finishHandler)
			projectGroup.POST("/discovered", discoveredHandler)
		}
		apiGroup.GET("/ws", websocketHandler)
	}
	if err := g.Run(); err != nil {
		panic(err)
	}
}
