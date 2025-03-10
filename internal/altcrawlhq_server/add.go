package altcrawlhqserver

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/internetarchive/gocrawlhq"
	"github.com/saveweb/altcrawlhq_server/internal/sqlc_model"
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

	ctx := context.TODO()
	tx, err := dbWrite.BeginTx(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()
	qtx := dbWriteSqlc.WithTx(tx)

	for _, url := range addPayload.URLs {
		if url.ID == "" {
			url.ID = uuid.New().String()
		}
		if url.Type == "" {
			url.Type = "seed"
		}
		if url.Status == "" {
			url.Status = "FRESH"
		}

		err := qtx.CreateURL(ctx, ToCreateURLParams(URL2SqlcURL(&url, project)))
		if err != nil {
			if err.Error() == "sqlite3: constraint failed: UNIQUE constraint failed: urls.project, urls.type, urls.value" {
				slog.Debug("URL already exists in LQ", "value", url.Value, "via", url.Via)
				continue
			}

			slog.Error("Failed to create URL", "error", err.Error(), "url", url)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			panic(err)
		}

		if !addPayload.BypassSeencheck {
			qtx.CreateSeen(ctx, sqlc_model.CreateSeenParams{
				Project: project,
				Type:    url.Type,
				Value:   url.Value,
			})
		}
	}

	err = tx.Commit()
	if err == nil {
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
