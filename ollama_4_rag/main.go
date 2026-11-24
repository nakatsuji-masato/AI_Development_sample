package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	OllamaURL  = "http://localhost:11434"
	LLMModel   = "llama3.2"
	EmbedModel = "nomic-embed-text"

	DataFile     = "data.json"
	CategoryFile = "category.json"
	VectorDBFile = "vector_db/vectors.json"

	ProductTopK = 5
)

var httpClient = &http.Client{
	Timeout: 180 * time.Second,
}

// ------------------------------------
// データ構造
// ------------------------------------

// 商品データ
type Product struct {
	Name        string   `json:"name"`
	Category    []string `json:"category"`
	Model       string   `json:"model"`
	Price       int      `json:"price"`
	Description string   `json:"description"`
}

type ProductData struct {
	Items []Product `json:"items"`
}

// カテゴリーデータ
type Category struct {
	Name     string   `json:"name"`
	Keywords []string `json:"keywords"`
}

// ベクトルDB
type VectorEntry struct {
	Kind    string          `json:"kind"`
	Name    string          `json:"name"`
	Content string          `json:"content"`
	Payload json.RawMessage `json:"payload"`
	Vector  []float64       `json:"vector"`
}

type VectorDB struct {
	Entries []VectorEntry `json:"entries"`
}

// Embedding API
type EmbedRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type EmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
}

// Chat API
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Message Message `json:"message"`
}

// 検索結果
type SearchResult struct {
	Entry VectorEntry
	Score float64
}

// ------------------------------------
// JSON読み込み
// ------------------------------------

func loadJSON(path string, v interface{}) error {

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// ------------------------------------
// Ollama Embedding
// ------------------------------------

func getEmbeddings(texts []string) ([][]float64, error) {

	reqBody := EmbedRequest{
		Model: EmbedModel,
		Input: texts,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		OllamaURL+"/api/embed",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"embedding API error: %s",
			string(b),
		)
	}

	var result EmbedResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Embeddings) != len(texts) {
		return nil, fmt.Errorf("embedding count mismatch")
	}

	for _, v := range result.Embeddings {
		if len(v) == 0 {
			return nil, fmt.Errorf("empty embedding vector")
		}
	}

	return result.Embeddings, nil
}

// ------------------------------------
// 初回ベクトルDB作成
// ------------------------------------

func createVectorDB() (*VectorDB, error) {

	var products ProductData
	var categories []Category

	// 商品データ読み込み
	if err := loadJSON(DataFile, &products); err != nil {
		return nil, err
	}

	// カテゴリーデータ読み込み
	if err := loadJSON(CategoryFile, &categories); err != nil {
		return nil, err
	}

	db := &VectorDB{}

	var texts []string

	// --------------------------------
	// 商品データ登録
	// --------------------------------

	for _, p := range products.Items {

		content := fmt.Sprintf(
			`
			商品名: %s カテゴリー: %s 型番: %s 価格: %d円 説明: %s
			
			各項目別スコア:
			・カテゴリー: x10
			・商品名: x0.7
			・説明: x0.5
			・価格: x0.3
			`,
			p.Name,
			strings.Join(p.Category, " "),
			p.Model,
			p.Price,
			p.Description,
		)

		payload, err := json.Marshal(p)
		if err != nil {
			return nil, err
		}

		db.Entries = append(db.Entries, VectorEntry{
			Kind:    "product",
			Name:    p.Name,
			Content: content,
			Payload: payload,
		})

		texts = append(texts, content)
	}

	// --------------------------------
	// カテゴリーデータ登録
	// --------------------------------

	for _, c := range categories {

		content := fmt.Sprintf(`		
			カテゴリー: 
			%s 
			
			関連キーワード: 
			%s
			`,
			c.Name,
			strings.Join(c.Keywords, " "),
		)

		payload, err := json.Marshal(c)
		if err != nil {
			return nil, err
		}

		db.Entries = append(db.Entries, VectorEntry{
			Kind:    "category",
			Name:    c.Name,
			Content: content,
			Payload: payload,
		})

		texts = append(texts, content)
	}

	// --------------------------------
	// 一括Embedding
	// --------------------------------

	fmt.Println("ベクトルを作成しています...")

	vectors, err := getEmbeddings(texts)
	if err != nil {
		return nil, err
	}

	for i := range db.Entries {
		db.Entries[i].Vector = vectors[i]
	}

	// --------------------------------
	// ベクトルDB保存
	// --------------------------------

	if err := os.MkdirAll("vector_db", 0755); err != nil {
		return nil, err
	}

	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(VectorDBFile, data, 0644); err != nil {
		return nil, err
	}

	fmt.Println("ベクトルDBを作成しました")

	return db, nil
}

// ------------------------------------
// ベクトルDB読み込み
// ------------------------------------

func loadOrCreateVectorDB() (*VectorDB, error) {

	if _, err := os.Stat(VectorDBFile); err == nil {

		fmt.Println("既存のベクトルDBを読み込みます")

		var db VectorDB

		if err := loadJSON(VectorDBFile, &db); err != nil {
			return nil, err
		}

		if len(db.Entries) == 0 {
			return nil, fmt.Errorf("vector DB is empty")
		}

		// ベクトル次元チェック
		dimension := len(db.Entries[0].Vector)

		for _, entry := range db.Entries {
			if dimension == 0 || len(entry.Vector) != dimension {
				return nil, fmt.Errorf("invalid vector dimension")
			}
		}

		return &db, nil
	}

	// ベクトルDBが存在しない場合は作成
	fmt.Println("ベクトルDBがありません。初回作成します")

	return createVectorDB()
}

// ------------------------------------
// Cosine Similarity
// ------------------------------------

func cosineSimilarity(a, b []float64) float64 {

	if len(a) != len(b) || len(a) == 0 {
		return -1
	}

	var dot float64
	var normA float64
	var normB float64

	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return -1
	}

	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// ------------------------------------
// カテゴリー検索
// ------------------------------------

func searchCategories(
	db *VectorDB,
	queryVector []float64,
	topK int,
) []SearchResult {

	var results []SearchResult

	for _, entry := range db.Entries {

		// カテゴリー以外は除外
		if entry.Kind != "category" {
			continue
		}

		score := cosineSimilarity(
			queryVector,
			entry.Vector,
		)

		if score < 0 {
			continue
		}

		results = append(results, SearchResult{
			Entry: entry,
			Score: score,
		})
	}

	// 類似度の高い順
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

// ------------------------------------
// カテゴリー一致判定
// ------------------------------------

func belongsToCategory(
	product Product,
	categoryName string,
) bool {

	for _, category := range product.Category {

		if category == categoryName {
			return true
		}
	}

	return false
}

// ------------------------------------
// 指定カテゴリー内の商品検索
// ------------------------------------

func searchProductsByCategory(
	db *VectorDB,
	queryVector []float64,
	categoryName string,
	topK int,
) []SearchResult {

	var results []SearchResult

	for _, entry := range db.Entries {

		// 商品以外は除外
		if entry.Kind != "product" {
			continue
		}

		// 商品情報を取得
		var product Product

		if err := json.Unmarshal(
			entry.Payload,
			&product,
		); err != nil {
			continue
		}

		// 選択カテゴリーに属さない商品は除外
		if !belongsToCategory(product, categoryName) {
			continue
		}

		// 商品の類似度を計算
		score := cosineSimilarity(
			queryVector,
			entry.Vector,
		)

		if score < 0 {
			continue
		}

		results = append(results, SearchResult{
			Entry: entry,
			Score: score,
		})
	}

	// 類似度の高い順
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

// ------------------------------------
// Llama 3.2で回答生成
// ------------------------------------

func generateAnswer(
	query string,
	categoryName string,
	products []SearchResult,
) (string, error) {

	var context strings.Builder

	context.WriteString(
		"【選択されたカテゴリー】\n",
	)

	context.WriteString(categoryName + "\n")

	context.WriteString(
		"\n【検索された商品】\n",
	)

	for _, result := range products {
		context.WriteString(
			string(result.Entry.Payload),
		)
		context.WriteString("\n")
	}

	systemPrompt := `
あなたは商品検索専門のAIアシスタントです。

以下の検索結果だけを根拠に回答してください。

ルール:
- 検索結果に存在しない商品を紹介しないでください。
- 商品名、価格、型番、説明は検索結果を使用してください。
- ユーザーの希望に近い商品を優先してください。
- 該当商品がない場合は、その旨を回答してください。
- 価格は円で表示してください。
- 回答は日本語で行ってください。
`

	userPrompt := fmt.Sprintf(`
【ユーザーの質問】
%s

【RAG検索結果】
%s

検索結果を使って回答してください。
`, query, context.String())

	reqBody := ChatRequest{
		Model: LLMModel,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		Stream: false,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		"POST",
		OllamaURL+"/api/chat",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf(
			"chat API error: %s",
			string(b),
		)
	}

	var result ChatResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Message.Content, nil
}

// ------------------------------------
// 商品検索メイン処理
// ------------------------------------

func searchProduct(
	db *VectorDB,
	query string,
) error {

	fmt.Println("\n検索キーワード:", query)

	// --------------------------------
	// ① 検索キーワードをEmbedding
	// --------------------------------

	vectors, err := getEmbeddings([]string{query})
	if err != nil {
		return err
	}

	queryVector := vectors[0]

	// --------------------------------
	// ② カテゴリー類似度検索
	// --------------------------------

	categoryResults := searchCategories(
		db,
		queryVector,
		3,
	)

	if len(categoryResults) == 0 {
		fmt.Println("カテゴリーが見つかりません")
		return nil
	}

	fmt.Println("\n--- カテゴリー検索結果 ---")

	for _, result := range categoryResults {
		fmt.Printf(
			"%s : %.4f\n",
			result.Entry.Name,
			result.Score,
		)
	}

	// --------------------------------
	// ③ 最も類似度が高いカテゴリー
	// --------------------------------

	selectedCategory := categoryResults[0].Entry.Name

	fmt.Println(
		"\n選択カテゴリー:",
		selectedCategory,
	)

	// --------------------------------
	// ④ 選択カテゴリー内の商品検索
	// --------------------------------

	productResults := searchProductsByCategory(
		db,
		queryVector,
		selectedCategory,
		ProductTopK,
	)

	if len(productResults) == 0 {
		fmt.Println("該当する商品がありません")
		return nil
	}

	fmt.Println("\n--- 商品検索結果 ---")

	for _, result := range productResults {

		fmt.Printf(
			"%s : %.4f\n",
			result.Entry.Name,
			result.Score,
		)
	}

	// --------------------------------
	// ⑤ Llama 3.2で回答生成
	// --------------------------------

	answer, err := generateAnswer(
		query,
		selectedCategory,
		productResults,
	)

	if err != nil {
		return err
	}

	fmt.Println("\n--- AI回答 ---")
	fmt.Println(answer)

	return nil
}

// ------------------------------------
// メイン
// ------------------------------------

func main() {

	// ベクトルDB読み込みまたは初回作成
	db, err := loadOrCreateVectorDB()

	if err != nil {
		fmt.Println("エラー:", err)
		return
	}

	fmt.Println("\n商品検索AIを起動しました")
	fmt.Println("終了するには exit と入力してください")

	scanner := bufio.NewScanner(os.Stdin)

	for {

		fmt.Print("\n検索キーワード > ")

		if !scanner.Scan() {
			break
		}

		query := strings.TrimSpace(scanner.Text())

		if query == "" {
			continue
		}

		if strings.EqualFold(query, "exit") {
			break
		}

		if err := searchProduct(db, query); err != nil {
			fmt.Println("検索エラー:", err)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("入力エラー:", err)
	}

	fmt.Println("終了しました")
}
