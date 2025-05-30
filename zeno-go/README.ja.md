# Zeno-Go コンパイラ

Zeno プログラミング言語から Go への変換を行うコンパイラです。Rust版からの移植作業の一環として開発されています。

## 特徴

- **TypeScript風のImport文**: `import {println} from "std/fmt"` のような構文をサポート
- **関数定義と呼び出し**: パラメータ、戻り値型、return文をサポート
- **未使用変数検出**: コンパイル時に未使用変数を検出してエラーを出力
- **Import検証**: 関数が適切にimportされているかをチェック
- **バイナリ式サポート**: 数学演算（+, -, *, /, %）と比較演算子をサポート
- **型注釈**: オプションの型注釈 `let x: int = 42`
- **多言語エラーメッセージ**: `-jp` フラグで日本語エラーメッセージも表示
- **変数宣言**: letキーワードによる変数宣言をサポート
- **組み込みリンター**: コード品質と規約のための静的解析。
- **浮動小数点リテラル**: 小数点を含む数値のサポート (例: `3.14`)。

## インストール

```bash
cd zeno-go
go build ./cmd/zeno
```

## 使用方法

### 基本的な使用方法

```bash
# Zenoファイルをコンパイルして実行
./zeno run example.zeno

# Zenoファイルをコンパイル（.goファイルを生成）
./zeno compile example.zeno

# 後方互換性：直接ファイル名を指定してコンパイル
./zeno example.zeno

# 日本語エラーメッセージも表示
./zeno -jp run example.zeno
```

## Zenoコードのリンティング

Zenoには、Zenoソースファイル内の潜在的な問題を特定し、コーディング規約を強制するのに役立つ組み込みリンターが含まれています。

### 使用方法

`lint` サブコマンドを使用してリンターを実行できます：

-   **単一ファイルをリントする:**
    ```bash
    ./zeno lint path/to/yourfile.zeno
    ```
-   **ディレクトリ内のすべての `.zeno` および `.zn` ファイルをリントする (再帰的):**
    ```bash
    ./zeno lint path/to/your_directory
    ```

リンターは、見つかった問題を以下の形式でコンソールに出力します：
`filepath:line:column: [rule-name] message`

リンティングの問題が見つかった場合、コマンドはステータスコード1で終了します。それ以外の場合は0で終了します。

### サポートされているルール (初期セット)

リンターは現在、以下をチェックします：

1.  **`unused-variable`**: `let` で宣言されたが使用されていない変数を検出します。(ルール L1)
2.  **`unused-function`**: 定義されているが使用されていない非公開関数 (`fn`) を検出します (`main` 関数を除く)。(ルール L2)
3.  **`function-naming-convention`**: 非公開関数 (`fn`) が `lowerCamelCase` であり、公開関数 (`pub fn`) が `UpperCamelCase` であることを保証します。(ルール L3)
4.  **`variable-naming-convention`**: `let` で宣言された変数が `lowerCamelCase` であることを保証します (`_` 識別子を無視)。(ルール L4)
5.  **`unused-import`**: モジュールからインポートされたが現在のファイルで使用されていないシンボルを検出します。(ルール L5)

*(注意: イシューレポート内の行番号と列番号は現在プレースホルダー (0:0) であり、将来的にパーサーがASTノードに位置情報を含むように強化されることで改善されます。)*

*(将来の機能強化には、有効なルールとそのパラメータをカスタマイズするための設定ファイルが含まれる可能性があります。)*

### Zeno言語の例

#### 基本的な例
```zeno
import { println } from "std/fmt"

let x = 10
let y = 20
let result = x + y
println(result)
```

#### 関数の例
```zeno
import {println} from "std/fmt"

fn add(a: int, b: int): int {
    return a + b
}

fn greet(name: string) {
    println("Hello, " + name + "!")
}

fn main() {
    let result: int = add(5, 3)
    println("Result: ", result)
    
    greet("Zeno")
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
import {println, print} from "std/fmt"
```

### 変数宣言
```zeno
let x = 42           // 変数宣言
let y: int = 100     // 型注釈付き
let pi = 3.14        // 浮動小数点数
```

### 関数定義
```zeno
fn add(a: int, b: int): int {
    return a + b
}

import { println } from "std/fmt" // この例のためにprintlnがインポートされていると仮定
fn greet(name: string) {
    println("Hello, " + name + "!")
}
```

### 関数呼び出し
```zeno
let result = add(10, 20)
greet("World")
```

### main関数
```zeno
import { println } from "std/fmt" // printlnがインポートされていると仮定

fn main() {
    // プログラムのエントリーポイント
    println("Hello, World!")
}
```

### バイナリ式
```zeno
let sum = 10 + 20
let product = 5 * 6
let comparison = x > y
```

### コンソールへの出力 (std/fmt を使用)
出力処理は `std/fmt` モジュールの関数によって行われます。使用前にインポートする必要があります。
```zeno
print("Hello")       // import {print} from "std/fmt" が必要
println("World")     // import {println} from "std/fmt" が必要
```

## エラー検出機能

### 未使用変数の検出
```zeno
import {println} from "std/fmt"

let x = 10
let unused = 42  // エラー: Unused variables found: unused
let y = x + 5
println(y)
```

### Import検証
```zeno
// エラー: println is not imported from std/fmt
let x = 10
println(x)  // import文がない場合はエラー
```

## 標準ライブラリ

現在サポートされているモジュール:

- `std/fmt`: `print`, `println` 関数
- `std/io`: `readFile`, `writeFile`, `remove`, `pwd` 関数
- `std/json`: JSONパース (`parse`) 及び文字列化 (`stringify`) 関数

### std/io モジュールの使用法

`std/io` モジュールは、シンプルで直感的なファイルI/O操作を提供します：

```zeno
import { println } from "std/fmt"
import { readFile, writeFile } from "std/io"

fn main() {
    // ファイルにコンテンツを書き込み
    let content = "こんにちは、Zeno!\nテストファイルです。"
    writeFile("example.txt", content)
    println("ファイルが正常に書き込まれました！")
    
    // ファイルからコンテンツを読み込み
    let fileContent = readFile("example.txt")
    println("ファイルの内容:")
    println(fileContent)
    
    // 構造化データの書き込み
    let jsonData = "{\"name\": \"Zeno\", \"version\": \"1.0\"}"
    writeFile("config.json", jsonData)
    
    let configData = readFile("config.json")
    println("設定: ", configData)
}
```

#### std/io 関数

- `writeFile(filename: string, content: string)`: 自動エラーハンドリング付きでファイルにコンテンツを書き込み
- `readFile(filename: string): string`: ファイル内容を読み込んで文字列として返却、エラー時は空文字列を返却
- `remove(filename: string): bool`: 指定されたファイルまたは空のディレクトリを削除します。成功時に `true`、失敗時に `false` を返します。
- `pwd(): string`: 現在の作業ディレクトリを絶対パスとして返します。失敗時には空文字列を返します。

### std/json モジュールの使用法

`std/json` モジュールは、JSON文字列をZenoのデータ構造にパースする機能と、Zenoのデータ構造をJSON文字列に変換する機能を提供します。

```zeno
import { println, print } from "std/fmt"
import { parse, stringify } from "std/json"

fn main() {
    let jsonString = "{\"name\": \"Zeno\", \"version\": 0.2, \"active\": true}"
    println("元のJSON文字列: " + jsonString)

    let parsedData = parse(jsonString)
    // 現状、'parsedData' は 'any' 型です。その構造（マップのキーアクセスや配列要素アクセスなど）を
    // 直接操作するには、将来のZeno言語の型検査や 'any' 型操作機能に依存します。

    let reStringified = stringify(parsedData)
    print("再文字列化されたJSON: ")
    println(reStringified)

    let zenoData = "単純な文字列" // Zenoのプリミティブ値の例
    let jsonFromZeno = stringify(zenoData)
    print("Zeno文字列 '単純な文字列' からのJSON: ")
    println(jsonFromZeno) // 期待値: "\"単純な文字列\""
    
    let invalidJson = "{\"key\": value_not_string}" // 注意: この行が有効なZenoの行であるためには、value_not_stringがZeno文字列である必要があります
    let parsedError = parse(invalidJson)
    print("不正なJSONをパースした結果: ")
    println(stringify(parsedError)) // 期待値: "null"
}
```

#### std/json 関数

- `parse(jsonString: string): any`: JSON文字列をパースします。パースされたデータを `any` 型（Zenoの文字列、数値、ブール値、リスト、またはマップを表す）として返します。パースエラーの場合はZenoの `nil` 相当（JSONの `null` に文字列化される値）を返します。
- `stringify(value: any): string`: Zenoのデータ（`any`型で、プリミティブ、リスト、またはマップで構成されることを期待）をJSON文字列に変換します。文字列化エラーの場合は空文字列 `""` を返します。

## 実装されている機能

✅ **完了済み:**
- Import文の解析と検証
- 変数宣言（let）
- バイナリ式（算術演算、比較演算）
- [x] Print関数呼び出しへの変換: print/printlnはキーワードから`std/fmt`モジュールの関数に変更
- 未使用変数検出
- 多言語エラーメッセージ（英語/日本語）
- トークン解析（Lexer）
- AST構築（Parser）
- Goコード生成（Generator）
- 標準ライブラリ: std/json モジュール (parse, stringify)
- 浮動小数点リテラルのパースと生成

🔲 **今後の予定:**
- 関数定義と呼び出し
- 制御フロー（if/else、while、loop）
- 型システムの拡張
- 標準ライブラリの拡充

## コンパイラの使用方法

### コンパイラのビルド

```bash
cd zeno-go
go build ./cmd/zeno
```

### 基本的な使用方法

```bash
# Zenoファイルをコンパイルして実行
./zeno run example.zeno

# Zenoファイルをコンパイル（.goファイルを生成）
./zeno compile example.zeno

# 日本語エラーメッセージも表示
./zeno -jp run example.zeno
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
2. `go build ./cmd/zeno` でコンパイラをビルド
3. テストファイルで動作確認
4. 新機能の実装や改善を行う

### バグ報告

GitHub Issuesにてバグ報告や機能要求をお待ちしています。
