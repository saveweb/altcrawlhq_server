package altcrawlhqserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/internetarchive/gocrawlhq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongodb record
type URLRecord struct {
	ID     primitive.ObjectID `json:"id" bson:"_id"`
	URL    string             `json:"url" bson:"url"`
	Hop    int                `json:"hop" bson:"hop"`
	Via    string             `json:"via" bson:"via"`
	Status string             `json:"status" bson:"status"`
}

type feedRequest struct {
	Size     int    `json:"size" form:"size" binding:"required" validate:"min=0,max=100"`
	Strategy string `json:"strategy" form:"strategy" binding:"required" validate:"oneof=lifo fifo"`
}

func feedHandler(c *gin.Context) {
	project := c.Param("project")
	request := feedRequest{}
	records := []URLRecord{}

	if !isAuthorized(c) {
		slog.Error("Unauthorized")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := c.Bind(&request); err != nil {
		slog.Error("Bad Request", "error", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	collection := mongoDatabase.Collection(project)
	record := URLRecord{}

	var opts *options.FindOneAndUpdateOptions
	switch request.Strategy {
	case "lifo":
		opts = options.FindOneAndUpdate().SetSort(bson.M{"_id": -1})
	case "fifo":
		opts = options.FindOneAndUpdate().SetSort(bson.M{"_id": 1})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Strategy"})
		return
	}

	// TODO: request.Size
	singleResult := collection.FindOneAndUpdate(context.TODO(), bson.M{"status": "TODO"}, bson.M{"$set": bson.M{"status": "PROCESSING"}}, opts)
	err := singleResult.Decode(&record)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.String(http.StatusNoContent, "")
			return
		}

		slog.Error("No URLs", "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	records = append(records, record)

	URLs := []gocrawlhq.URL{}
	// URLs = append(URLs, gocrawlhq.URL{
	// 	ID:    "", // uuid, 为空则客户端自动生成
	// 	Value: "https://example.com/",
	// 	Path:  "", // L 的数量表示 hop 深度
	// 	Via:   "", // 从哪儿 discovered 的链接
	// })
	for _, record := range records {
		URLs = append(URLs, gocrawlhq.URL{
			ID:    record.ID.Hex(),
			Value: record.URL,
			Path:  strings.Repeat("L", record.Hop),
			Via:   record.Via,
		})
	}

	feedResp := gocrawlhq.FeedResponse{
		Project: project,
		URLs:    URLs,
	}
	c.JSON(http.StatusOK, feedResp)
}
