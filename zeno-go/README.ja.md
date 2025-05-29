# Zeno-Go コンパイラ

Zeno プログラミング言語から Go への変換を行うコンパイラです。Rust版からの移植作業の一環として開発されています。

## 特徴

- **TypeScript風のImport文**: `import {println} from "std/fmt"` のような構文をサポート
- **関数定義と呼び出し**: パラメータ、戻り値型、return文をサポート
- **未使用変数検出**: コンパイル時に未使用変数を検出してエラーを出力
- **Import検証**: 関数が適切にimportされているかをチェック
- **バイナリ式サポート**: 数学演算（+, -, *, /, %）と比較演算子をサポート
- **型注釈**: オプションの型注釈 `let x: int = 42;`
- **多言語エラーメッセージ**: `-jp` フラグで日本語エラーメッセージも表示
- **変数宣言**: letキーワードによる変数宣言をサポート

## インストール

```bash
cd zeno-go
go build ./cmd/zeno-compiler
```

## 使用方法

### 基本的な使用方法

```bash
# Zenoファイルをコンパイルして実行
./zeno-compiler run example.zeno

# Zenoファイルをコンパイル（.goファイルを生成）
./zeno-compiler compile example.zeno

# 後方互換性：直接ファイル名を指定してコンパイル
./zeno-compiler example.zeno

# 日本語エラーメッセージも表示
./zeno-compiler -jp run example.zeno
```

### Zeno言語の例

#### 基本的な例
```zeno
import {println} from "std/fmt";

let x = 10;
let y = 20;
let result = x + y;
println(result);
```

#### 関数の例
```zeno
import {println} from "std/fmt";

fn add(a: int, b: int): int {
    return a + b;
}

fn greet(name: string) {
    println("Hello, " + name + "!");
}

fn main() {
    let result: int = add(5, 3);
    println("Result: ", result);
    
    greet("Zeno");
}
```

生成されるGoコード:

```go
package main

import (
	"fmt"
)

func add(a int64, b int64) int64 {
	return (a + b)
}

func greet(name string) {
	fmt.Println(("Hello, " + name + "!"))
}

func main() {
	var result int64 = add(5, 3)
	fmt.Println("Result: ", result)
	greet("Zeno")
}
```

## サポートされている構文

### Import文
```zeno
import {println, print} from "std/fmt";
```

### 変数宣言
```zeno
let x = 42;           // 変数宣言
let y: int = 100;     // 型注釈付き
```

### 関数定義
```zeno
fn add(a: int, b: int): int {
    return a + b;
}

fn greet(name: string) {
    println("Hello, " + name);
}
```

### 関数呼び出し
```zeno
let result = add(10, 20);
greet("World");
```

### main関数
```zeno
fn main() {
    // プログラムのエントリーポイント
    println("Hello, World!");
}
```

### バイナリ式
```zeno
let sum = 10 + 20;
let product = 5 * 6;
let comparison = x > y;
```

### Print文
```zeno
print("Hello");       // import {print} from "std/fmt"; が必要
println("World");     // import {println} from "std/fmt"; が必要
```

## エラー検出機能

### 未使用変数の検出
```zeno
import {println} from "std/fmt";

let x = 10;
let unused = 42;  // エラー: Unused variables found: unused
let y = x + 5;
println(y);
```

### Import検証
```zeno
// エラー: println is not imported from std/fmt
let x = 10;
println(x);  // import文がない場合はエラー
```

## 標準ライブラリ

現在サポートされているモジュール:

- `std/fmt`: `print`, `println` 関数
- `std/io`: `readFile`, `writeFile` 関数

### std/io モジュールの使用法

`std/io` モジュールは、シンプルで直感的なファイルI/O操作を提供します：

```zeno
import { println } from "std/fmt";
import { readFile, writeFile } from "std/io";

fn main() {
    // ファイルにコンテンツを書き込み
    let content = "こんにちは、Zeno!\nテストファイルです。";
    writeFile("example.txt", content);
    println("ファイルが正常に書き込まれました！");
    
    // ファイルからコンテンツを読み込み
    let fileContent = readFile("example.txt");
    println("ファイルの内容:");
    println(fileContent);
    
    // 構造化データの書き込み
    let jsonData = "{\"name\": \"Zeno\", \"version\": \"1.0\"}";
    writeFile("config.json", jsonData);
    
    let configData = readFile("config.json");
    println("設定: ", configData);
}
```

#### std/io 関数

- `writeFile(filename: string, content: string)`: 自動エラーハンドリング付きでファイルにコンテンツを書き込み
- `readFile(filename: string): string`: ファイル内容を読み込んで文字列として返却、エラー時は空文字列を返却

## 実装されている機能

✅ **完了済み:**
- Import文の解析と検証
- 変数宣言（let）
- バイナリ式（算術演算、比較演算）
- Print文の変換
- 未使用変数検出
- 多言語エラーメッセージ（英語/日本語）
- トークン解析（Lexer）
- AST構築（Parser）
- Goコード生成（Generator）

🔲 **今後の予定:**
- 関数定義と呼び出し
- 制御フロー（if/else、while、loop）
- 可変変数（mut）
- 型システムの拡張
- 標準ライブラリの拡充

## コンパイラの使用方法

### コンパイラのビルド

```bash
cd zeno-go
go build ./cmd/zeno-compiler
```

### 基本的な使用方法

```bash
# Zenoファイルをコンパイルして実行
./zeno-compiler run example.zeno

# Zenoファイルをコンパイル（.goファイルを生成）
./zeno-compiler compile example.zeno

# 日本語エラーメッセージも表示
./zeno-compiler -jp run example.zeno
```

### テストファイルの例

プロジェクトには以下のテストファイルが含まれています：

- `test_simple.zeno` - 正常動作のテスト
- `test_import.zeno` - Import文のテスト
- `test_unused.zeno` - 未使用変数検出のテスト
- `test_no_import.zeno` - Import不足エラーのテスト

## 開発ツール

デバッグ用のツールも含まれています：

- `debug_lexer.go` - レクサーの動作確認用
- `debug_parser.go` - パーサーの動作確認用

## 貢献

貢献を歓迎します！協力できる領域については `TODO.md` をご覧ください。

### 開発の進め方

1. プロジェクトをクローン
2. `go build ./cmd/zeno-compiler` でコンパイラをビルド
3. テストファイルで動作確認
4. 新機能の実装や改善を行う

### バグ報告

GitHub Issuesにてバグ報告や機能要求をお待ちしています。
