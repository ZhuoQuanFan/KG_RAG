package handlers

import (
	"context"
	"log"
	"net/http"

	"RAG/backend/services"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func ClearNeo4jHandler(driver neo4j.DriverWithContext) gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := services.ClearNeo4jDatabase(driver); err != nil {
            log.Printf("Failed to clear Neo4j database: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear Neo4j database"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "Neo4j database cleared successfully"})
    }
}

func ClearPGVectorHandler(store *services.VectorStoreService) gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := store.ClearPGVector(context.Background()); err != nil {
            log.Printf("Failed to clear PGVector table: %v", err)
            c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear PGVector table"})
            return
        }
        c.JSON(http.StatusOK, gin.H{"message": "PGVector table cleared successfully"})
    }
}