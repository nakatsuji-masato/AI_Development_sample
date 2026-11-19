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

const setDataPath = "./___DATA___"

var msg = []Message{}

func main() {

	fmt.Println("--- Ollama AI Model Test Sample 2 ----")
	fmt.Println("LLM = " + llmModel)
	fmt.Println("")

	args := os.Args

	if len(args) > 1 {
		if args[1] == "-i" || args[1] == "-input" {
			Input()
			return
		}
	}

	Ques()
}

// データのインプット
func Input() {

	fmt.Println("DATA INPUT")
	fmt.Println("")

	for {
		// 入力メッセージを取得
		fmt.Print("Enter the data you want the AI ​​to learn here.: ")
		var setData string
		_, err := fmt.Scan(&setData)
		if err != nil {
			panic(err)
		}

		// ファイルベースでデータに追記
		file, err := os.OpenFile(setDataPath,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0644,
		)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		_, err = file.WriteString(setData + "\n")
		if err != nil {
			panic(err)
		}

		fmt.Println(".... I've memorized it.")
	}
}

func Ques() {
	fmt.Println("QUE START.")
	fmt.Println("")

	Begin()

	// 処理をループ
	for {
		// 入力メッセージを取得
		fmt.Print("Please enter your question regarding existence X.: ")
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

// 初期段階の学習用リクエスト
func Begin() {
	// 初期段階で学習
	getData, err := os.ReadFile(setDataPath)

	if err != nil {
		panic(err)
	}

	str := `
下記内容は存在Xについてのデータです。

` + string(getData) + `

これから質問をしますが、下記の条件を守って答えてください。
・答えは何ですか？と聞かれるまでは存在Xの正体を言わないでください
`

	_, err = Ollama(str, "user")
	if err != nil {
		panic(err)
	}

	// AIからの結果を出力
	// fmt.Println("Role    = " + res.Message.Role)
	//	fmt.Println("AI: " + res.Message.Content)
	//	fmt.Println("")
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
