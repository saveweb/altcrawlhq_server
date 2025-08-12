package admin

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/db"
)

//go:embed list_projects.html
var ListProjectsHTML []byte

func getAllProjects() ([]string, error) {
	projects, err := db.DbWriteSqlc.GetAllProjects(context.TODO())
	if err != nil {
		return nil, err
	}
	return projects, nil
}

func ListProjectsHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", ListProjectsHTML)
}

func ProjectsTableHandler(c *gin.Context) {
	projects, err := getAllProjects()
	if err != nil {
		c.Data(http.StatusInternalServerError, "text/html; charset=utf-8", []byte(`
			<div class="error">Error loading projects: `+err.Error()+`</div>
		`))
		return
	}

	if len(projects) == 0 {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
			<div class="no-projects">No projects found</div>
			<script>document.getElementById('project-count').textContent = '0';</script>
		`))
		return
	}

	html := `
		<table class="projects-table">
			<thead>
				<tr>
					<th>Project Name</th>
					<th>Actions</th>
				</tr>
			</thead>
			<tbody>
	`

	for _, project := range projects {
		html += fmt.Sprintf(`
			<tr>
				<td><strong>%s</strong></td>
				<td class="actions">
					<button class="btn btn-info" onclick="window.open('/api/projects/%s/urls', '_blank')">View URLs</button>
				</td>
			</tr>
		`, project, project)
	}

	html += `
			</tbody>
		</table>
		<script>document.getElementById('project-count').textContent = '` + fmt.Sprintf("%d", len(projects)) + `';</script>
	`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
