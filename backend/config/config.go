package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/llms/ollama"
)

type Config struct {
	Embedder        *embeddings.EmbedderImpl // Changed from *embeddings.Embedder to embeddings.Embedder
	Neo4jDriver     neo4j.DriverWithContext
	PGVectorConnURL string
}

func NewConfig() (*Config, error) {
	godotenv.Load()
	// Initialize Ollama embedder
	ollamaEmbedder, err := ollama.New(
		ollama.WithModel("nomic-embed-text"),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Ollama embedder: %v", err)
		return nil, err
	}
	embedder, err := embeddings.NewEmbedder(ollamaEmbedder)
	if err != nil {
		log.Fatalf("Failed to create embedder: %v", err)
		return nil, err
	}

	// Initialize Neo4j driver
	neo4jDriver, err := neo4j.NewDriverWithContext(
		"bolt://localhost:7687",
		neo4j.BasicAuth(os.Getenv("NEO4J_USERNAME"), os.Getenv("NEO4J_PASSWORD"),""),
	)
	if err != nil {
		log.Fatalf("Failed to initialize Neo4j driver: %v", err)
		return nil, err
	}
	log.Println("postgres://"+os.Getenv("Postgres_Username")+":"+os.Getenv("Postgres_Password")+"@localhost:5432/"+os.Getenv("Postgres_DBName"))

	return &Config{
		Embedder:        embedder, // Assign embedder directly (type *embeddings.EmbedderImpl implements embeddings.Embedder)
		Neo4jDriver:     neo4jDriver,
		PGVectorConnURL: "postgres://"+os.Getenv("Postgres_Username")+":"+os.Getenv("Postgres_Password")+"@localhost:5432/"+os.Getenv("Postgres_DBName"),
	}, nil
}
