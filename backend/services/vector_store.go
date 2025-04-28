package services

import (
	"context"
	"log"

	"github.com/tmc/langchaingo/embeddings"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/pgvector"
)

type VectorStoreService struct {
	store pgvector.Store
}

func NewVectorStoreService(ctx context.Context, embedder *embeddings.EmbedderImpl, connURL string) (*VectorStoreService, error) {
	store, err := pgvector.New(
		ctx,
		pgvector.WithConnectionURL(connURL),
		pgvector.WithEmbedder(embedder),
	)
	if err != nil {
		return nil, err
	}
	return &VectorStoreService{store: store}, nil
}

func (s *VectorStoreService) AddDocuments(ctx context.Context, docs []schema.Document) error {
	_, err := s.store.AddDocuments(ctx, docs)
	return err
}

func (s *VectorStoreService) SimilaritySearch(ctx context.Context, query string, numDocs int) ([]schema.Document, error) {
	return s.store.SimilaritySearch(ctx, query, numDocs)
}

func (s *VectorStoreService) ClearPGVector(ctx context.Context) error {
	pgStore:= s.store
  

	err := pgStore.DropTables(ctx)
    // err := pgStore.DropTables(ctx)
    if err != nil {
        log.Printf("清空 PGVector 表失败: %v", err)
        return err
    }

    log.Println("PGVector 表已清空")
    return nil
}