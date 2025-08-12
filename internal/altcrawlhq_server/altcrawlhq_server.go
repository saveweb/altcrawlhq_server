package altcrawlhqserver

import (
	_ "embed"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/api/projects"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/api/ws"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/db"
)

type FeedRequest struct {
	Size     int    `json:"size"`
	Strategy string `json:"strategy"`
}

func ServeHTTP() {
	db.Start()
	defer db.Shutdown()

	g := gin.New()
	g.GET("/", func(c *gin.Context) {
		time.Sleep(1 * time.Second)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
		})
	})

	apiG := g.Group("/api")
	{
		projectsG := apiG.Group("/projects/:project/")
		{
			projectsG.POST("/urls", projects.AddHandler)
			projectsG.GET("/urls", projects.GetHandler)
			projectsG.DELETE("/urls", projects.DeleteHandler)

			projectsG.POST("/seencheck", projects.SeencheckHandler)
			projectsG.POST("/reset/:ID", projects.ResetHandler)
		}
		apiG.GET("/ws", ws.WebsocketHandler)
	}
	adminG := g.Group("/admin")
	{

		adminApiG := adminG.Group("/api")
		{
			adminApiG.GET("/online", ws.OnlineClientsHandler)
			adminApiG.GET("/send-signal/:identifier", ws.SendSignalHandler)
		}
	}
	if err := g.Run(); err != nil {
		panic(err)
	}
}
