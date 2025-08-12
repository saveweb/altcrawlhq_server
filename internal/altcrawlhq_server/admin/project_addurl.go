package admin

import (
	"context"
	_ "embed"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/db"
	"github.com/saveweb/altcrawlhq_server/internal/sqlc_model"
)

//go:embed project_addurl.html
var ProjectAddURLHTML []byte

// ProjectAddURLHandler serves the add URL page
func ProjectAddURLHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", ProjectAddURLHTML)
}

// AddURLHandler handles the form submission for creating new URLs
func AddURLHandler(c *gin.Context) {
	project := c.PostForm("project")
	id := c.PostForm("id")
	value := c.PostForm("value")
	via := c.PostForm("via")
	host := c.PostForm("host")
	path := c.PostForm("path")
	urlType := c.PostForm("type")
	crawler := c.PostForm("crawler")
	status := c.PostForm("status")

	// Validate required fields
	if project == "" || id == "" || value == "" || urlType == "" {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(`
			<div class="error">Missing required fields: project, id, value, and type are required.</div>
		`))
		return
	}

	// Validate type
	if urlType != "seed" && urlType != "asset" {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(`
			<div class="error">Invalid type. Must be 'seed' or 'asset'.</div>
		`))
		return
	}

	// Validate status
	if status == "" {
		status = "FRESH" // Default status
	}
	if status != "FRESH" && status != "CLAIMED" && status != "DONE" {
		c.Data(http.StatusBadRequest, "text/html; charset=utf-8", []byte(`
			<div class="error">Invalid status. Must be 'FRESH', 'CLAIMED', or 'DONE'.</div>
		`))
		return
	}

	// Create URL parameters
	params := sqlc_model.CreateURLParams{
		Project:   project,
		ID:        id,
		Value:     value,
		Via:       via,
		Host:      host,
		Path:      path,
		Type:      urlType,
		Crawler:   crawler,
		Status:    status,
		Timestamp: time.Now().Unix(),
	}

	// Create the URL in the database
	err := db.DbWriteSqlc.CreateURL(context.TODO(), params)
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(`
			<div class="error">Failed to create URL: `+err.Error()+`</div>
		`))
		return
	}

	// Return success message
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
		<div class="success">
			<h3>✅ URL Created Successfully!</h3>
			<p><strong>Project:</strong> `+project+`</p>
			<p><strong>ID:</strong> `+id+`</p>
			<p><strong>Value:</strong> `+value+`</p>
			<p><strong>Type:</strong> `+urlType+`</p>
			<p><strong>Status:</strong> `+status+`</p>
		</div>
		<script>
			setTimeout(function() {
				document.getElementById('url-form').reset();
				document.getElementById('result').innerHTML = '';
			}, 3000);
		</script>
	`))
}
