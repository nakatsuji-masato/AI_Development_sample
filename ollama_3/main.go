package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type ChatResponse struct {
	Message Message `json:"message"`
	Done    bool    `json:"done"`
}

var msg = []Message{}

const url = "http://localhost:11434/api/chat"

const llmModel = "llama3.2"

func main() {
	fmt.Println("--- Ollama AI Model Test Sample 3 ----")
	fmt.Println("LLM = " + llmModel)
	fmt.Println("")

	// AIに規則性データを読み込み
	Begin()

	// 処理をループ
	for {
		// 入力メッセージを取得
		fmt.Print("Please Question: ")
		var question string
		_, err := fmt.Scan(&question)
		if err != nil {
			panic(err)
		}

		// AIへ入力メッセージを送信
		res, err := Ollama(question, "user")

		if err != nil {
			panic(err)
		}

		// AIからの結果を出力
		// fmt.Println("Role    = " + res.Message.Role)
		fmt.Println("AI: " + res.Message.Content)
		fmt.Println("")
	}

}

// AIに規則性を読み込み
func Begin() {

	now := time.Now().Format("2006-01-02 15:04:05")
	week := time.Now().Weekday().String()

	rules, err := os.ReadFile("./rules.json")
	if err != nil {
		panic(err)
	}

	//	fmt.Println(now + " " + week)

	str := `
以下の規則データに従って回答してください。

ただし優先度の高い情報を順に記載しています。
・現在日時: ` + now + `
・現在曜日: ` + week + `
・下記の規則データを準拠する
` + string(rules)

	_, err = Ollama(str, "user")
	if err != nil {
		panic(err)
	}

	//	fmt.Println(res.Message.Content)

	fmt.Print("....Read Rule OK.\n")
}

// Ollamaへのリクエスト
func Ollama(message string, role string) (*ChatResponse, error) {

	// メッセージリストに追加
	msg = append(msg, Message{
		Role:    role,
		Content: message,
	})

	// リクエストボディ作成
	reqBody := ChatRequest{
		Model:    llmModel,
		Messages: msg,
		Stream:   false,
	}

	// リクエストボディのJSON化
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 5 * time.Minute,
	}

	// AI(OllamaサーバーのHTTP API)へリクエスト開始
	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// レスポンスがOKであるかどうかの確認
	if resp.StatusCode != http.StatusOK {
		// OK出ない場合はエラー
		body, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf(
			"Ollama API error: %s",
			string(body),
		))
	}

	var result ChatResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// AIからのメッセージをメッセージリストに追加
	msg = append(msg, result.Message)

	return &result, nil
}
