package services

import (
	"context"
	"log"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)


func Query(prompt string)(string,error){
	var response string
	var err error

	llm, err := openai.New(
		openai.WithBaseURL(os.Getenv("OPENAI_API_BASE")),
		openai.WithToken(os.Getenv("OPENAI_API_KEY")),
		openai.WithModel(os.Getenv("OPENAI_MODEL")),
	)
	if err != nil {
		log.Printf("Failed to initialize OpenAI client: %v", err)
		response = "Failed to initialize AI model"
		return response,err
	}

	response, err = llm.Call(context.Background(), prompt, llms.WithTemperature(0.7))
	if err != nil {
		log.Printf("ChatGPT call failed: %v", err)
		response = "Failed to generate response"
		return response,err
	}
	return response,err
}
