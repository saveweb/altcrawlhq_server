package admin

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed projects.html
var projectsHTML []byte

//go:embed index.html
var indexHTML []byte

func IndexHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
}

func ProjectsHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", projectsHTML)
}
