# Ollama (RAGテストサンプル)

規則性のJSONデータを用意して  
それをAIに読み込ませ、その規則性からAIに質問を投げて回答してもらう。

# 規則性データについて

``rules.json``にて商品情報が入っている。  
この内容に即した質問をするとAIが答えるはず。

# 使い方

Ollama + llama3.2が必要なのでそれについては  
[../ollama/readme.md] こちらを参照

バイナリexeファイル``ollama_3.exe``を実行する

```
$ ./ollama_3

--- Ollama AI Model Test Sample 3 ----
LLM = llama3.2

....Read Rule OK.
Please Question: 
```

以後、商品に関する質問を入力する

```
$ ./ollama_3

--- Ollama AI Model Test Sample 3 ----
LLM = llama3.2

....Read Rule OK.
Please Question: ワイヤレスイヤホンの商品の要約してまとめて

AI: ワイヤレスイヤホン ABC

*   商品名：ワイヤレスイヤホン ABC
*   価格：5980円
*   特徴：Bluetooth 5.3対応、ノイズキャンセリング搭載、最大30時間再生、防水対応
*   説明：通勤や通学、運動時にも使いやすいワイヤレスイヤホンです。

Please Question: ワイヤレスイヤホンの再生時間はどのくらい？
AI: ワイヤレスイヤホンの再生時間は最大30時間です。
```
