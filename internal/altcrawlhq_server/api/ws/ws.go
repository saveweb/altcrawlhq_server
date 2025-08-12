package ws

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/internetarchive/gocrawlhq"
	"github.com/jellydator/ttlcache/v3"
	"github.com/saveweb/altcrawlhq_server/internal/altcrawlhq_server/clientauth"
)

type IDMsgAndConn struct {
	Msg      gocrawlhq.IdentifyMessage
	Conn     *websocket.Conn
	LastPing time.Time
}

var onlineClientsStats = ttlcache.New(
	ttlcache.WithTTL[string, IDMsgAndConn](time.Minute*1),
	ttlcache.WithDisableTouchOnHit[string, IDMsgAndConn](),
)

func init() {
	go onlineClientsStats.Start()
	fmt.Println("Initialized onlineClientsStats!")
}

type ClientInfo struct {
	Project    string    `json:"project"`
	Identifier string    `json:"identifier"`
	LastPing   time.Time `json:"last_ping"`
}

func GetOnlineClients() []ClientInfo {
	clients := make([]ClientInfo, 0)
	for _, client := range onlineClientsStats.Items() {
		clientInfo := ClientInfo{
			Identifier: client.Value().Msg.Identifier,
			Project:    client.Value().Msg.Project,
			LastPing:   client.Value().LastPing,
		}
		clients = append(clients, clientInfo)
	}

	// sort by project::identifier
	slices.SortFunc(clients, func(a, b ClientInfo) int {
		if cmp := strings.Compare(a.Project, b.Project); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.Identifier, b.Identifier)
	})
	return clients
}

func OnlineClientsHandler(c *gin.Context) {
	// if !isAuthorized(c) {
	// 	c.JSON(http.StatusUnauthorized, gin.H{
	// 		"error": "Unauthorized",
	// 	})
	// 	return
	// }

	clients := GetOnlineClients()
	c.JSON(http.StatusOK, clients)
}

func WebsocketHandler(c *gin.Context) {
	if !clientauth.IsAuthorized(c) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}
	upGrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	ws, err := upGrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		panic(err)
	}

	defer func() {
		closeSocketErr := ws.Close()
		if closeSocketErr != nil {
			panic(err)
		}
	}()

	for {
		wsMsgType, wsMsg, err := ws.ReadMessage()
		if err != nil {
			// client disconnected
			slog.Error("WebSocket read error", "error", err.Error())
			return
		}
		slog.Info("WebSocket message received", "type", wsMsgType, "message", string(wsMsg))

		if wsMsgType != websocket.TextMessage {
			panic("Message type is not text")
		}

		// {"type":"identify","payload":`+string(marshalled)+`}`
		msgType := struct {
			Type string `json:"type"`
		}{}
		if err := json.Unmarshal(wsMsg, &msgType); err != nil {
			panic(err)
		}

		// if msgType.Type != "identify" {
		// 	panic("Message type is not identify")
		// }

		switch msgType.Type {
		case "identify":
			handleIdentifyMessage(ws, wsMsg)
		default:
			panic(fmt.Sprintf("Unknown message type: %s", msgType.Type))
		}

	}
}

func handleIdentifyMessage(ws *websocket.Conn, wsMsg []byte) error {
	identifyMessage := struct {
		Payload gocrawlhq.IdentifyMessage `json:"payload"`
	}{}
	if err := json.Unmarshal(wsMsg, &identifyMessage); err != nil {
		return fmt.Errorf("failed to unmarshal identify message: %w", err)
	}

	onlineClientsStats.Set(identifyMessage.Payload.Identifier, IDMsgAndConn{
		Msg:      identifyMessage.Payload,
		Conn:     ws,
		LastPing: time.Now(),
	}, ttlcache.DefaultTTL)

	fmt.Printf("Identify Message: %+v\n", identifyMessage)

	if err := ws.WriteJSON(struct {
		Type  string `json:"type"`
		Reply string `json:"reply"`
	}{
		Type:  "identify",
		Reply: "Echo...",
	}); err != nil {
		return fmt.Errorf("failed to write JSON reply: %w", err)
	}

	return nil
}

func findWebsocketConn(identifier string) (*websocket.Conn, error) {
	idAndConn := onlineClientsStats.Get(identifier)
	if idAndConn == nil {
		return nil, fmt.Errorf("no websocket connection found for identifier: %s", identifier)
	}
	return idAndConn.Value().Conn, nil
}

func SendSignalHandler(c *gin.Context) {
	identifier := c.Param("identifier")

	sig, err := strconv.ParseInt(c.Query("signal"), 10, 0)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid signal"})
		return
	}

	// syscall.SIGILL

	if err := sendSignalMessage(identifier, syscall.Signal(sig)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "signal sent"})
}

func sendSignalMessage(identifier string, sig os.Signal) error {
	ws, err := findWebsocketConn(identifier)
	if err != nil {
		return err
	}

	return ws.WriteJSON(struct {
		Type   string    `json:"type"`
		Signal os.Signal `json:"signal"`
	}{
		Type:   "signal",
		Signal: sig,
	})
}
