package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/responses"
)

const llmModel = "gpt-6-luna"

func main() {

	keyBytes, err := os.ReadFile(".env")

	if err != nil {
		panic(err)
	}

	if keyBytes == nil {
		fmt.Println("Error: Secret key Not Found.")
		return
	}

	apiKey := strings.TrimSpace(string(keyBytes))

	fmt.Println("--- OpenAI Test Sample ----")
	fmt.Println("Model = " + llmModel)
	fmt.Println("")

	// 会話ループ
	for {

		// 入力メッセージを取得
		fmt.Print("YOUR: ")
		var question string
		_, err := fmt.Scan(&question)
		if err != nil {
			panic(err)
		}

		// OpenAIに質問を送信
		res, err := OpenAI(apiKey, question)

		if err != nil {
			log.Fatalf("OpenAI API Error: %v", err)
		}

		// AIの回答を表示
		fmt.Println("AI: " + res.OutputText())
		fmt.Println("-------------------------")
		fmt.Println("Usage Token")
		fmt.Println("Input  = ", res.Usage.InputTokens)
		fmt.Println("Output = ", res.Usage.OutputTokens)
		fmt.Println("Total  = ", res.Usage.TotalTokens)
		fmt.Println("-------------------------")
		fmt.Println("")
	}
}

func OpenAI(apiKey string, message string) (*responses.Response, error) {

	// OpenAIクライアント作成
	client := openai.NewClient(option.WithAPIKey(apiKey))

	// OpenAI APIに質問を送信
	resp, err := client.Responses.New(
		context.Background(),
		responses.ResponseNewParams{
			Model: llmModel,
			Input: responses.ResponseNewParamsInputUnion{
				OfString: openai.String(
					message,
				),
			},
		},
	)

	return resp, err
}
