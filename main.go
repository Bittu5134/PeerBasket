package main

import (
        "context"
        "fmt"
        "log"
        "os"
        "os/signal"
        "strconv"
        "strings"
        "syscall"
        "time"

        "embed"
        "github.com/gin-contrib/cors"
        "github.com/gin-contrib/gzip"
        "github.com/gin-gonic/gin"
        "net/http"

        "github.com/joho/godotenv"
        "github.com/redis/go-redis/v9"
        "log/slog"
)

//go:embed public/*
var embeddedFiles embed.FS

const heartbeatTimeout = 30 * time.Second

type JoinRequest struct {
        PeerID string `json:"peer_id" binding:"required"`
}

type Server struct {
        rdb *redis.Client
}

func NewServer(rdb *redis.Client) *Server {
        return &Server{rdb: rdb}
}

func serveRawFile(c *gin.Context, filepath string, contentType string) {    
        data, err := embeddedFiles.ReadFile("public/" + filepath)
        if err != nil {
                c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
                return
        }
        c.Header("Cache-Control", "public, max-age=1800, s-maxage=3600")
        c.Data(http.StatusOK, contentType, data)
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

        rdb := redis.NewClient(&redis.Options{
                Addr: redisAddr,
                DB:   0,
        })

        var err error
        for i := 1; i <= 5; i++ {
                ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
                _, err = rdb.Ping(ctx).Result()
                cancel()
                if err == nil {
                        break
                }
                slog.Warn("Redis not ready, retrying...", "attempt", i, "error", err)
                time.Sleep(2 * time.Second)
        }
        if err != nil {
                log.Fatalf("Redis connection failed: %v", err)
        }

        corsConfig := cors.Config{
                AllowAllOrigins: true,
                AllowMethods:    []string{http.MethodGet, http.MethodPost, http.MethodOptions, http.MethodPut, http.MethodDelete},
                AllowHeaders:    []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
                MaxAge:          12 * time.Hour,
        }

        router := gin.Default()
        router.Use(cors.New(corsConfig))
        router.Use(gzip.Gzip(gzip.DefaultCompression))
        _ = router.SetTrustedProxies([]string{"127.0.0.1", "::1"})

        server := NewServer(rdb)

        // STATIC FILE ROUTES
        router.GET("/", func(c *gin.Context) { serveRawFile(c, "index.html", "text/html; charset=utf-8") })
        router.GET("/index.html", func(c *gin.Context) { serveRawFile(c, "index.html", "text/html; charset=utf-8") })
        router.GET("/logo.svg", func(c *gin.Context) { serveRawFile(c, "logo.svg", "image/svg+xml") })
        router.GET("/logo.webp", func(c *gin.Context) { serveRawFile(c, "logo.webp", "image/webp") })
        router.GET("/banner.webp", func(c *gin.Context) { serveRawFile(c, "banner.webp", "image/webp") })

        // API ROUTES
        router.POST("/basket/:id", server.handleBasket)
        router.GET("/ping", server.handlePing)

        // FALLBACK
        router.NoRoute(server.handleNoRoute)

        srv := &http.Server{Addr: ":" + port, Handler: router}

        go func() {
                slog.Info("starting server", "port", port)
                if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
                        slog.Error("server failed", "error", err)
                        os.Exit(1)
                }
        }()

        // GRACEFUL SHUTDOWN
        quit := make(chan os.Signal, 1)
        signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
        <-quit
        slog.Info("shutting down server")

        shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer shutCancel()
        if err := srv.Shutdown(shutCtx); err != nil {
                slog.Error("server forced to shutdown", "error", err)       
        }
}

// SERVER ROUTES

func (s *Server) handleBasket(c *gin.Context) {
        basketID := c.Param("id")
        var req JoinRequest
        if err := c.ShouldBindJSON(&req); err != nil {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body. 'peer_id' is required and must be a valid string."})
                return
        }

        if !isValidID(basketID) {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'basket_id'. Must be 1-64 characters long and contain only alphanumeric characters, hyphens, underscores, dots, or colons."})
                return
        }

        if !isValidID(req.PeerID) {
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid 'peer_id'. Must be 1-64 characters long and contain only alphanumeric characters, hyphens, underscores, dots, or colons."})
                return
        }

        redisKey := fmt.Sprintf("basket:%s", basketID)
        now := time.Now().UnixMilli()
        cutoff := now - heartbeatTimeout.Milliseconds()

        limitStr := c.DefaultQuery("limit", "100")
        limit, _ := strconv.Atoi(limitStr)
        if limit > 1000 {
                limit = 1000
        }
        if limit <= 0 {
                limit = 100
        }

        ctx := c.Request.Context()
        pipe := s.rdb.Pipeline()

        pipe.ZAdd(ctx, redisKey, redis.Z{
                Score:  float64(now),
                Member: req.PeerID,
        })
        pipe.ZRemRangeByScore(ctx, redisKey, "-inf", strconv.FormatInt(cutoff, 10))
        peersCmd := pipe.ZRevRange(ctx, redisKey, 0, int64(limit-1))        
        cardCmd := pipe.ZCard(ctx, redisKey)
        pipe.Expire(ctx, redisKey, time.Hour)

        _, err := pipe.Exec(ctx)
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
}

func (s *Server) handleNoRoute(c *gin.Context) {
        if c.Request.Method == http.MethodGet && !strings.HasPrefix(c.Request.URL.Path, "/basket") {
                serveRawFile(c, "404.html", "text/html; charset=utf-8")     
        } else if strings.HasPrefix(c.Request.URL.Path, "/basket") {        
                c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body. 'peer_id' is required and must be a valid string. Also '/basket/:id' must be a POST request"})
        } else {
                c.JSON(http.StatusNotFound, gin.H{"error": "Route Not Found"})
        }
}

func (s *Server) handlePing(c *gin.Context) {
        ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
        defer cancel()

        if err := s.rdb.Ping(ctx).Err(); err != nil {
                c.JSON(http.StatusInternalServerError, gin.H{
                        "status":  "Error",
                        "message": "Redis connection offline",
                        "error":   err.Error(),
                })
                return
        }

        c.JSON(http.StatusOK, gin.H{"status": "A-OK!", "message": "Mostly harmless..."})
}

func isValidID(id string) bool {
        if len(id) == 0 || len(id) > 64 {
                return false
        }
        for i := 0; i < len(id); i++ {
                ch := id[i]
                if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' || ch == '.' || ch == ':') {
                        return false
                }
        }
        return true
}