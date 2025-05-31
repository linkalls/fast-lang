"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const node_1 = require("vscode-languageserver/node");
const vscode_languageserver_textdocument_1 = require("vscode-languageserver-textdocument");
// Connection を作成
const connection = (0, node_1.createConnection)(node_1.ProposedFeatures.all);
// テキストドキュメントマネージャーを作成
const documents = new node_1.TextDocuments(vscode_languageserver_textdocument_1.TextDocument);
let hasConfigurationCapability = false;
let hasWorkspaceFolderCapability = false;
let hasDiagnosticRelatedInformationCapability = false;
connection.onInitialize((params) => {
    const capabilities = params.capabilities;
    hasConfigurationCapability = !!(capabilities.workspace && !!capabilities.workspace.configuration);
    hasWorkspaceFolderCapability = !!(capabilities.workspace && !!capabilities.workspace.workspaceFolders);
    hasDiagnosticRelatedInformationCapability = !!(capabilities.textDocument &&
        capabilities.textDocument.publishDiagnostics &&
        capabilities.textDocument.publishDiagnostics.relatedInformation);
    const result = {
        capabilities: {
            textDocumentSync: node_1.TextDocumentSyncKind.Incremental,
            completionProvider: {
                resolveProvider: true,
                triggerCharacters: ['.', ':']
            },
            hoverProvider: true,
            definitionProvider: true
        }
    };
    if (hasWorkspaceFolderCapability) {
        result.capabilities.workspace = {
            workspaceFolders: {
                supported: true
            }
        };
    }
    return result;
});
connection.onInitialized(() => {
    if (hasConfigurationCapability) {
        connection.client.register(node_1.DidChangeConfigurationNotification.type, undefined);
    }
    if (hasWorkspaceFolderCapability) {
        connection.workspace.onDidChangeWorkspaceFolders(_event => {
            connection.console.log('Workspace folder change event received.');
        });
    }
});
const defaultSettings = { maxNumberOfProblems: 1000 };
let globalSettings = defaultSettings;
let documentSettings = new Map();
connection.onDidChangeConfiguration(change => {
    if (hasConfigurationCapability) {
        documentSettings.clear();
    }
    else {
        globalSettings = ((change.settings.zenoLanguageServer || defaultSettings));
    }
    documents.all().forEach(validateTextDocument);
});
function getDocumentSettings(resource) {
    if (!hasConfigurationCapability) {
        return Promise.resolve(globalSettings);
    }
    let result = documentSettings.get(resource);
    if (!result) {
        result = connection.workspace.getConfiguration({
            scopeUri: resource,
            section: 'zenoLanguageServer'
        });
        documentSettings.set(resource, result);
    }
    return result;
}
documents.onDidClose(e => {
    documentSettings.delete(e.document.uri);
});
documents.onDidChangeContent(change => {
    validateTextDocument(change.document);
});
async function validateTextDocument(textDocument) {
    const settings = await getDocumentSettings(textDocument.uri);
    const text = textDocument.getText();
    const pattern = /\b[A-Z]{2,}\b/g;
    let m;
    let problems = 0;
    const diagnostics = [];
    while ((m = pattern.exec(text)) && problems < settings.maxNumberOfProblems) {
        problems++;
        const diagnostic = {
            severity: node_1.DiagnosticSeverity.Warning,
            range: {
                start: textDocument.positionAt(m.index),
                end: textDocument.positionAt(m.index + m[0].length)
            },
            message: `${m[0]} is all uppercase.`,
            source: 'zeno'
        };
        if (hasDiagnosticRelatedInformationCapability) {
            diagnostic.relatedInformation = [
                {
                    location: {
                        uri: textDocument.uri,
                        range: Object.assign({}, diagnostic.range)
                    },
                    message: 'Spelling matters'
                }
            ];
        }
        diagnostics.push(diagnostic);
    }
    connection.sendDiagnostics({ uri: textDocument.uri, diagnostics });
}
// ホバー機能
connection.onHover((params) => {
    const document = documents.get(params.textDocument.uri);
    if (!document) {
        return undefined;
    }
    const position = params.position;
    const text = document.getText();
    const offset = document.offsetAt(position);
    // 現在位置の単語を取得
    const word = getWordAtPosition(text, offset);
    if (!word) {
        return undefined;
    }
    // 型情報を取得
    const typeInfo = getTypeInfo(word, text);
    if (typeInfo) {
        return {
            contents: {
                kind: node_1.MarkupKind.Markdown,
                value: typeInfo
            }
        };
    }
    return undefined;
});
// 補完機能
connection.onCompletion((_textDocumentPosition) => {
    return [
        {
            label: 'fn',
            kind: node_1.CompletionItemKind.Keyword,
            data: 1
        },
        {
            label: 'let',
            kind: node_1.CompletionItemKind.Keyword,
            data: 2
        },
        {
            label: 'if',
            kind: node_1.CompletionItemKind.Keyword,
            data: 3
        },
        {
            label: 'else',
            kind: node_1.CompletionItemKind.Keyword,
            data: 4
        },
        {
            label: 'println',
            kind: node_1.CompletionItemKind.Function,
            data: 5
        },
        {
            label: 'int',
            kind: node_1.CompletionItemKind.TypeParameter,
            data: 6
        },
        {
            label: 'string',
            kind: node_1.CompletionItemKind.TypeParameter,
            data: 7
        }
    ];
});
connection.onCompletionResolve((item) => {
    if (item.data === 1) {
        item.detail = 'Function declaration';
        item.documentation = 'Define a new function';
    }
    else if (item.data === 2) {
        item.detail = 'Variable declaration';
        item.documentation = 'Declare a new variable';
    }
    else if (item.data === 5) {
        item.detail = 'println(message: string)';
        item.documentation = 'Print a line to stdout';
    }
    return item;
});
// ヘルパー関数
function getWordAtPosition(text, offset) {
    const before = text.slice(0, offset);
    const after = text.slice(offset);
    const wordPattern = /\w+/;
    const beforeMatch = before.match(/\w+$/);
    const afterMatch = after.match(/^\w+/);
    const beforePart = beforeMatch ? beforeMatch[0] : '';
    const afterPart = afterMatch ? afterMatch[0] : '';
    return beforePart + afterPart || undefined;
}
function getTypeInfo(word, text) {
    // 変数宣言パターン
    const letPattern = new RegExp(`let\\s+${word}\\s*:\\s*([\\w\\[\\]]+)`, 'g');
    const letMatch = letPattern.exec(text);
    if (letMatch) {
        return `**${word}**: \`${letMatch[1]}\``;
    }
    // 関数定義パターン
    const funcPattern = new RegExp(`fn\\s+${word}\\s*\\([^)]*\\)\\s*:\\s*([\\w\\[\\]]+)`, 'g');
    const funcMatch = funcPattern.exec(text);
    if (funcMatch) {
        return `**function ${word}**: returns \`${funcMatch[1]}\``;
    }
    // 型推論
    const assignPattern = new RegExp(`let\\s+${word}\\s*=\\s*([^\\n;]+)`, 'g');
    const assignMatch = assignPattern.exec(text);
    if (assignMatch) {
        const value = assignMatch[1].trim();
        if (/^\d+$/.test(value)) {
            return `**${word}**: \`int\` (inferred)`;
        }
        if (/^".*"$/.test(value)) {
            return `**${word}**: \`string\` (inferred)`;
        }
        if (value === 'true' || value === 'false') {
            return `**${word}**: \`bool\` (inferred)`;
        }
    }
    return undefined;
}
// ドキュメントマネージャーをlisten
documents.listen(connection);
// コネクションをlisten
connection.listen();
//# sourceMappingURL=server.js.map