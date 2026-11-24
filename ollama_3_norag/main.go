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

const url = "http://localhost:11434/api/chat"

const llmModel = "llama3.2"

func main() {
	fmt.Println("--- Ollama AI Model Test Sample 4(RAG) ----")
	fmt.Println("LLM = " + llmModel)
	fmt.Println("")

	// 処理をループ
	for {

		var msg = []Message{}

		// 入力メッセージを取得
		fmt.Print("Please Question: ")
		var query string
		_, err := fmt.Scan(&query)
		if err != nil {
			panic(err)
		}

		// AIへ入力メッセージを送信
		res, err := question(query, msg)

		if err != nil {
			panic(err)
		}

		// AIからの結果を出力
		// fmt.Println("Role    = " + res.Message.Role)
		fmt.Println("AI: " + res.Message.Content)
		fmt.Println("")
	}

}

func question(question string, msg []Message) (*ChatResponse, error) {

	catList, err := os.ReadFile("./category.json")
	if err != nil {
		panic(err)
	}

	p0 := fmt.Sprintf(`
あなたは商品カテゴリを分類するアシスタントです。

カテゴリリスト:
%s

質問:
%s

ルール：
1. 必ず上記5つのカテゴリリストから2つだけ選択すること。
2. 該当するカテゴリが判断できない場合は「その他」を選択すること。
3. 説明文は不要。
4. JSON形式のみで回答すること。

回答形式：
["カテゴリ名1","カテゴリ名2"]`, catList, question)

	res, err := Ollama(msg, p0, "user")
	if err != nil {
		panic(err)
	}

	items, err := os.ReadFile("./data.json")

	if err != nil {
		panic(err)
	}

	p1 := fmt.Sprintf(`
あなたは商品検索を分類するAIです。
	
カテゴリ:
%s

商品データ:
%s

質問:
%s

ルール：
1. 上記カテゴリに該当する商品情報より質問内容に合致する商品を1つ選択し、上記質問内容に回答すること

回答形式:
自由(商品名、価格、概要など文章は考えてください)
`, res.Message.Content, items, question)

	res2, err := Ollama(msg, p1, "user")
	if err != nil {
		panic(err)
	}

	return res2, err
}

// Ollamaへのリクエスト
func Ollama(msg []Message, message string, role string) (*ChatResponse, error) {

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
