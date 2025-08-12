package admin

import (
	_ "embed"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/api/ws"
)

//go:embed online_clients.html
var OnlineClientsHTML []byte

func OnlineClientsHandler(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", OnlineClientsHTML)
}

func ClientsTableHandler(c *gin.Context) {
	clients := ws.GetOnlineClients()

	if len(clients) == 0 {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
			<div class="no-clients">No clients online</div>
			<script>document.getElementById('client-count').textContent = '0';</script>
		`))
		return
	}

	html := `
		<table class="clients-table">
			<thead>
				<tr>
					<th>Status</th>
					<th>Identifier</th>
					<th>Project</th>
					<th>Last Ping</th>
					<th>Actions</th>
				</tr>
			</thead>
			<tbody>
	`

	for _, client := range clients {
		statusEmoji := getStatusEmoji(client.LastPing)
		formattedTime := client.LastPing.Format("2006-01-02 15:04:05")

		html += fmt.Sprintf(`
			<tr data-identifier="%s">
				<td><span class="status-indicator">%s</span></td>
				<td><strong>%s</strong></td>
				<td>%s</td>
				<td>%s</td>
				<td class="actions">
					<button class="btn" onclick="sendSignal('%s', 10, 'SIGUSR1')">USR1</button>
					<button class="btn" onclick="sendSignal('%s', 12, 'SIGUSR2')">USR2</button>
					<button class="btn btn-warning" onclick="sendSignal('%s', 15, 'SIGTERM')">TERM</button>
					<button class="btn btn-danger" onclick="sendSignal('%s', 9, 'SIGKILL')">KILL</button>
				</td>
			</tr>
		`, client.Identifier, statusEmoji, client.Identifier, client.Project, formattedTime,
			client.Identifier, client.Identifier, client.Identifier, client.Identifier)
	}

	html += `
			</tbody>
		</table>
		<script>document.getElementById('client-count').textContent = '` + fmt.Sprintf("%d", len(clients)) + `';</script>
	`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func getStatusEmoji(lastPing time.Time) string {
	now := time.Now()
	diffSeconds := now.Sub(lastPing).Seconds()

	if diffSeconds <= 5 {
		return "🟢"
	} else if diffSeconds <= 30 {
		return "🟡"
	} else {
		return "🔴"
	}
}
