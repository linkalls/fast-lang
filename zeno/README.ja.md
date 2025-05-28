# Zeno プログラミング言語

Zenoは、GoとTypeScriptに触発された構文を持つ静的型付けプログラミング言語で、シンプルでありながら強力であることを目指して設計されています。ZenoのコンパイラはRust（Edition 2024）で実装されており、現在はZenoコードをRustにコンパイルします。

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
- **コンパイルターゲット:** Rustコードを生成します。

## 現在の状況
- レキサー: 実装済み。
- パーサー: 実装済み、オプショナルなセミコロンをサポート。
- コードジェネレーター: 実装済み、ASTからRustコードを生成。
- コンパイラドライバー: 開発中。
- プロジェクトはRust Edition 2024を使用。

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
    break // ループを抜ける
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
print("This prints on one line. ")
println("This prints on a new line.")
let name = "Zeno"
println("Hello, " + name + "!") // 文字列結合（+が文字列に対して機能すると仮定）
```

## Zenoコンパイラの使用 (CLI)

Zenoコンパイラを使用するには、まずソースからビルドする必要があります。

### コンパイラのビルド
1.  Zenoプロジェクトのルートディレクトリ（`zeno/` ディレクトリ）に移動します。
2.  次のCargoコマンドを実行します:
    ```bash
    cargo build --release
    ```
3.  コンパイラの実行可能ファイルは `target/release/zeno`（Windowsの場合は `target\release\zeno.exe`）に配置されます。

### コマンドラインインターフェース
基本的なコマンド構造は次のとおりです:
```bash
./target/release/zeno <SOURCE_FILE.zeno> [OPTIONS]
```

**一般的な操作と例:**

1.  **Zenoファイルをコンパイルして生成されたRustコードを表示する:**
    これにより、翻訳されたRustコードが含まれる `output.rs`（`-o` が使用されない場合はデフォルトで `<SOURCE_FILE>.rs`）が作成されます。
    ```bash
    ./target/release/zeno examples/hello.zeno --output-rust-file output.rs --keep-rs
    # またはデフォルトの .rs 出力名を使用する場合 (例: examples/hello.rs):
    ./target/release/zeno examples/hello.zeno --keep-rs
    ```

2.  **Zenoファイルを直接実行可能ファイルにコンパイルする:**
    これによりRustコードが生成され、`rustc` を使用してコンパイルされ、実行可能ファイル（例: `my_program`）が作成されます。
    ```bash
    ./target/release/zeno examples/variables.zeno --compile --output-executable-file my_program
    ```
    `--output-executable-file` が省略された場合、実行可能ファイルはソースファイルと同じ名前になります（拡張子なし、例: `examples/variables`）。

3.  **Zenoファイルをコンパイルして即座に実行する:**
    これは簡単なテストに便利です。中間的なRustファイルは、`--keep-rs` が指定されない限り、デフォルトで削除されます。
    ```bash
    ./target/release/zeno examples/controlflow.zeno --compile --run
    ```

**重要なフラグ:**
-   `<source_file>`: (必須) Zenoソースファイルへのパス (例: `examples/hello.zeno`)。
-   `--output-rust-file <path>`, `-o <path>`: 生成されたRustコードの出力ファイルを指定します。
-   `--output-executable-file <path>`, `-O <path>`: コンパイルされた実行可能ファイルの出力ファイル名を指定します。
-   `--compile`, `-c`: 生成されたRustコードを `rustc` を使用して実行可能ファイルにコンパイルします。
-   `--run`, `-r`: コンパイルされた実行可能ファイルを実行します。`--compile` が必要です。
-   `--keep-rs`: コンパイル後に中間的な `.rs` ファイルの削除を防ぎます。
-   `--help`: CLI引数に関するヘルプ情報を表示します。

## コントリビューション
コントリビューションを歓迎します！貢献できる分野については `TODO.md` を参照してください。
（さらなるコントリビューションガイドラインは後日追加予定です）。
```
