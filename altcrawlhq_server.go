package altcrawlhqserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func isAuthorized(c *gin.Context) bool {
	authKey := c.GetHeader("X-Auth-Key")
	authSecret := c.GetHeader("X-Auth-Secret")
	// identifier := c.GetHeader("X-Identifier")

	if authKey == "" || authSecret == "" {
		return false
	}

	if authKey == "saveweb_key" && authSecret == "saveweb_sec" {
		return true
	}

	return false
}

type FeedRequest struct {
	Size     int    `json:"size"`
	Strategy string `json:"strategy"`
}

var MONGODB_URI string = os.Getenv("MONGODB_URI")

var mongoDatabase *mongo.Database

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
	fmt.Println("Connected to MongoDB!")

	db := client.Database("crawlhq")
	mongoDatabase = db
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
		apiGroup.GET("/online", onlineClientsHandler)
	}
	if err := g.Run(); err != nil {
		panic(err)
	}
}
