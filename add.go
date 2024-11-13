package altcrawlhqserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/internetarchive/gocrawlhq"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func addHandler(c *gin.Context) {
	project := c.Param("project")
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	// identifier := c.GetHeader("X-Identifier")

	addPayload := gocrawlhq.AddPayload{}
	err := c.BindJSON(&addPayload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("Discovered: %v\n", addPayload)

	collection := mongoDatabase.Collection(project)

	opts := options.InsertMany().SetOrdered(false)
	urls := make([]interface{}, len(addPayload.URLs))
	for i, url := range addPayload.URLs {
		urls[i] = url
	}
	result, err := collection.InsertMany(context.TODO(), urls, opts)

	if addPayload.BypassSeencheck {

	}

	c.JSON(http.StatusCreated, gin.H{"message": "Added"})
}

// func inspectSeencheck(collection *mongo.Collection, URLs []gocrawlhq.URL, URLType string) ([]gocrawlhq.URL, error) {
// 	values := []string{}
// 	for _, URL := range URLs {
// 		values = append(values, URL.Value)
// 	}

// 	result, err := collection.Find(
// 		context.TODO(),
// 		bson.M{
// 			"type": URLType,
// 			"value": bson.M{
// 				"$in": values,
// 			},
// 		},
// 	)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer result.Close(context.Background())

// 	URLs = []gocrawlhq.URL{}
// 	for result.Next(context.Background()) {
// 		var URL gocrawlhq.URL
// 		err := result.Decode(&URL)
// 		if err != nil {
// 			panic(err)
// 		}
// 		URLs = append(URLs, URL)
// 	}

// }
