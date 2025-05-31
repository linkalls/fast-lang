エラーが発生していますが、`libsecret-1`ライブラリが不足しているのが原因です。これを解決してからVS Code拡張機能をテストしましょう。

## 1. 依存関係をインストール

````bash
# libsecret-1-devをインストール
sudo apt update
sudo apt install libsecret-1-dev

# vsceを再インストール
bun install -g vsce
````

## 2. VS Code拡張機能のテスト手順

### 方法1: 開発モードでテスト（推奨）

````bash
# zeno-vscodeディレクトリに移動
cd ~/zeno-lang/zeno-vscode

# VS Codeで開く
code .
````

VS Codeで開いたら：
1. **F5キー**を押すか、**Run > Start Debugging**を選択
2. 新しいVS Codeウィンドウ（Extension Development Host）が開きます
3. 新しいウィンドウでテストファイルを作成：

````zeno
func main() {
    print("Hello Zeno!")
    let x = 42
    if x > 0 {
        print("Positive number")
    }
}
````

### 方法2: パッケージ化してテスト

````bash
# zeno-vscodeディレクトリで実行
cd ~/zeno-lang/zeno-vscode

cd /home/poteto/zeno-lang/zeno-vscode && bun run compile

 cd /home/poteto/zeno-lang/zeno-vscode && rm -f zeno-language-features-0.0.1.vsix && vsce package --no-dependencies

 cd /home/poteto/zeno-lang/zeno-vscode && code --uninstall-extension ZenoLang.zeno-language-features && code --install-extension zeno-language-features-0.0.1.vsix

# 拡張機能をパッケージ化
vsce package

# 生成されたvsixファイルをインストール
code --install-extension zeno-language-features-0.0.1.vsix
````

## 3. 確認すべき機能

- `.zn`や`.zeno`ファイルでのシンタックスハイライト
- コメント機能（や`/* */`）
- 括弧の自動補完
- スニペットの動作

拡張機能が正常に動作しているかテストしてみてください！