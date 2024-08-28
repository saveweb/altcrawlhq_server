package altcrawlhqserver

import (
	"fmt"
	"net/http"

	"git.archive.org/wb/gocrawlhq"
	"github.com/gin-gonic/gin"
)

func discoveredHandler(c *gin.Context) {
	project := c.Param("project")
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	discoveredPayload := gocrawlhq.DiscoveredPayload{}
	err := c.BindJSON(&discoveredPayload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("Discovered: %v\n", discoveredPayload)

	// discoveredPayload.SeencheckOnly 只是 inspect 一下，优先判断

	discoveredResp := gocrawlhq.DiscoveredResponse{
		Project: project,
	}
	c.JSON(http.StatusCreated, discoveredResp)
}
