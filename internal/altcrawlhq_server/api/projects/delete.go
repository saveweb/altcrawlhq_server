package projects

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/internetarchive/gocrawlhq"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/auth"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/db"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/tracking"
	"github.com/saveweb/altcrawlhq_server/internal/sqlc_model"
)

func DeleteHandler(c *gin.Context) {
	project := c.Param("project")
	if !auth.IsAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	finishedPayload := gocrawlhq.DeletePayload{}
	err := c.BindJSON(&finishedPayload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.TODO()
	tx, err := db.DbWrite.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()
	qtx := db.DbWriteSqlc.WithTx(tx)
	for _, url := range finishedPayload.URLs {
		err := qtx.DeleteURL(ctx, sqlc_model.DeleteURLParams{
			Project: project,
			ID:      url.ID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	err = tx.Commit()

	if err == nil {
		tracking.StatusAdd(project, finishedPayload.LocalCrawls)
		c.JSON(http.StatusNoContent, gin.H{
			"message":      "Deleted",
			"DeletedCount": len(finishedPayload.URLs),
		})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
