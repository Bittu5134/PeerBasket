package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"embed"
	"github.com/gin-gonic/gin"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

//go:embed static/*
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
	_ = router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	templ := template.Must(template.New("").ParseFS(embeddedFiles, "static/*.html"))
	router.SetHTMLTemplate(templ)

	staticFiles, err := fs.Sub(embeddedFiles, "static")
	if err != nil {
		log.Fatal(err)
	}
	router.StaticFS("/static", http.FS(staticFiles))

	// HOME ROUTE
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	// API ROUTE
	router.POST("/basket/:id", func(c *gin.Context) {
		basketID := c.Param("id")
		var req JoinRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body. 'peer_id' is required and must be a valid string."})
			c.Abort()
			return
		}

		redisKey := fmt.Sprintf("basket:%s", basketID)
		now := time.Now().UnixMilli()
		cutoff := now - heartbeatTimeout.Milliseconds()

		// Extract 'limit' parameter (optional)
		limitStr := c.DefaultQuery("limit", "100")
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			limit = 100
		}

		// Using pipeline to improve performance
		pipe := rdb.Pipeline()

		//Add/update the current peer. Score is the raw numeric timestamp.
		pipe.ZAdd(ctx, redisKey, redis.Z{
			Score:  float64(now),
			Member: req.PeerID,
		})

		// Clear out all dead peers instantly inside Redis.
		pipe.ZRemRangeByScore(ctx, redisKey, "-inf", strconv.FormatInt(cutoff, 10))

		// returns active peers (sorted newest first).
		peersCmd := pipe.ZRevRange(ctx, redisKey, 0, int64(limit-1))

		// total active count
		cardCmd := pipe.ZCard(ctx, redisKey)

		// Keep the basket key alive for an hour
		pipe.Expire(ctx, redisKey, 1*time.Hour)

		// bulk execute all commands
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
			"total_peers":    totalActive,
			"peers_returned": len(activePeers),
			"peers":          activePeers,
		})
	})

	// 404 Redirect to Home
	router.NoRoute(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Redirect(http.StatusTemporaryRedirect, "/")
			c.Abort()
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid API Route"})

	})

	log.Printf("App running on port %s...", port)
	router.Run("[::]:" + port)
}
