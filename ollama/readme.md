# Ollama (オフラインAI開発サンプル)

Ollama + llama3.2 を使ったオフラインでのAI開発テストサンプル

# 使い方

## 1. Ollamaのインストール

下記からDownload manuallyボタンを押したらインストーラーのダウンロードができる

[https://ollama.com/download](https://ollama.com/download)

インストールが完了したら下記コマンドで  
Ollamaがインストールされていることを確認

```
$ ollama --version
```

## 2. AIモデル(llama3.2)のダウンロード

次にAIモデルが必要になる  
テストサンプルではllama3.2固定にしているので  
下記コマンドでllama3.2をダウンロード

```
$ ollama pull llama3.2
```

なお上記AIモデルは4GB必要

## 3. Ollamaの実行

下記コマンドでOllamaを起動する

```
$ ollama serve
```

## 4. テストサンプル(バイナリexe)の実行

Windows端末上であればすでに``ollama.exe``があるのでそれを普通に実行する  
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

以後はループとなる

# 補足

- ollamaサーバーを起動した状態でollama.exeを実行すること  
  ollamaサーバーが停止している場合は質問を投げた時点で強制終了する仕様
- 端末のスペックや質問の内容次第では30秒以上かかることもある  
  (どうもCPUよりもGPUを優先的に使うらしい)
- gemma等の別のAIモデルで試したい場合は``main.go``のllmModelの変更と  
  ``ollama pull``コマンドでAIモデルをダウンロードしたら可能  
  (ただしこのサンプルはGo言語なので再ビルドが必要)
- このサンプルでは単にOllamaのHTTP APIでのリクエスト送受信を行っただけなので  
  別の言語に置き換えても可能
