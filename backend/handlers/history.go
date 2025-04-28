package handlers

import (
	"RAG/backend/services"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func HistoryHandler(driver neo4j.DriverWithContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		history, err := services.GetQueryHistory(driver)
		if err != nil {
			log.Printf("Failed to retrieve history: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history"})
			return
		}

		c.JSON(http.StatusOK, history)
	}
}