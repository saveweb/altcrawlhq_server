package altcrawlhqserver

import (
	"fmt"
	"net/http"

	"git.archive.org/wb/gocrawlhq"
	"github.com/gin-gonic/gin"
)

func finishHandler(c *gin.Context) {
	project := c.Param("project")
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	finishedPayload := gocrawlhq.FinishedPayload{}
	err := c.BindJSON(&finishedPayload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("Finished: %v\n", finishedPayload)
	finishedResp := gocrawlhq.FinishedResponse{
		Project: project,
	}
	c.JSON(http.StatusOK, finishedResp)
}
