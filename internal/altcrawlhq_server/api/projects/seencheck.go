package projects

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/internetarchive/gocrawlhq"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/db"
	"github.com/saveweb/altcrawlhq_server/internal/sqlc_model"
)

func SeencheckHandler(c *gin.Context) {
	project := c.Param("project")

	URLs := []gocrawlhq.URL{}
	err := c.BindJSON(&URLs)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "message": "Failed to bind JSON"})
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

	newURLs := []gocrawlhq.URL{}
	for _, url := range URLs {
		// Available: Value, Type
		count, err := qtx.CountSeen(ctx, sqlc_model.CountSeenParams{
			Project: project,
			Type:    url.Type,
			Value:   url.Value,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to get seen count", "url": url})
			return
		}
		if count == 0 {
			newURLs = append(newURLs, url)
			err := qtx.CreateSeen(ctx, sqlc_model.CreateSeenParams{
				Project: project,
				Type:    url.Type,
				Value:   url.Value,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to create seen record", "url": url})
				return
			}
		} else {
			err := qtx.RefreshSeen(ctx, sqlc_model.RefreshSeenParams{
				Project: project,
				Type:    url.Type,
				Value:   url.Value,
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to refresh seen record", "url": url})
				return
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to commit transaction"})
		return
	}

	if len(newURLs) == 0 {
		c.JSON(http.StatusNoContent, gin.H{"message": "No new URLs"})
		return
	} else {
		c.JSON(http.StatusOK, newURLs)
		return
	}

}
