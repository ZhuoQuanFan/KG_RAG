package handlers

import (
	"RAG/backend/services"
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/tmc/langchaingo/schema"
)

type QueryRequest struct {
	Content string `json:"content"`
}

type QueryResponse struct {
	Answer string `json:"answer"`
	Graph  struct {
		Nodes []map[string]string `json:"nodes"`
		Links []map[string]string `json:"links"`
	} `json:"graph"`
}

type vectorResult struct {
	docs    []schema.Document
	err     error
	context string
}

func QueryHandler(store *services.VectorStoreService, driver neo4j.DriverWithContext) gin.HandlerFunc {
	return func(c *gin.Context) {

		var req QueryRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			log.Printf("Failed to parse request body: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
			return
		}

		if req.Content == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Content field cannot be empty"})
			return
		}

		queryCache := sync.Map{}
		if cached, ok := queryCache.Load(req.Content); ok {
			resp := cached.(QueryResponse)
			c.JSON(http.StatusOK, resp)
			return
		}

		vectorChan := make(chan vectorResult, 1)
		kgChan := make(chan services.KGResult, 1)
		historyChan := make(chan services.QueryHistory, 1)


		// 查询历史记录
		go func() {
		
			history, err := services.FindQueryHistory(driver, req.Content)
			historyChan <- history
			if err != nil {
				log.Printf("No matching history found: %v", err)
			}
		}()
		
		// 查询向量数据库
		go func() {
			docs, err := store.SimilaritySearch(context.Background(), req.Content, 3)
			if err != nil {
				vectorChan <- vectorResult{err: err}
				return
			}
			docContext := "Related Documents:\n"
			for _, doc := range docs {
				docContext += fmt.Sprintf("- %s\n", doc.PageContent)
			}
			vectorChan <- vectorResult{docs: docs, context: docContext}
		}()

		// 查询知识图谱
		go func() {
			kgChan <- services.QueryKnowledgeGraph(driver, req.Content)
		}()

		var fastResult, docContext, kgContext string
		var nodes []map[string]string
		var links []map[string]string


		
		var wg sync.WaitGroup
		wg.Add(1)

		flag := false

		// 查询历史记录
		go func(){
			defer wg.Done()
			hr := <-historyChan
			if hr.Response != "" {
				fastResult = hr.Response
				log.Println(fastResult)
				resp := QueryResponse{Answer: fastResult}
				resp.Graph.Nodes = nodes
				resp.Graph.Links = links
				// 历史存在的不用再保存历史
				// services.SaveQueryHistory(driver, req.Content, fastResult)
				c.JSON(http.StatusOK, resp)
				// 找到历史记录就提前返回，不再执行下面的逻辑
				flag = true
				return
			}
		}()
		wg.Wait()	
		if flag{
			return
		}

		wg.Add(2)

		// 查询向量数据库
		go func(){
			defer wg.Done()
			vr := <-vectorChan
			if vr.err != nil {
				log.Printf("Vector search failed: %v", vr.err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
			}else{
				docContext = vr.context
			}
		}()
		
		// 查询知识图谱
		go func(){
			defer wg.Done()
			kr := <-kgChan
			if kr.Err != nil {
				log.Printf("Neo4j query failed: %v", kr.Err)
				kgContext = "Unable to retrieve knowledge graph information\n"
			} else {
				kgContext = kr.Context
				nodes = kr.Nodes
				links = kr.Links
			}
		}()

		wg.Wait()

		// 整合结果并调用 LLM
		if docContext == "" && kgContext == "" {
			c.JSON(http.StatusOK, QueryResponse{Answer: "No relevant information found."})
			return
		}

		prompt := fmt.Sprintf(
			"You are a professional AWS technical assistant. Provide a concise and accurate answer based on the following documents and knowledge graph. If no relevant content is found, apologize and state that you cannot answer.Answer in Chinese. \n\n%s\n%s\nQuestion: %s\nAnswer:",
			docContext, kgContext, req.Content)

		response,err := services.Query(prompt)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": response})
			return
		}

		resp := QueryResponse{Answer: response}
		resp.Graph.Nodes = nodes
		resp.Graph.Links = links

		queryCache.Store(req.Content, resp)
		services.SaveQueryHistory(driver, req.Content, response)
		c.JSON(http.StatusOK, resp)
	}
}