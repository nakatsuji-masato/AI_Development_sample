# Ollama (単純なデータセットのテストサンプル)

1つのモノの特徴が入った単純なデータセットを用意した後で、  
質問をAIに投げると、データ内容からモノの特徴などを回答をするテストサンプル

データはファイルベースで``__DATA__``ファイルに保存された内容を  
質問を開始する前にAIに事前に読み込ませる

あとはその内容からモノの特徴を答える。

# 使い方

Ollama + llama3.2が必要なのでそれについては  
[../ollama/readme.md] こちらを参照

## 1. データのセット

まずはモノの特徴を覚えさせる必要がある。  
バイナリexeファイルをコンソールから下記コマンド``ollama_2.exe -i``で実行

```
$ ./ollama_2 -i
--- Ollama AI Model Test Sample 2 ----
LLM = llama3.2

DATA INPUT

Enter the data you want the AI ​​to learn here.: 
```

ここでモノの特徴を1つずつ入力する  
ただ``__DATA__``ファイルには改行ベースで保存されるだけなので直接ファイルにセットしても良い

例えばモノをトマトと仮定すると下記の特徴が挙げられる

```
答えはトマト
野菜です
色は赤いです
夏が旬です
生でも食べれます
上にへたがあります
生だと酸味がある
煮込むと甘くなります
ケチャップの原料
苦手な人も多い
原産地は南米
未熟だと青い
熟すと赤くなる
タコではない
丸い形をしている
甘い種類もある
桃太郎という品種が有名
```

これをデータセットすることで、この内容を優先的に回答する

## 2. AIへの質問を開始

バイナリexeファイル``ollama_2.exe``を実行する

```
$ ./ollama_2

--- Ollama AI Model Test Sample 2 ----
LLM = llama3.2

QUE START.

Please enter your question regarding existence X.: 
```

"Please enter your question regarding existence X" が表示されれば  
あとは質問を入力(日本語でもOK)

一度実行時にAIにデータを読み込ませるためのリクエストを投げている為  
若干次の対話に来るのに時間がかかる

後は下記のように質問を続けていくと  
AIはセットされたデータ内容からの回答をする。 

```
$ ./ollama_2

--- Ollama AI Model Test Sample 2 ----
LLM = llama3.2

QUE START.

Please enter your question regarding existence X.: 食べれる？
AI: あります。存在Xは、生でも食べることができる野菜です。

Please enter your question regarding existence X.: どんな色？
AI: 赤です。

Please enter your question regarding existence X.: 海にある？
AI: いいえ、海にありません。

Please enter your question regarding existence X.: 味はどんなの？
AI: 酸味があります。生で食べると酸味があります。

Please enter your question regarding existence X.: 料理するとどうなる？
AI: 煮込むと甘くなります。

Please enter your question regarding existence X.: 旬の時期はいつ？
AI: 夏です。
```

# 補足

- データ内容にないことを聞かれると回答を外すかもしくはよくわからないと答えるケースあり
- 答えは？と言われない限りは答えを言ってはいけないという条件をあらかじめプロンプトで入れている為  
  答えを言うことはない(と思うが...)