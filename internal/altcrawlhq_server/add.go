package altcrawlhqserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/internetarchive/gocrawlhq"
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

	// TODO: seencheck
	if addPayload.BypassSeencheck {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bypassSeencheck is not implemented yet"})
		return
	}

	fmt.Printf("Discovered: %v\n", addPayload)

	urls := make([]interface{}, len(addPayload.URLs))
	for i, url := range addPayload.URLs {
		urls[i] = url
	}

	ctx := context.TODO()
	tx, err := dbWrite.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()
	qtx := dbWriteSqlc.WithTx(tx)

	for _, url := range addPayload.URLs {
		_, err := qtx.CreateURL(ctx, ToCreateURLParams(URL2SqlcURL(&url, project)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if tx.Commit() == nil {
		c.JSON(http.StatusCreated, gin.H{
			"message":           "Added",
			"InsertedURLsCount": len(addPayload.URLs),
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
	}
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
