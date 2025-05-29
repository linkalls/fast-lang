# Zeno プログラミング言語 (Go実装)

Zenoは、GoとTypeScriptに触発された構文を持つ静的型付けプログラミング言語で、シンプルでありながら強力であることを目指して設計されています。このGoによるZenoコンパイラの実装は、現在ZenoコードをGoにコンパイルします。

## 特徴
- **Go/TypeScriptに触発された構文:** 可読性と現代的な開発プラクティスを目指します。
- **静的型付け:** 簡潔さのための型推論を備えています。
    - `let name = value;` (不変、型推論)
    - `let name: type = value;` (不変、明示的な型)
    - `mut name = value;` (可変、型推論)
    - `mut name: type = value;` (可変、明示的な型)
- **オプショナルなセミコロン:** 文末のセミコロンはオプションです。
- **基本型:** `int`, `float`, `bool`, `string`。
- **制御フロー:** `if/else if/else`, `loop`, `while`, `for`。
- **出力:** `print()` および `println()` 関数。
- **コメント:** `// 単一行コメント` および `/* 複数行コメント */`。
- **コンパイルターゲット:** Goコードを生成します。

## 現在の状況
- レキサー: 実装済み。
- パーサー: 実装済み、オプショナルなセミコロンをサポート。
- コードジェネレーター: 実装済み、ASTからGoコードを生成。
- コンパイラドライバー: 実装済み。
- プロジェクトはGo 1.21+を使用。

## 言語構文概要

**変数宣言:**
```zeno
// 不変、型推論
let message = "Hello, Zeno!"
let count = 100

// 可変、明示的な型
mut temperature: float = 25.5
mut is_active: bool = true

is_active = false
```

**制御フロー:**
```zeno
if count > 50 {
    println("Count is greater than 50")
} else if count == 50 {
    println("Count is exactly 50")
} else {
    println("Count is less than 50")
}

loop {
    println("Looping...")
    break // ループを終了
}

mut i = 0
while i < 3 {
    print(i)
    i = i + 1
} // 出力: 012

for let j = 0; j < 3; j = j + 1 {
    print(j)
} // 出力: 012
```

**出力:**
```zeno
print("これは一行で出力されます。 ")
println("これは新しい行に出力されます。")
let name = "Zeno"
println("Hello, " + name + "!") // 文字列の連結
```

## Zenoコンパイラの使用方法 (CLI)

Zenoコンパイラを使用するには、まずソースからビルドする必要があります。

### コンパイラのビルド
1.  Zenoプロジェクトのルートディレクトリ（`zeno-go/`ディレクトリ）に移動します。
2.  以下のGoコマンドを実行します：
    ```bash
    go build -o zeno ./cmd/zeno-compiler
    ```
3.  コンパイラの実行ファイルは`./zeno`に配置されます。

### コマンドラインインターフェース
基本的なコマンド構造は以下の通りです：
```bash
./zeno <SOURCE_FILE.zeno> [OPTIONS]
```

**一般的な操作と例:**

1.  **Zenoファイルをコンパイルして生成されたGoコードを表示:**
    これにより、変換されたGoコードを含む`output.go`（または`-o`が使用されていない場合はデフォルトで`<SOURCE_FILE>.go`）が作成されます。
    ```bash
    ./zeno examples/hello.zeno --output-go-file output.go --keep-go
    # またはデフォルトの .go 出力名を使用する場合（例：examples/hello.go）:
    ./zeno examples/hello.zeno --keep-go
    ```

2.  **Zenoファイルを直接実行ファイルにコンパイル:**
    これによりGoコードが生成され、`go build`を使用してコンパイルし、実行ファイル（例：`my_program`）を作成します。
    ```bash
    ./zeno examples/variables.zeno --compile --output-executable-file my_program
    ```
    `--output-executable-file`が省略された場合、実行ファイルはソースファイルと同じ名前（拡張子なし、例：`examples/variables`）になります。

3.  **Zenoファイルをコンパイルして即座に実行:**
    これは簡単なテストに便利です。中間Goファイルは`--keep-go`が指定されない限り、デフォルトで削除されます。
    ```bash
    ./zeno examples/controlflow.zeno --compile --run
    ```

**重要なフラグ:**
-   `<source_file>`: （必須）Zenoソースファイルへのパス（例：`examples/hello.zeno`）。
-   `--output-go-file <path>`, `-o <path>`: 生成されるGoコードの出力ファイルを指定します。
-   `--output-executable-file <path>`, `-O <path>`: コンパイルされた実行ファイルの出力ファイル名を指定します。
-   `--compile`, `-c`: 生成されたGoコードを`go build`を使用して実行ファイルにコンパイルします。
-   `--run`, `-r`: コンパイルされた実行ファイルを実行します。`--compile`が必要です。
-   `--keep-go`: コンパイル後の中間`.go`ファイルの削除を防ぎます。
-   `--help`: CLIの引数に関するヘルプ情報を表示します。

## 貢献
貢献を歓迎します！協力できる領域については`TODO.md`をご覧ください。
（さらなる貢献ガイドラインは後日追加予定）。
