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

type FeedRequest struct {
	Size     int    `json:"size"`
	Strategy string `json:"strategy"`
}

var MONGODB_URI string = os.Getenv("MONGODB_URI")
var mongoDatabase *mongo.Database

const MONGODB_DB string = "crawlhq"

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

	db := client.Database(MONGODB_DB)
	mongoDatabase = db
}

func init() {
	connect_to_mongodb()
}

func ServeHTTP() {
	g := gin.New()
	g.GET("/", func(c *gin.Context) {
		time.Sleep(1 * time.Second)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
		})
	})

	apiGroup := g.Group("/api")
	{
		projectsGroup := apiGroup.Group("/projects/:project/")
		{
			projectsGroup.POST("/urls", addHandler)
			projectsGroup.GET("/urls", feedHandler)
			projectsGroup.DELETE("/urls", finishHandler)

			projectsGroup.POST("/seencheck", seencheckHandler)
			projectsGroup.POST("/reset", resetHandler)
		}
		apiGroup.GET("/ws", websocketHandler)
		apiGroup.GET("/online", onlineClientsHandler)
	}
	if err := g.Run(); err != nil {
		panic(err)
	}
}
