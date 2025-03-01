package altcrawlhqserver

import (
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"database/sql"

	"github.com/gin-gonic/gin"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/saveweb/altcrawlhq_server/internal/sqlc_model"
)

type FeedRequest struct {
	Size     int    `json:"size"`
	Strategy string `json:"strategy"`
}

var dbRead, dbWrite *sql.DB
var dbReadSqlc, dbWriteSqlc *sqlc_model.Queries

//go:embed schema.sql
var ddl string

func init() {
	os.MkdirAll("data", 0755)

	dbRead, _ = sql.Open("sqlite3", "file:data/hq.db")
	dbRead.SetMaxOpenConns(runtime.NumCPU())

	dbWrite, _ = sql.Open("sqlite3", "file:data/hq.db")
	dbWrite.SetMaxOpenConns(1)
	dbWrite.Exec("PRAGMA journal_mode=WAL")

	if _, err := dbWrite.Exec(ddl); err != nil {
		fmt.Println(err)
	}

	dbReadSqlc = sqlc_model.New(dbRead)
	dbWriteSqlc = sqlc_model.New(dbWrite)
}

func shutdown() {
	dbRead.Close()
	dbWrite.Close()
}

func ServeHTTP() {
	defer shutdown()
	g := gin.New()
	g.GET("/", func(c *gin.Context) {
		time.Sleep(1 * time.Second)
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
		})
	})

	apiGroup := g.Group("/api")
	{
		projectsGroup := apiGroup.Group("/projects/:project/")
		{
			projectsGroup.POST("/urls", addHandler)
			projectsGroup.GET("/urls", getHandler)
			projectsGroup.DELETE("/urls", deleteHandler)

			projectsGroup.POST("/seencheck", seencheckHandler)
			projectsGroup.POST("/reset/:ID", resetHandler)
		}
		apiGroup.GET("/ws", websocketHandler)
		apiGroup.GET("/online", onlineClientsHandler)
	}
	if err := g.Run(); err != nil {
		panic(err)
	}
}
