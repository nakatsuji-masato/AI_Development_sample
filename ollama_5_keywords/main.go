package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type Product struct {
	Name        string   `json:"name"`
	Category    []string `json:"category"`
	Model       string   `json:"model"`
	Price       int      `json:"price"`
	Description string   `json:"description"`
	Metadata    []string `json:"metadata"`
}

type ProductData struct {
	Items []Product `json:"items"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

type SearchKeywords struct {
	Keywords []string `json:"keywords"`
}

type SearchResult struct {
	Product Product
	Score   int
}

const (
	ollamaURL        = "http://localhost:11434/api/generate"
	itemDataJson     = "data.json"
	scoreName        = 10
	scoreCategory    = 5
	scoreDescription = 3
	scoreMetadata    = 8
	searchMax        = 3
)

func loadProducts() ([]Product, error) {
	data, err := os.ReadFile(itemDataJson)
	if err != nil {
		return nil, err
	}

	var productData ProductData
	err = json.Unmarshal(data, &productData)
	if err != nil {
		return nil, err
	}

	return productData.Items, nil
}

// LLMで検索キーワードを抽出
func extractKeywords(query string) ([]string, error) {

	prompt := fmt.Sprintf(`
あなたは商品検索のキーワード抽出エンジンです。

ユーザーの検索文から、商品検索に使える重要なキーワードを抽出してください。

ルール:
- 商品名、カテゴリ、特徴、用途、属性を抽出する
- 漢字、ひらがな、カタカナ、英数字をすべて対象とする
- 検索意図に関係する重要な単語を優先する
- 元の検索文に存在する表記を維持する
- キーワードは重複させない
- 同義語や言い換えも抽出する
- 日本語と英語の両方が有効な場合は含める
- 検索に不要な表現は除外する
- JSON以外は出力しない

出力形式:
{"keywords":["キーワード1","キーワード2"]}

検索文:
%s`, query)

	reqBody := map[string]interface{}{
		"model":  "llama3.2",
		"prompt": prompt,
		"stream": false,
		"format": "json",
		"options": map[string]interface{}{
			"temperature": 0,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Post(
		ollamaURL,
		"application/json",
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf(
			"Ollama error: %s",
			string(responseBody),
		)
	}

	var ollamaResp OllamaResponse

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, err
	}

	var keywords SearchKeywords

	err = json.Unmarshal(
		[]byte(ollamaResp.Response),
		&keywords,
	)
	if err != nil {
		return nil, err
	}

	return keywords.Keywords, nil
}

// 商品の検索スコア計算
func calculateScore(
	product Product,
	keywords []string,
) int {

	score := 0

	for _, keyword := range keywords {

		if keyword == "" {
			continue
		}

		k := strings.ToLower(keyword)

		// 商品名
		if strings.Contains(
			strings.ToLower(product.Name), k,
		) {
			score += scoreName
		}

		// カテゴリ
		for _, category := range product.Category {
			if strings.Contains(
				strings.ToLower(category), k,
			) {
				score += scoreCategory
			}
		}

		// 商品説明
		if strings.Contains(
			strings.ToLower(product.Description), k,
		) {
			score += scoreDescription
		}

		// メタデータ
		for _, metadata := range product.Metadata {
			if strings.Contains(
				strings.ToLower(metadata), k,
			) {
				score += scoreMetadata
			}
		}
	}

	return score
}

// 商品検索
func searchProducts(
	products []Product,
	keywords []string,
) []SearchResult {

	var results []SearchResult

	for _, product := range products {

		score := calculateScore(product, keywords)

		if score > 0 {
			results = append(results, SearchResult{
				Product: product,
				Score:   score,
			})
		}
	}

	// スコアの高い順
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results
}

func main() {

	// 商品データ読み込み
	products, err := loadProducts()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {

		fmt.Print("\n商品検索: ")

		if !scanner.Scan() {
			break
		}

		query := strings.TrimSpace(scanner.Text())

		if query == "" {
			continue
		}
		// LLMで検索キーワード抽出
		keywords, err := extractKeywords(query)
		if err != nil {
			panic(err)
		}

		fmt.Println("抽出キーワード:", keywords)

		// 商品検索
		results := searchProducts(products, keywords)

		fmt.Println("\n検索結果")

		for i, result := range results {

			fmt.Printf(
				"%d. %s / %s / %d円 / Score:%d\n",
				i+1,
				result.Product.Name,
				result.Product.Model,
				result.Product.Price,
				result.Score,
			)
		}

	}

}
