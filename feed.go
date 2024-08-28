package altcrawlhqserver

import (
	"net/http"

	"git.archive.org/wb/gocrawlhq"
	"github.com/gin-gonic/gin"
)

func feedHandler(c *gin.Context) {
	const emptyStatusCode = 204

	project := c.Param("project")
	if !isAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	URLs := []gocrawlhq.URL{}

	URLs = append(URLs, gocrawlhq.URL{
		ID:    "", // uuid, 为空则客户端自动生成
		Value: "https://blog.othing.xyz/",
		Path:  "", // L 的数量表示 hop 深度
		Via:   "", // 从哪儿 discovered 的链接
	})

	feedResp := gocrawlhq.FeedResponse{
		Project: project,
		URLs:    URLs,
	}
	if len(URLs) == 0 {
		c.JSON(emptyStatusCode, gin.H{"error": "No URLs"})
		return
	}
	c.JSON(http.StatusOK, feedResp)
}
