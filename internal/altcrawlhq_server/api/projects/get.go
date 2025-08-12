package projects

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/internetarchive/gocrawlhq"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/auth"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/db"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/model"
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
	Size int `json:"size" form:"size" binding:"required" validate:"min=0,max=100"`
}

func GetHandler(c *gin.Context) {
	project := c.Param("project")
	request := feedRequest{}

	if !auth.IsAuthorized(c) {
		slog.Error("Unauthorized")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.Bind(&request); err != nil {
		slog.Error("Bad Request", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.TODO()
	tx, err := db.DbWrite.Begin()
	if err != nil {
		slog.Error("Failed to start transaction", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()

	qtx := db.DbWriteSqlc.WithTx(tx)

	sqlcUrls, err := qtx.GetFreshURLs(ctx, sqlc_model.GetFreshURLsParams{
		Project: project,
		Limit:   int64(request.Size),
	})
	if err != nil {
		slog.Error("Failed to get fresh URLs", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to get fresh URLs"})
		return
	}

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

		URLs = append(URLs, *model.SqlcURL2hqURL(&record))
	}

	err = tx.Commit()
	if err == nil {
		if len(URLs) == 0 {
			c.JSON(http.StatusNoContent, gin.H{"message": "No URLs"})
			return
		} else {
			slog.Info("Feed", "project", project, "urlsCount", len(URLs))
			c.JSON(http.StatusOK, URLs)
			return
		}
	} else {
		slog.Error("Failed to commit transaction", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
