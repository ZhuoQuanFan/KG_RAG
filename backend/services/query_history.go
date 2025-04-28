package services

import (
	"context"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type QueryHistory struct {
	Content  string `json:"content"`
	Response string `json:"response"`
}

func SaveQueryHistory(driver neo4j.DriverWithContext, content, response string) error {
	session := driver.NewSession(context.Background(), neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(context.Background())

	_, err := session.Run(context.Background(),
		"CREATE (h:QueryHistory {content: $content, response: $response, timestamp: $timestamp})",
		map[string]interface{}{
			"content":   content,
			"response":  response,
			"timestamp": time.Now().Unix(),
		},
	)
	if err != nil {
		log.Printf("Failed to save query history: %v", err)
		return err
	}

	log.Println("Query history saved successfully")
	return nil
}

func GetQueryHistory(driver neo4j.DriverWithContext) ([]QueryHistory, error) {
	session := driver.NewSession(context.Background(), neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(context.Background())

	result, err := session.Run(context.Background(),
		"MATCH (h:QueryHistory) RETURN h.content, h.response ORDER BY h.timestamp DESC",
		nil,
	)
	if err != nil {
		log.Printf("Failed to query history: %v", err)
		return nil, err
	}

	var history []QueryHistory
	for result.Next(context.Background()) {
		record := result.Record()
		content, _ := record.Get("h.content")
		response, _ := record.Get("h.response")

		history = append(history, QueryHistory{
			Content:  content.(string),
			Response: response.(string),
		})
	}

	if err := result.Err(); err != nil {
		log.Printf("Failed to read history records: %v", err)
		return nil, err
	}

	return history, nil
}

func FindQueryHistory(driver neo4j.DriverWithContext, content string) (QueryHistory, error) {
	session := driver.NewSession(context.Background(), neo4j.SessionConfig{DatabaseName: "neo4j"})
	defer session.Close(context.Background())

	result, err := session.Run(context.Background(),
		"MATCH (h:QueryHistory {content: $content}) RETURN h.content, h.response LIMIT 1",
		map[string]interface{}{"content": content},
	)
	if err != nil {
		log.Printf("Failed to query history: %v", err)
		return QueryHistory{}, err
	}

	if result.Next(context.Background()) {
		record := result.Record()
		content, _ := record.Get("h.content")
		response, _ := record.Get("h.response")
		return QueryHistory{
			Content:  content.(string),
			Response: response.(string),
		}, nil
	}

	return QueryHistory{}, nil // No history found
}