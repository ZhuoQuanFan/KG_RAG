package services

import (
	"context"
	"fmt"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type KGResult struct {
	Nodes   []map[string]string
	Links   []map[string]string
	Context string
	Err     error
}

func QueryKnowledgeGraph(driver neo4j.DriverWithContext, query string) KGResult {
	session := driver.NewSession(context.Background(), neo4j.SessionConfig{})
	defer session.Close(context.Background())

	result, err := session.Run(
		context.Background(),
		"MATCH (e:Entity {name: $name})-[r]->(n) "+
			"RETURN e.name AS source, r.original_name AS rel, n.name AS target, n.description AS desc",
		map[string]interface{}{"name": query},
	)
	if err != nil {
		return KGResult{Err: err}
	}

	nodes := make([]map[string]string, 0)
	links := make([]map[string]string, 0)
	nodeSet := make(map[string]bool)
	kgContext := "Knowledge Graph Information:\n"

	for result.Next(context.Background()) {
		record := result.Record()
		source, _ := record.Get("source")
		relType, _ := record.Get("rel")
		target, _ := record.Get("target")
		desc, _ := record.Get("desc")
		if desc == nil {
			desc = "No description"
		}
		if relType == nil {
			relType = "Unknown relation"
		}

		sourceNode := map[string]string{"id": source.(string), "label": source.(string)}
		targetNode := map[string]string{"id": target.(string), "label": target.(string), "description": desc.(string)}
		if !nodeSet[source.(string)] {
			nodes = append(nodes, sourceNode)
			nodeSet[source.(string)] = true
		}
		if !nodeSet[target.(string)] {
			nodes = append(nodes, targetNode)
			nodeSet[target.(string)] = true
		}

		links = append(links, map[string]string{
			"source": source.(string),
			"target": target.(string),
			"label":  relType.(string),
		})

		kgContext += fmt.Sprintf("- %s %s %s (%s)\n", source, relType, target, desc)
	}

	if err := result.Err(); err != nil {
		return KGResult{Err: err}
	}

	return KGResult{Nodes: nodes, Links: links, Context: kgContext, Err: nil}
}

func StoreToNeo4j(driver neo4j.DriverWithContext, entity1, entity2, relation string) error {
	session := driver.NewSession(context.Background(), neo4j.SessionConfig{})
	defer session.Close(context.Background())

	_, err := session.Run(
		context.Background(),
		"MERGE (e:Entity {name: $name}) SET e.description = $desc",
		map[string]interface{}{
			"name": entity1,
			"desc": "No description",
		},
	)
	if err != nil {
		return err
	}

	_, err = session.Run(
		context.Background(),
		"MERGE (e:Entity {name: $name}) SET e.description = $desc",
		map[string]interface{}{
			"name": entity2,
			"desc": "No description",
		},
	)
	if err != nil {
		return err
	}

	_, err = session.Run(
		context.Background(),
		"MATCH (e1:Entity {name: $name1}), (e2:Entity {name: $name2}) "+
			"MERGE (e1)-[r:RELATION {original_name: $original_name}]->(e2)",
		map[string]interface{}{
			"name1":         entity1,
			"name2":         entity2,
			"original_name": relation,
		},
	)
	return err
}

func ClearNeo4jDatabase(driver neo4j.DriverWithContext) error {
    session := driver.NewSession(context.Background(), neo4j.SessionConfig{})
    defer session.Close(context.Background())

    _, err := session.Run(
        context.Background(),
        "MATCH (n) DETACH DELETE n",
        nil,
    )
    if err != nil {
        log.Printf("清空 Neo4j 数据库失败: %v", err)
        return err
    }

    log.Println("Neo4j 数据库已清空")
    return nil
}