package altcrawlhqserver

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/internetarchive/gocrawlhq"
	"github.com/saveweb/altcrawlhq_server/internal/sqlc_model"
)

// mongodb record
type URLRecord struct {
	gocrawlhq.URL
	// ID primitive.ObjectID `json:"id" bson:"_id"`
	// URL    string             `json:"url" bson:"url"`
	Hop    int    `json:"hop" bson:"hop"`
	Via    string `json:"via" bson:"via"`
	Status string `json:"status" bson:"status"`
}

type feedRequest struct {
	Size     int    `json:"size" form:"size" binding:"required" validate:"min=0,max=100"`
	Strategy string `json:"strategy" form:"strategy" binding:"required" validate:"oneof=lifo fifo"`
}

func getHandler(c *gin.Context) {
	project := c.Param("project")
	request := feedRequest{}

	if !isAuthorized(c) {
		slog.Error("Unauthorized")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.Bind(&request); err != nil {
		slog.Error("Bad Request", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// var opts *options.FindOneAndUpdateOptions
	switch request.Strategy {
	case "lifo":
		// opts = options.FindOneAndUpdate().SetSort(bson.M{"_id": -1})
	case "fifo":
		// opts = options.FindOneAndUpdate().SetSort(bson.M{"_id": 1})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Strategy"})
		return
	}

	ctx := context.TODO()
	tx, err := dbWrite.Begin()
	if err != nil {
		slog.Error("Failed to start transaction", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	qtx := dbWriteSqlc.WithTx(tx)

	sqlcUrls, err := qtx.GetFreshURLs(ctx, sqlc_model.GetFreshURLsParams{
		Project: project,
		Limit:   int64(request.Size),
	})

	URLs := []gocrawlhq.URL{}
	for _, record := range sqlcUrls {
		err := qtx.ClaimThisURL(ctx, sqlc_model.ClaimThisURLParams{
			Project: project,
			ID:      record.ID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to claim URL", "url": record})
			return
		}

		URLs = append(URLs, *SqlcURL2hqURL(&record))
	}

	err = tx.Commit()
	if err == nil {
		if len(URLs) == 0 {
			c.JSON(http.StatusNoContent, gin.H{"message": "No URLs"})
			return
		} else {
			slog.Info("Feed", "project", project, "urls", URLs)
			c.JSON(http.StatusOK, URLs)
			return
		}
	} else {
		slog.Error("Failed to commit transaction", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
