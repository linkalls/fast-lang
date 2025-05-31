"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.activate = activate;
exports.deactivate = deactivate;
const vscode = __importStar(require("vscode"));
// Zeno言語のTokenタイプを定義
const tokenTypes = new Map();
const tokenModifiers = new Map();
// Legend（凡例）を作成
const legend = (function () {
    const tokenTypesLegend = [
        'comment', 'string', 'keyword', 'number', 'regexp', 'operator', 'namespace',
        'type', 'struct', 'class', 'interface', 'enum', 'typeParameter', 'function',
        'method', 'decorator', 'macro', 'variable', 'parameter', 'property', 'label'
    ];
    tokenTypesLegend.forEach((tokenType, index) => tokenTypes.set(tokenType, index));
    const tokenModifiersLegend = [
        'declaration', 'definition', 'readonly', 'static', 'deprecated', 'abstract',
        'async', 'modification', 'documentation', 'defaultLibrary'
    ];
    tokenModifiersLegend.forEach((tokenModifier, index) => tokenModifiers.set(tokenModifier, index));
    return new vscode.SemanticTokensLegend(tokenTypesLegend, tokenModifiersLegend);
})();
function activate(context) {
    console.log('Zeno Language Features extension is now active!');
    // Semantic Tokens Providerを登録
    const provider = vscode.languages.registerDocumentSemanticTokensProvider({ language: 'zeno' }, new ZenoSemanticTokensProvider(), legend);
    // ホバープロバイダーを登録
    const hoverProvider = vscode.languages.registerHoverProvider('zeno', {
        provideHover(document, position, token) {
            const range = document.getWordRangeAtPosition(position);
            if (!range)
                return undefined;
            const word = document.getText(range);
            const typeInfo = getTypeInfo(word, document, position);
            if (typeInfo) {
                return new vscode.Hover(new vscode.MarkdownString(typeInfo));
            }
        }
    });
    // 定義プロバイダーを登録
    const definitionProvider = vscode.languages.registerDefinitionProvider('zeno', {
        provideDefinition(document, position, token) {
            const range = document.getWordRangeAtPosition(position);
            if (!range)
                return undefined;
            const word = document.getText(range);
            return findDefinition(word, document);
        }
    });
    // 補完プロバイダーを登録
    const completionProvider = vscode.languages.registerCompletionItemProvider('zeno', {
        provideCompletionItems(document, position, token, context) {
            return getCompletionItems(document, position);
        }
    }, '.', ':');
    context.subscriptions.push(provider, hoverProvider, definitionProvider, completionProvider);
}
function deactivate() {
    return undefined;
}
// Semantic Tokens Provider クラス
class ZenoSemanticTokensProvider {
    async provideDocumentSemanticTokens(document, token) {
        const allTokens = this._parseText(document.getText());
        const builder = new vscode.SemanticTokensBuilder(legend);
        allTokens.forEach((token) => {
            builder.push(token.line, token.startCharacter, token.length, this._encodeTokenType(token.tokenType), this._encodeTokenModifiers(token.tokenModifiers));
        });
        return builder.build();
    }
    _encodeTokenType(tokenType) {
        if (tokenTypes.has(tokenType)) {
            return tokenTypes.get(tokenType);
        }
        else if (tokenType === 'notInLegend') {
            return tokenTypes.size + 2;
        }
        return 0;
    }
    _encodeTokenModifiers(strTokenModifiers) {
        let result = 0;
        for (let i = 0; i < strTokenModifiers.length; i++) {
            const tokenModifier = strTokenModifiers[i];
            if (tokenModifiers.has(tokenModifier)) {
                result = result | (1 << tokenModifiers.get(tokenModifier));
            }
            else if (tokenModifier === 'notInLegend') {
                result = result | (1 << tokenModifiers.size + 2);
            }
        }
        return result;
    }
    _parseText(text) {
        const r = [];
        const lines = text.split(/\r\n|\r|\n/);
        for (let i = 0; i < lines.length; i++) {
            const line = lines[i];
            let currentOffset = 0;
            // コメントをパース
            const commentMatch = line.match(/\/\/.*$/);
            if (commentMatch) {
                const start = line.indexOf(commentMatch[0]);
                r.push({
                    line: i,
                    startCharacter: start,
                    length: commentMatch[0].length,
                    tokenType: 'comment',
                    tokenModifiers: []
                });
            }
            // 文字列リテラルをパース
            const stringRegex = /"[^"]*"/g;
            let stringMatch;
            while ((stringMatch = stringRegex.exec(line)) !== null) {
                r.push({
                    line: i,
                    startCharacter: stringMatch.index,
                    length: stringMatch[0].length,
                    tokenType: 'string',
                    tokenModifiers: []
                });
            }
            // 数値をパース
            const numberRegex = /\b\d+\b/g;
            let numberMatch;
            while ((numberMatch = numberRegex.exec(line)) !== null) {
                r.push({
                    line: i,
                    startCharacter: numberMatch.index,
                    length: numberMatch[0].length,
                    tokenType: 'number',
                    tokenModifiers: []
                });
            }
            // キーワードをパース
            const keywords = ['fn', 'let', 'if', 'else', 'while', 'for', 'return', 'import', 'pub'];
            keywords.forEach(keyword => {
                const keywordRegex = new RegExp(`\\b${keyword}\\b`, 'g');
                let keywordMatch;
                while ((keywordMatch = keywordRegex.exec(line)) !== null) {
                    r.push({
                        line: i,
                        startCharacter: keywordMatch.index,
                        length: keyword.length,
                        tokenType: 'keyword',
                        tokenModifiers: []
                    });
                }
            });
            // 型をパース
            const types = ['int', 'string', 'bool', 'float', 'array'];
            types.forEach(type => {
                const typeRegex = new RegExp(`\\b${type}\\b`, 'g');
                let typeMatch;
                while ((typeMatch = typeRegex.exec(line)) !== null) {
                    r.push({
                        line: i,
                        startCharacter: typeMatch.index,
                        length: type.length,
                        tokenType: 'type',
                        tokenModifiers: []
                    });
                }
            });
            // 関数定義をパース
            const functionDefRegex = /fn\s+(\w+)\s*\(/g;
            let funcMatch;
            while ((funcMatch = functionDefRegex.exec(line)) !== null) {
                const funcName = funcMatch[1];
                const funcStart = funcMatch.index + funcMatch[0].indexOf(funcName);
                r.push({
                    line: i,
                    startCharacter: funcStart,
                    length: funcName.length,
                    tokenType: 'function',
                    tokenModifiers: ['declaration']
                });
            }
            // 変数定義をパース
            const varDefRegex = /let\s+(\w+)/g;
            let varMatch;
            while ((varMatch = varDefRegex.exec(line)) !== null) {
                const varName = varMatch[1];
                const varStart = varMatch.index + varMatch[0].indexOf(varName);
                r.push({
                    line: i,
                    startCharacter: varStart,
                    length: varName.length,
                    tokenType: 'variable',
                    tokenModifiers: ['declaration']
                });
            }
            // 関数呼び出しをパース
            const funcCallRegex = /(\w+)\s*\(/g;
            let callMatch;
            while ((callMatch = funcCallRegex.exec(line)) !== null) {
                const funcName = callMatch[1];
                // キーワードや変数定義は除外
                if (!keywords.includes(funcName) && !line.substring(0, callMatch.index).includes('let')) {
                    r.push({
                        line: i,
                        startCharacter: callMatch.index,
                        length: funcName.length,
                        tokenType: 'function',
                        tokenModifiers: []
                    });
                }
            }
        }
        return r;
    }
}
// 型情報を取得する関数
function getTypeInfo(word, document, position) {
    const text = document.getText();
    // 変数宣言パターンをマッチ（明示的な型注釈）
    const letPattern = new RegExp(`let\\s+${word}\\s*:\\s*([\\w\\[\\]]+)`, 'g');
    const letMatch = letPattern.exec(text);
    if (letMatch) {
        return `**${word}**: \`${letMatch[1]}\``;
    }
    // 関数パラメータパターンをマッチ
    const paramPattern = new RegExp(`${word}\\s*:\\s*([\\w\\[\\]]+)`, 'g');
    const paramMatch = paramPattern.exec(text);
    if (paramMatch) {
        return `**${word}**: \`${paramMatch[1]}\``;
    }
    // 関数定義パターンをマッチ
    const funcPattern = new RegExp(`fn\\s+${word}\\s*\\([^)]*\\)\\s*:\\s*([\\w\\[\\]]+)`, 'g');
    const funcMatch = funcPattern.exec(text);
    if (funcMatch) {
        return `**function ${word}**: returns \`${funcMatch[1]}\``;
    }
    // 型推論による簡単な判定
    const assignPattern = new RegExp(`let\\s+${word}\\s*=\\s*([^\\n;]+)`, 'g');
    const assignMatch = assignPattern.exec(text);
    if (assignMatch) {
        const value = assignMatch[1].trim();
        if (/^\d+$/.test(value)) {
            return `**${word}**: \`int\` (inferred)`;
        }
        if (/^".*"$/.test(value) || /^'.*'$/.test(value)) {
            return `**${word}**: \`string\` (inferred)`;
        }
        if (value === 'true' || value === 'false') {
            return `**${word}**: \`bool\` (inferred)`;
        }
        // 関数呼び出しの場合
        const funcCallMatch = value.match(/^(\w+)\s*\(/);
        if (funcCallMatch) {
            const funcName = funcCallMatch[1];
            const funcReturnType = getFunctionReturnType(funcName, text);
            if (funcReturnType) {
                return `**${word}**: \`${funcReturnType}\` (from ${funcName}())`;
            }
        }
    }
    return undefined;
}
// 関数の戻り値型を取得
function getFunctionReturnType(funcName, text) {
    const funcPattern = new RegExp(`fn\\s+${funcName}\\s*\\([^)]*\\)\\s*:\\s*([\\w\\[\\]]+)`, 'g');
    const match = funcPattern.exec(text);
    return match ? match[1] : undefined;
}
// 定義を検索する関数
function findDefinition(word, document) {
    const text = document.getText();
    const lines = text.split('\n');
    const locations = [];
    for (let i = 0; i < lines.length; i++) {
        const line = lines[i];
        // 関数定義をチェック
        const funcRegex = new RegExp(`fn\\s+${word}\\s*\\(`);
        if (funcRegex.test(line)) {
            const position = new vscode.Position(i, line.indexOf(word));
            locations.push(new vscode.Location(document.uri, position));
        }
        // 変数定義をチェック
        const letRegex = new RegExp(`let\\s+${word}\\s*[=:]`);
        if (letRegex.test(line)) {
            const position = new vscode.Position(i, line.indexOf(word));
            locations.push(new vscode.Location(document.uri, position));
        }
    }
    return locations;
}
// 補完アイテムを提供する関数
function getCompletionItems(document, position) {
    const items = [];
    // Zeno言語のキーワード
    const keywords = ['fn', 'let', 'if', 'else', 'while', 'for', 'return', 'import', 'pub'];
    keywords.forEach(keyword => {
        const item = new vscode.CompletionItem(keyword, vscode.CompletionItemKind.Keyword);
        item.insertText = keyword;
        item.detail = `Zeno keyword: ${keyword}`;
        items.push(item);
    });
    // 型の補完
    const types = ['int', 'string', 'bool', 'float', 'array'];
    types.forEach(type => {
        const item = new vscode.CompletionItem(type, vscode.CompletionItemKind.TypeParameter);
        item.insertText = type;
        item.detail = `Zeno type: ${type}`;
        items.push(item);
    });
    // 標準ライブラリ関数
    const stdFunctions = [
        { name: 'println', detail: 'println(message: string)', doc: 'Print a line to stdout' },
        { name: 'print', detail: 'print(message: string)', doc: 'Print to stdout' }
    ];
    stdFunctions.forEach(func => {
        const item = new vscode.CompletionItem(func.name, vscode.CompletionItemKind.Function);
        item.detail = func.detail;
        item.documentation = func.doc;
        item.insertText = new vscode.SnippetString(`${func.name}($1)`);
        items.push(item);
    });
    return items;
}
//# sourceMappingURL=extension_semantic.js.map