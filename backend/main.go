package main

import (
	"context"
	"log"

	"RAG/backend/config"
	"RAG/backend/handlers"
	"RAG/backend/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to initialize configurations: %v", err)
	}
	defer cfg.Neo4jDriver.Close(context.Background())

	vectorStoreService, err := services.NewVectorStoreService(context.Background(), cfg.Embedder, cfg.PGVectorConnURL)
	if err != nil {
		log.Fatalf("Failed to initialize vector store service: %v", err)
	}

	// 创建 Gin 路由器，启用 StrictSlash 选项
	r := gin.Default()
	r.Use(CORSMiddleware())

	// 确保路由处理严格路径（去除末尾斜杠）
	r.POST("/add", handlers.AddDocumentsHandler(vectorStoreService, cfg.Neo4jDriver))
	r.POST("/query", handlers.QueryHandler(vectorStoreService, cfg.Neo4jDriver))
	r.GET("/history", handlers.HistoryHandler(cfg.Neo4jDriver))
	r.POST("/clear-neo4j", handlers.ClearNeo4jHandler(cfg.Neo4jDriver))
	r.POST("/clear-pgvector", handlers.ClearPGVectorHandler(vectorStoreService))

	log.Println("Server starting on :9020")
	if err := r.Run(":9020"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 允许所有来源（仅用于开发环境）
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		// 日志记录请求信息以便调试
		log.Printf("Handling request: %s %s from origin: %s", c.Request.Method, c.Request.URL.Path, c.Request.Header.Get("Origin"))

		// 处理预检请求（OPTIONS）
		if c.Request.Method == "OPTIONS" {
			log.Println("Responding to OPTIONS preflight request")
			c.AbortWithStatus(204)
			return
		}

		c.Next()

		// 确保响应头包含 CORS 头（包括重定向）
		if c.Writer.Status() == 308 || c.Writer.Status() == 301 {
			log.Println("Adding CORS headers to redirect response")
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}
	}
}