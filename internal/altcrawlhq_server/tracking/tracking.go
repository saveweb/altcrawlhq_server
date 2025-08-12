package tracking

import (
	"log/slog"
	"os"
	"path"
	"strconv"
	"time"
)

// in memory tracking status data
// flush to disk every 10 seconds
// read from disk on startup

var trackingStatus = map[string]int{}

func StatusAdd(key string, value int) {
	if _, ok := trackingStatus[key]; !ok {
		trackingStatus[key] = 0
	}
	trackingStatus[key] = trackingStatus[key] + value
}

func maintainTrackingStatus() {
	slog.Info("Starting tracking status maintenance")
	os.MkdirAll(path.Join("data", "status"), 0755)

	for {
		time.Sleep(10 * time.Second)

		for key, value := range trackingStatus {
			f, err := os.OpenFile(path.Join("data", "status", key), os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				slog.Error("Failed to open file", "error", err.Error())
				continue
			}
			_, err = f.WriteString(strconv.Itoa(value))
			if err != nil {
				slog.Error("Failed to write to file", "error", err.Error())
			}
			f.Close()
		}
	}
}

func init() {
	go maintainTrackingStatus()
}
