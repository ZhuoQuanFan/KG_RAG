package handlers

import (
	"RAG/backend/services"
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/tmc/langchaingo/schema"
)

type AddRequest struct {
	Documents []struct {
		Text string `json:"text"`
	} `json:"documents"`
}

func AddDocumentsHandler(store *services.VectorStoreService, driver neo4j.DriverWithContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("Starting to add documents...")

		var req AddRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Failed to parse request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		docs := make([]schema.Document, len(req.Documents))
		for i, doc := range req.Documents {
			docs[i] = schema.Document{
				PageContent: doc.Text,
				Metadata:    map[string]interface{}{},
			}
		}

		if err := store.AddDocuments(context.Background(), docs); err != nil {
			log.Printf("Failed to store documents: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store documents"})
			return
		}

		
		prompt := "Extract entities and relationships from the following text in the format: entity1|entity2|relationship\n" +
			"If no relationships are found, return an empty string\n" +
			"Text: " + req.Documents[0].Text
		response, err := services.Query(prompt)
		
		if err != nil {
			log.Printf("OpenAI call failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": response})
			return
		}

		if response == "" {
			log.Printf("No entities or relationships found in document: %s", req.Documents)
			c.JSON(http.StatusOK, gin.H{"message": "No entities or relationships found in document"})
			return
		}

		entitiesAndRelations := strings.Split(response, "\n")
		for _, line := range entitiesAndRelations {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Split(line, "|")
			if len(parts) != 3 {
				log.Printf("Invalid relationship format: %s", line)
				continue
			}
			entity1, entity2, relation := parts[0], parts[1], parts[2]
			if err := services.StoreToNeo4j(driver, entity1, entity2, relation); err != nil {
				log.Printf("Failed to store to Neo4j: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store to Neo4j"})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Documents added successfully"})
	}
}