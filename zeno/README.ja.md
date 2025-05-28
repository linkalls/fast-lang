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

## ビルド方法 (プレースホルダー)
詳細なビルド手順は、コンパイラドライバー（`src/main.rs`）が完成次第追加されます。
コンパイラプロジェクト自体（Rustプロジェクト）をビルドするには：
```bash
# (zenoディレクトリにいない場合は移動してください)
cargo build
```

## コントリビューション
コントリビューションを歓迎します！貢献できる分野については `TODO.md` を参照してください。
（さらなるコントリビューションガイドラインは後日追加予定です）。
```
