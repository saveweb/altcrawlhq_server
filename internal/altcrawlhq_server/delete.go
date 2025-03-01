package altcrawlhqserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/internetarchive/gocrawlhq"
	"github.com/saveweb/altcrawlhq_server/internal/sqlc_model"
)

func deleteHandler(c *gin.Context) {
	project := c.Param("project")
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	finishedPayload := gocrawlhq.DeletePayload{}
	err := c.BindJSON(&finishedPayload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Finished: %v\n", finishedPayload)

	// finishedPayload.URLs

	ctx := context.TODO()
	tx, err := dbWrite.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to start transaction"})
		return
	}
	defer tx.Rollback()
	qtx := dbWriteSqlc.WithTx(tx)
	for _, url := range finishedPayload.URLs {
		err := qtx.DoneURL(ctx, sqlc_model.DoneURLParams{
			Project: project,
			ID:      url.ID,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	err = tx.Commit()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	} else {
		statusAdd(project, finishedPayload.LocalCrawls)
		c.JSON(http.StatusNoContent, gin.H{
			"message":      "Deleted",
			"DeletedCount": len(finishedPayload.URLs),
		})
	}
}
