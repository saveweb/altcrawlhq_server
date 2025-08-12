package projects

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/auth"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/db"
	"github.com/saveweb/altcrawlhq_server/internal/sqlc_model"
)

func ResetHandler(c *gin.Context) {
	if !auth.IsAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	ID := c.Param("id")
	project := c.Param("project")
	if ID == "" || project == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID and Project are required"})
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

	err = qtx.ResetURL(ctx, sqlc_model.ResetURLParams{
		Project: project,
		ID:      ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "message": "Failed to reset URL"})
		return
	}

	err = tx.Commit()

	if err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Reset"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
