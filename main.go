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

const heartbeatTimeout = 60 * time.Second

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

		// Enforce valid JSON payload containing the Peer ID
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body, peer_id is required"})
			return
		}

		redisKey := fmt.Sprintf("basket:%s", basketID)
		now := time.Now().UnixMilli()

		// Step A: Update current peer's heartbeat record in Redis
		err := rdb.HSet(ctx, redisKey, req.PeerID, now).Err()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update connection heartbeat"})
			return
		}
		rdb.Expire(ctx, redisKey, 1*time.Hour)

		// Step B: Extract optional '?limit=' parameter
		limitStr := c.DefaultQuery("limit", "100")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 100
		}

		// Step C: Fetch all records in the lobby
		allPeers, err := rdb.HGetAll(ctx, redisKey).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan basket presence"})
			return
		}

		var activePeers []string

		// Step D: Filter alive nodes vs dead records
		for peerID, timestampStr := range allPeers {
			timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
			if err != nil {
				continue
			}

			if now-timestamp < heartbeatTimeout.Milliseconds() {
				activePeers = append(activePeers, peerID)
			} else {
				// Wipe stale keys async to maintain light server footprint
				go rdb.HDel(ctx, redisKey, peerID)
			}
		}

		totalActive := len(activePeers)

		// Step E: Truncate to match user limit criteria
		if totalActive > limit {
			activePeers = activePeers[:limit]
		}

		if activePeers == nil {
			activePeers = []string{}
		}

		// Step F: Return dynamic connection payload back down the pipe
		c.JSON(http.StatusOK, gin.H{
			"basket_id":      basketID,
			"peers":          activePeers,
			"total_peers":    totalActive,
			"peers_returned": len(activePeers),
		})
	})

	log.Printf("App running on port %s...", port)
	router.Run(":" + port)
}