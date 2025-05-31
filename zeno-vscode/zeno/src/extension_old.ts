import * as vscode from 'vscode';

// Zenoホバープロバイダー
class ZenoHoverProvider implements vscode.HoverProvider {
    provideHover(document: vscode.TextDocument, position: vscode.Position): vscode.ProviderResult<vscode.Hover> {
        const range = document.getWordRangeAtPosition(position);
        if (!range) {
            return undefined;
        }

        const word = document.getText(range);
        
        // 型情報を提供
        const typeInfo = this.getTypeInfo(word, document, position);
        if (typeInfo) {
            return new vscode.Hover(typeInfo, range);
        }

        return undefined;
    }

    private getTypeInfo(word: string, document: vscode.TextDocument, position: vscode.Position): string | undefined {
        const text = document.getText();
        
        // 変数の型推論
        const varPattern = new RegExp(`let\\s+${word}\\s*:\\s*(\\w+)`, 'g');
        const varMatch = varPattern.exec(text);
        if (varMatch) {
            return `**${word}**: ${varMatch[1]}`;
        }

        // 型推論なしの変数
        const inferPattern = new RegExp(`let\\s+${word}\\s*=\\s*(.+)`, 'g');
        const inferMatch = inferPattern.exec(text);
        if (inferMatch) {
            const value = inferMatch[1].trim();
            if (/^\d+$/.test(value)) {
                return `**${word}**: int (inferred)`;
            } else if (/^\d+\.\d+$/.test(value)) {
                return `**${word}**: float (inferred)`;
            } else if (/^".*"$/.test(value)) {
                return `**${word}**: string (inferred)`;
            } else if (/^(true|false)$/.test(value)) {
                return `**${word}**: bool (inferred)`;
            }
        }

        // 関数定義
        const funcPattern = new RegExp(`fn\\s+${word}\\s*\\(([^)]*)\\)\\s*(?:->\\s*(\\w+))?`, 'g');
        const funcMatch = funcPattern.exec(text);
        if (funcMatch) {
            const params = funcMatch[1] || '';
            const returnType = funcMatch[2] || 'void';
            return `**function ${word}**(${params}) -> ${returnType}`;
        }

        // キーワードの説明
        const keywords: { [key: string]: string } = {
            'fn': 'Keyword: Function definition',
            'let': 'Keyword: Variable declaration',
            'if': 'Keyword: Conditional statement',
            'else': 'Keyword: Alternative branch',
            'while': 'Keyword: Loop statement',
            'for': 'Keyword: Iteration statement',
            'return': 'Keyword: Return statement',
            'import': 'Keyword: Import statement',
            'pub': 'Keyword: Public visibility modifier',
            'from': 'Keyword: Import source',
            'in': 'Keyword: Iteration operator',
            'true': 'Boolean literal: true',
            'false': 'Boolean literal: false',
            'null': 'Null literal'
        };

        return keywords[word];
    }
}

// Zeno定義プロバイダー
class ZenoDefinitionProvider implements vscode.DefinitionProvider {
    provideDefinition(document: vscode.TextDocument, position: vscode.Position): vscode.ProviderResult<vscode.Definition> {
        const range = document.getWordRangeAtPosition(position);
        if (!range) {
            return undefined;
        }

        const word = document.getText(range);
        const text = document.getText();

        // 関数定義を検索
        const funcPattern = new RegExp(`fn\\s+${word}\\s*\\(`, 'g');
        const funcMatch = funcPattern.exec(text);
        if (funcMatch) {
            const pos = document.positionAt(funcMatch.index);
            return new vscode.Location(document.uri, pos);
        }

        // 変数定義を検索
        const varPattern = new RegExp(`let\\s+${word}\\b`, 'g');
        const varMatch = varPattern.exec(text);
        if (varMatch) {
            const pos = document.positionAt(varMatch.index);
            return new vscode.Location(document.uri, pos);
        }

        return undefined;
    }
}

// Zeno補完プロバイダー
class ZenoCompletionProvider implements vscode.CompletionItemProvider {
    provideCompletionItems(document: vscode.TextDocument, position: vscode.Position): vscode.ProviderResult<vscode.CompletionItem[]> {
        const completions: vscode.CompletionItem[] = [];

        // キーワード補完
        const keywords = [
            { label: 'fn', detail: 'Function definition', kind: vscode.CompletionItemKind.Keyword },
            { label: 'let', detail: 'Variable declaration', kind: vscode.CompletionItemKind.Keyword },
            { label: 'if', detail: 'Conditional statement', kind: vscode.CompletionItemKind.Keyword },
            { label: 'else', detail: 'Alternative branch', kind: vscode.CompletionItemKind.Keyword },
            { label: 'while', detail: 'Loop statement', kind: vscode.CompletionItemKind.Keyword },
            { label: 'for', detail: 'Iteration statement', kind: vscode.CompletionItemKind.Keyword },
            { label: 'return', detail: 'Return statement', kind: vscode.CompletionItemKind.Keyword },
            { label: 'import', detail: 'Import statement', kind: vscode.CompletionItemKind.Keyword },
            { label: 'pub', detail: 'Public visibility', kind: vscode.CompletionItemKind.Keyword },
            { label: 'from', detail: 'Import source', kind: vscode.CompletionItemKind.Keyword },
            { label: 'in', detail: 'Iteration operator', kind: vscode.CompletionItemKind.Keyword },
            { label: 'true', detail: 'Boolean true', kind: vscode.CompletionItemKind.Value },
            { label: 'false', detail: 'Boolean false', kind: vscode.CompletionItemKind.Value },
            { label: 'null', detail: 'Null value', kind: vscode.CompletionItemKind.Value }
        ];

        keywords.forEach(keyword => {
            const item = new vscode.CompletionItem(keyword.label, keyword.kind);
            item.detail = keyword.detail;
            item.insertText = keyword.label;
            completions.push(item);
        });

        // 型補完
        const types = [
            { label: 'int', detail: 'Integer type' },
            { label: 'string', detail: 'String type' },
            { label: 'bool', detail: 'Boolean type' },
            { label: 'float', detail: 'Floating point type' },
            { label: 'array', detail: 'Array type' },
            { label: 'map', detail: 'Map type' }
        ];

        types.forEach(type => {
            const item = new vscode.CompletionItem(type.label, vscode.CompletionItemKind.TypeParameter);
            item.detail = type.detail;
            item.insertText = type.label;
            completions.push(item);
        });

        // 標準ライブラリ関数
        const stdFunctions = [
            { label: 'print', detail: 'Print to console', snippet: 'print($1)' },
            { label: 'println', detail: 'Print line to console', snippet: 'println($1)' },
            { label: 'len', detail: 'Get length', snippet: 'len($1)' },
            { label: 'push', detail: 'Push to array', snippet: 'push($1, $2)' },
            { label: 'pop', detail: 'Pop from array', snippet: 'pop($1)' },
            { label: 'get', detail: 'Get from map', snippet: 'get($1, $2)' },
            { label: 'set', detail: 'Set in map', snippet: 'set($1, $2, $3)' }
        ];

        stdFunctions.forEach(func => {
            const item = new vscode.CompletionItem(func.label, vscode.CompletionItemKind.Function);
            item.detail = func.detail;
            item.insertText = new vscode.SnippetString(func.snippet);
            completions.push(item);
        });

        return completions;
    }
}

// 拡張機能のアクティベーション
export function activate(context: vscode.ExtensionContext) {
    console.log('Zeno Language Features extension is now active');

    // ホバープロバイダーを登録
    const hoverProvider = vscode.languages.registerHoverProvider('zeno', new ZenoHoverProvider());
    context.subscriptions.push(hoverProvider);

    // 定義プロバイダーを登録
    const definitionProvider = vscode.languages.registerDefinitionProvider('zeno', new ZenoDefinitionProvider());
    context.subscriptions.push(definitionProvider);

    // 補完プロバイダーを登録
    const completionProvider = vscode.languages.registerCompletionItemProvider('zeno', new ZenoCompletionProvider());
    context.subscriptions.push(completionProvider);

    console.log('Zeno extension activated with hover, definition, and completion providers');
}

export function deactivate() {
    console.log('Zeno Language Features extension is now deactivated');
}
