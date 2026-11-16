# Ollama (オフラインAI開発サンプル)

Ollama + llama3.2 を使ったオフラインでのAI開発テストサンプル

# 使い方

## Ollamaのインストール

下記からDownload manuallyボタンを押したらインストーラーのダウンロードができる

[https://ollama.com/download](https://ollama.com/download)

インストールが完了したら下記コマンドで  
Ollamaがインストールされていることを確認

```
$ ollama --version
```

## AIモデル(llama3.2)のインストール

次にAIモデルが必要になる  
テストサンプルではllama3.2固定にしているので  
下記コマンドでllama3.2をインストール

```
$ ollama pull llama3.2
```

なお上記AIモデルは4GB必要

## Ollamaの実行

下記コマンドでOllamaを起動する

```
$ ollama serve
```

## テストサンプル(バイナリexe)の実行

``ollama.exe``があるのでそれを普通に実行する  
コンソール画面にて下記が出たらOK

```
--- Ollama AI Model Test Sample ----
LLM = llama3.2

YOUR: 
```

下記のように質問をしたらAIから答えが出てくる

```
YOUR: 3+2はいくつになる？
AI: 3+2は5です。

YOUR:
```