package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

//go:embed public/*
var embeddedFiles embed.FS

var rdb *redis.Client
var ctx = context.Background()

const heartbeatTimeout = 30 * time.Second

type JoinRequest struct {
	PeerID string `json:"peer_id" binding:"required"`
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	router := gin.Default()
	router.Use(CORSMiddleware())

	templ := template.Must(template.New("").ParseFS(embeddedFiles, "public/*"))
	router.SetHTMLTemplate(templ)

	// 1. HOME ROUTE (Serves documentation dashboard)
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// 2. THE CONSOLIDATED DISCOVERY ROUTE (Register presence + Get Active Room State)
	router.POST("/basket/:id", func(c *gin.Context) {
		basketID := c.Param("id")
		var req JoinRequest

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body, peer_id is required"})
			return
		}

		redisKey := fmt.Sprintf("basket:%s", basketID)
		now := time.Now().UnixMilli()
		cutoff := now - heartbeatTimeout.Milliseconds()

		// Extract optional '?limit=' parameter early
		limitStr := c.DefaultQuery("limit", "100")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 100
		}

		// OPTIMIZATION: Use a Redis pipeline to batch commands into 1 network trip
		pipe := rdb.Pipeline()

		// 1. Add/update the current peer. Score is the raw numeric timestamp.
		pipe.ZAdd(ctx, redisKey, redis.Z{
			Score:  float64(now),
			Member: req.PeerID,
		})

		// 2. Clear out all dead peers instantly inside Redis.
		pipe.ZRemRangeByScore(ctx, redisKey, "-inf", strconv.FormatInt(cutoff, 10))

		// 3. Query only the requested limit of active peers (sorted newest first).
		// We use limit-1 because stop index is inclusive.
		peersCmd := pipe.ZRevRange(ctx, redisKey, 0, int64(limit-1))

		// 4. Query total active count in the room efficiently.
		cardCmd := pipe.ZCard(ctx, redisKey)

		// 5. Keep the basket key alive for an hour
		pipe.Expire(ctx, redisKey, 1*time.Hour)

		// Execute all 5 commands in a single atomic round-trip
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cluster state"})
			return
		}

		activePeers := peersCmd.Val()
		totalActive := int(cardCmd.Val())

		if activePeers == nil {
			activePeers = []string{}
		}

		c.JSON(http.StatusOK, gin.H{
			"basket_id":      basketID,
			"peers":          activePeers,
			"total_peers":    totalActive,
			"peers_returned": len(activePeers),
		})
	})

	log.Printf("App running on port %s...", port)
	router.Run("[::]:" + port)
}
