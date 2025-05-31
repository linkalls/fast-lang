import * as vscode from 'vscode';

// JSDocスタイルのドキュメント情報を格納する型
interface JSDocInfo {
  description: string;
  params: Array<{
    name: string;
    type?: string;
    description: string;
  }>;
  returns?: string;
  example?: string;
  since?: string;
  deprecated?: string;
  throws?: Array<{
    type: string;
    description: string;
  }>;
  author?: string;
  version?: string;
}

export function activate(context: vscode.ExtensionContext) {
  console.log('Zeno Language Features extension is now active!');

  // JSDocスタイルのホバープロバイダーを登録
  const hoverProvider = vscode.languages.registerHoverProvider('zeno', {
    provideHover(document, position, token) {
      const range = document.getWordRangeAtPosition(position);
      if (!range) return undefined;

      const word = document.getText(range);
      
      // TypeScript風の美しい型情報を提供
      const hoverInfo = getEnhancedTypeInfo(word, document, position);
      if (hoverInfo) {
        const markdown = new vscode.MarkdownString(hoverInfo.content);
        markdown.isTrusted = true;
        markdown.supportHtml = true;
        return new vscode.Hover(markdown, range);
      }
    }
  });

  // 定義プロバイダーを登録
  const definitionProvider = vscode.languages.registerDefinitionProvider('zeno', {
    provideDefinition(document, position, token) {
      const range = document.getWordRangeAtPosition(position);
      if (!range) return undefined;

      const word = document.getText(range);

      // 関数定義を検索
      return findDefinition(word, document);
    }
  });

  // 補完プロバイダーを登録
  const completionProvider = vscode.languages.registerCompletionItemProvider('zeno', {
    provideCompletionItems(document, position, token, context) {
      return getCompletionItems(document, position);
    }
  }, '.', ':');

  context.subscriptions.push(hoverProvider, definitionProvider, completionProvider);
}

export function deactivate(): Thenable<void> | undefined {
  return undefined;
}

// JSDocスタイルのコメントを解析する関数
function findJSDocComment(functionName: string, lines: string[]): JSDocInfo | null {
  // 関数定義を探す
  let functionLineIndex = -1;
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim();
    const funcPattern = new RegExp(`(pub\\s+)?fn\\s+${functionName}\\s*\\(`);
    if (funcPattern.test(line)) {
      functionLineIndex = i;
      break;
    }
  }

  if (functionLineIndex === -1) {
    return null;
  }

  // 関数定義の前のコメントを探す
  const commentLines: string[] = [];
  let commentStartIndex = -1;

  // 関数定義の直前から上に向かってコメントを探す
  for (let i = functionLineIndex - 1; i >= 0; i--) {
    const line = lines[i].trim();
    
    if (line === '') {
      // 空行は許可
      continue;
    }
    
    if (line.startsWith('/**') || line.startsWith('/*')) {
      // JSDocコメントの開始
      commentStartIndex = i;
      break;
    } else if (line.startsWith('//')) {
      // 単行コメント
      commentLines.unshift(line.substring(2).trim());
    } else {
      // コメント以外の行が見つかったら終了
      break;
    }
  }

  // JSDocスタイルのコメントブロックを処理
  if (commentStartIndex !== -1) {
    let inJSDoc = false;
    for (let i = commentStartIndex; i < functionLineIndex; i++) {
      const line = lines[i].trim();
      
      if (line.startsWith('/**') || line.startsWith('/*')) {
        inJSDoc = true;
        // 開始行にコメントがある場合
        const comment = line.replace(/^\/\*\*?/, '').replace(/\*\/$/, '').trim();
        if (comment) {
          commentLines.push(comment);
        }
      } else if (line.endsWith('*/')) {
        inJSDoc = false;
        // 終了行にコメントがある場合
        const comment = line.replace(/\*\/$/, '').replace(/^\*/, '').trim();
        if (comment) {
          commentLines.push(comment);
        }
      } else if (inJSDoc && (line.startsWith('*') || line.startsWith(' *'))) {
        const comment = line.replace(/^\s*\*/, '').trim();
        if (comment) {
          commentLines.push(comment);
        }
      }
    }
  }

  if (commentLines.length === 0) {
    return null;
  }

  // JSDocタグを解析
  const jsDocInfo: JSDocInfo = {
    description: '',
    params: [],
    returns: undefined,
    example: undefined,
    since: undefined,
    deprecated: undefined,
    throws: [],
    author: undefined,
    version: undefined
  };

  let currentSection = 'description';
  let descriptionLines: string[] = [];
  let exampleLines: string[] = [];

  for (const line of commentLines) {
    if (line.startsWith('@param')) {
      currentSection = 'param';
      // @param {type} name description または @param name description
      const paramMatch = line.match(/@param\s+(?:\{([^}]+)\}\s+)?(\w+)\s*(.*)/);
      if (paramMatch) {
        jsDocInfo.params.push({
          name: paramMatch[2],
          type: paramMatch[1] || undefined,
          description: paramMatch[3] || ''
        });
      }
    } else if (line.startsWith('@returns') || line.startsWith('@return')) {
      currentSection = 'returns';
      const returnMatch = line.match(/@returns?\s+(.*)/);
      if (returnMatch) {
        jsDocInfo.returns = returnMatch[1];
      }
    } else if (line.startsWith('@example')) {
      currentSection = 'example';
      const exampleMatch = line.match(/@example\s*(.*)/);
      if (exampleMatch && exampleMatch[1]) {
        exampleLines.push(exampleMatch[1]);
      }
    } else if (line.startsWith('@since')) {
      const sinceMatch = line.match(/@since\s+(.*)/);
      if (sinceMatch) {
        jsDocInfo.since = sinceMatch[1];
      }
    } else if (line.startsWith('@deprecated')) {
      const deprecatedMatch = line.match(/@deprecated\s*(.*)/);
      jsDocInfo.deprecated = deprecatedMatch ? deprecatedMatch[1] : 'This function is deprecated';
    } else if (line.startsWith('@throws') || line.startsWith('@throw')) {
      const throwMatch = line.match(/@throws?\s+(?:\{([^}]+)\}\s+)?(.*)/);
      if (throwMatch) {
        jsDocInfo.throws = jsDocInfo.throws || [];
        jsDocInfo.throws.push({
          type: throwMatch[1] || 'Error',
          description: throwMatch[2] || ''
        });
      }
    } else if (line.startsWith('@author')) {
      const authorMatch = line.match(/@author\s+(.*)/);
      if (authorMatch) {
        jsDocInfo.author = authorMatch[1];
      }
    } else if (line.startsWith('@version')) {
      const versionMatch = line.match(/@version\s+(.*)/);
      if (versionMatch) {
        jsDocInfo.version = versionMatch[1];
      }
    } else {
      // 通常のコメント行
      if (currentSection === 'description') {
        descriptionLines.push(line);
      } else if (currentSection === 'example') {
        exampleLines.push(line);
      } else if (currentSection === 'param' && jsDocInfo.params.length > 0) {
        // 前のパラメータの説明の続き
        const lastParam = jsDocInfo.params[jsDocInfo.params.length - 1];
        lastParam.description += ' ' + line;
      } else if (currentSection === 'returns' && jsDocInfo.returns) {
        // 戻り値の説明の続き
        jsDocInfo.returns += ' ' + line;
      }
    }
  }

  jsDocInfo.description = descriptionLines.join(' ').trim();
  if (exampleLines.length > 0) {
    jsDocInfo.example = exampleLines.join('\n').trim();
  }

  return jsDocInfo;
}

// TypeScript風の洗練されたホバー情報を取得する関数
function getEnhancedTypeInfo(word: string, document: vscode.TextDocument, position: vscode.Position): { content: string; category: string } | undefined {
  const text = document.getText();
  const lines = text.split('\n');

  // 1. 標準ライブラリ関数の場合
  const stdFunctions = {
    'println': {
      signature: 'fn println(...args: any): void',
      description: 'Outputs values to the console followed by a newline',
      module: 'std::io',
      example: 'println("Hello, World!")\nprintln("Value:", x, 42, true)',
      returns: 'void',
      category: 'std function',
      tags: ['io', 'output', 'console']
    },
    'print': {
      signature: 'fn print(...args: any): void',
      description: 'Outputs values to the console without a trailing newline',
      module: 'std::io',
      example: 'print("Hello, ")\nprint("Value: ", x)',
      returns: 'void',
      category: 'std function',
      tags: ['io', 'output', 'console']
    },
    'read_line': {
      signature: 'fn read_line(): string',
      description: 'Reads a line from standard input',
      module: 'std::io',
      example: 'let input = read_line()\nlet name = read_line()',
      returns: 'string',
      category: 'std function',
      tags: ['io', 'input', 'console']
    },
    'write_file': {
      signature: 'fn write_file(path: string, content: string): Result<(), Error>',
      description: 'Writes content to a file at the specified path',
      module: 'std::fs',
      example: 'write_file("output.txt", "Hello, World!")\nwrite_file("data.json", json_str)',
      returns: 'Result<(), Error>',
      category: 'std function',
      tags: ['io', 'filesystem', 'write']
    },
    'read_file': {
      signature: 'fn read_file(path: string): Result<string, Error>',
      description: 'Reads the entire content of a file as a string',
      module: 'std::fs',
      example: 'let content = read_file("input.txt")\nlet config = read_file("config.json")',
      returns: 'Result<string, Error>',
      category: 'std function',
      tags: ['io', 'filesystem', 'read']
    },
    'len': {
      signature: 'fn len<T>(collection: T): int where T: Collection',
      description: 'Returns the length of a collection (string, array, etc.)',
      module: 'std::collections',
      example: 'len("hello")  // 5\nlen([1, 2, 3])  // 3',
      returns: 'int',
      category: 'std function',
      tags: ['collections', 'utility']
    },
    'push': {
      signature: 'fn push<T>(array: &mut [T], item: T): void',
      description: 'Adds an element to the end of an array',
      module: 'std::collections',
      example: 'let mut arr = [1, 2]\npush(&mut arr, 3)  // [1, 2, 3]',
      returns: 'void',
      category: 'std function',
      tags: ['collections', 'array', 'mutating']
    }
  };

  if (stdFunctions[word as keyof typeof stdFunctions]) {
    const func = stdFunctions[word as keyof typeof stdFunctions];
    
    let content = `\`\`\`zeno\n${func.signature}\n\`\`\`\n\n`;
    content += `$(symbol-function) **${word}** *(${func.module})*\n\n`;
    content += `${func.description}\n\n`;
    
    // タグを美しく表示
    if (func.tags.length > 0) {
      content += `**Tags**: ${func.tags.map(tag => `\`${tag}\``).join(', ')}\n\n`;
    }
    
    content += `**Returns**: \`${func.returns}\`\n\n`;
    content += `**Examples**:\n\`\`\`zeno\n${func.example}\n\`\`\``;
    
    return { content, category: func.category };
  }

  // 2. ユーザー定義関数の場合（JSDocサポート付き）
  const funcPattern = new RegExp(`(pub\\s+)?fn\\s+${word}\\s*\\(([^)]*)\\)\\s*:\\s*([\\w\\[\\]<>,\\s]+)`);
  const funcMatch = text.match(funcPattern);
  if (funcMatch) {
    const visibility = funcMatch[1] ? 'pub ' : '';
    const params = funcMatch[2].trim();
    const returnType = funcMatch[3].trim();

    let signature = `${visibility}fn ${word}(${params}): ${returnType}`;
    let content = `\`\`\`zeno\n${signature}\n\`\`\`\n\n`;
    
    content += `$(symbol-function) **Function** \`${word}\`\n\n`;
    
    if (visibility) {
      content += `**Visibility**: \`${visibility.trim()}\`\n\n`;
    }

    // JSDocスタイルのコメントを解析
    const docInfo = findJSDocComment(word, lines);
    if (docInfo) {
      if (docInfo.description) {
        content += `**Description**:\n${docInfo.description}\n\n`;
      }
      
      // パラメータ説明を追加（JSDocとシグネチャを統合）
      if (params) {
        content += `**Parameters**:\n`;
        const paramList = params.split(',').map(p => p.trim()).filter(p => p);
        paramList.forEach(param => {
          const [name, type] = param.split(':').map(s => s.trim());
          const docParam = docInfo.params.find(p => p.name === name);
          if (docParam) {
            const paramType = docParam.type || type;
            content += `- \`${name}\` (\`${paramType}\`): ${docParam.description}\n`;
          } else {
            content += `- \`${name}\`: \`${type}\`\n`;
          }
        });
        content += '\n';
      }
      
      // 戻り値説明を追加
      if (docInfo.returns) {
        content += `**Returns**: ${docInfo.returns}\n\n`;
      } else {
        content += `**Returns**: \`${returnType}\`\n\n`;
      }
      
      // 例を追加
      if (docInfo.example) {
        content += `**Example**:\n\`\`\`zeno\n${docInfo.example}\n\`\`\`\n\n`;
      }
      
      // その他のタグを追加
      if (docInfo.since) {
        content += `**Since**: ${docInfo.since}\n\n`;
      }
      
      if (docInfo.deprecated) {
        content += `⚠️ **Deprecated**: ${docInfo.deprecated}\n\n`;
      }
      
      if (docInfo.author) {
        content += `**Author**: ${docInfo.author}\n\n`;
      }
      
      if (docInfo.version) {
        content += `**Version**: ${docInfo.version}\n\n`;
      }
      
      if (docInfo.throws && docInfo.throws.length > 0) {
        content += `**Throws**:\n`;
        docInfo.throws.forEach(throwInfo => {
          content += `- \`${throwInfo.type}\`: ${throwInfo.description}\n`;
        });
        content += '\n';
      }
    } else {
      // JSDocコメントがない場合は基本情報のみ
      content += `**Returns**: \`${returnType}\`\n\n`;
      if (params) {
        content += `**Parameters**:\n`;
        const paramList = params.split(',').map(p => p.trim()).filter(p => p);
        paramList.forEach(param => {
          const [name, type] = param.split(':').map(s => s.trim());
          content += `- \`${name}\`: \`${type}\`\n`;
        });
        content += '\n';
      }
    }

    return { content, category: 'user function' };
  }

  // 3. 変数宣言の場合（型注釈あり）
  const letTypePattern = new RegExp(`let\\s+${word}\\s*:\\s*([\\w\\[\\]<>,\\s]+)\\s*=\\s*([^\\n;]+)`);
  const letTypeMatch = text.match(letTypePattern);
  if (letTypeMatch) {
    const type = letTypeMatch[1].trim();
    const value = letTypeMatch[2].trim();

    let content = `\`\`\`zeno\nlet ${word}: ${type}\n\`\`\`\n\n`;
    content += `$(symbol-variable) **Variable** \`${word}\`\n\n`;
    content += `**Type**: \`${type}\`\n\n`;
    content += `**Value**: \`${value}\`\n\n`;
    
    // 型の説明を追加
    const typeDescription = getTypeDescription(type);
    if (typeDescription) {
      content += `**Type Info**: ${typeDescription}`;
    }

    return { content, category: 'variable' };
  }

  // 4. 変数宣言の場合（型推論）
  const letInferPattern = new RegExp(`let\\s+${word}\\s*=\\s*([^\\n;]+)`);
  const letInferMatch = text.match(letInferPattern);
  if (letInferMatch) {
    const value = letInferMatch[1].trim();
    const inferredType = inferType(value, text);
    
    let content = `\`\`\`zeno\nlet ${word}: ${inferredType.type} // inferred\n\`\`\`\n\n`;
    content += `$(symbol-variable) **Variable** \`${word}\`\n\n`;
    content += `**Type**: \`${inferredType.type}\` *(inferred)*\n\n`;
    content += `**Value**: \`${value}\`\n\n`;
    content += `**Inference**: ${inferredType.description}`;

    return { content, category: 'variable' };
  }

  // 5. キーワードの場合
  const keywords = {
    'fn': {
      description: 'Declares a function',
      syntax: 'fn name(params): return_type { body }',
      example: 'fn add(a: int, b: int): int {\n    return a + b\n}'
    },
    'let': {
      description: 'Declares a variable binding',
      syntax: 'let name: type = value',
      example: 'let x: int = 42\nlet name = "Zeno"  // inferred'
    },
    'if': {
      description: 'Conditional execution',
      syntax: 'if condition { body } else { alternative }',
      example: 'if x > 0 {\n    println("positive")\n} else {\n    println("not positive")\n}'
    },
    'else': {
      description: 'Alternative execution branch',
      syntax: 'else { body }',
      example: 'if condition {\n    // ...\n} else {\n    println("alternative")\n}'
    },
    'while': {
      description: 'Loop with condition',
      syntax: 'while condition { body }',
      example: 'while i < 10 {\n    println(i)\n    i = i + 1\n}'
    },
    'for': {
      description: 'Iteration over collections',
      syntax: 'for item in collection { body }',
      example: 'for i in 0..10 {\n    println(i)\n}\nfor item in array {\n    println(item)\n}'
    },
    'return': {
      description: 'Returns a value from function',
      syntax: 'return expression',
      example: 'fn double(x: int): int {\n    return x * 2\n}'
    },
    'pub': {
      description: 'Makes declarations public',
      syntax: 'pub fn/let/type name',
      example: 'pub fn public_function() {\n    // visible outside module\n}'
    },
    'import': {
      description: 'Imports module or symbols',
      syntax: 'import module_name\nimport { symbol } from module',
      example: 'import std::io\nimport { println } from std::io'
    },
    'from': {
      description: 'Specifies import source',
      syntax: 'import { symbols } from module',
      example: 'import { HashMap } from std::collections'
    }
  };

  if (keywords[word as keyof typeof keywords]) {
    const keyword = keywords[word as keyof typeof keywords];
    
    let content = `\`\`\`zeno\n${keyword.syntax}\n\`\`\`\n\n`;
    content += `$(symbol-keyword) **Keyword** \`${word}\`\n\n`;
    content += `${keyword.description}\n\n`;
    content += `**Example**:\n\`\`\`zeno\n${keyword.example}\n\`\`\``;

    return { content, category: 'keyword' };
  }

  return undefined;
}

// 型の説明を取得するヘルパー関数
function getTypeDescription(type: string): string | undefined {
  const types: { [key: string]: string } = {
    'int': 'Signed integer type (32-bit)',
    'string': 'UTF-8 string type',
    'bool': 'Boolean type (true/false)',
    'float': 'Floating point number (64-bit)',
    'array': 'Dynamic array type',
    'void': 'Unit type representing no value'
  };

  return types[type];
}

// 型推論を行うヘルパー関数
function inferType(value: string, text: string): { type: string; description: string } {
  if (/^\d+$/.test(value)) {
    return { type: 'int', description: 'Integer literal' };
  } else if (/^\d+\.\d+$/.test(value)) {
    return { type: 'float', description: 'Floating point literal' };
  } else if (/^".*"$/.test(value)) {
    return { type: 'string', description: 'String literal' };
  } else if (value === 'true' || value === 'false') {
    return { type: 'bool', description: 'Boolean literal' };
  } else if (/^\[.*\]$/.test(value)) {
    return { type: 'array', description: 'Array literal' };
  } else {
    // 関数呼び出しの場合
    const funcCallMatch = value.match(/^(\w+)\s*\(/);
    if (funcCallMatch) {
      const funcName = funcCallMatch[1];
      const funcReturnType = getFunctionReturnType(funcName, text);
      if (funcReturnType) {
        return { type: funcReturnType, description: `Return value from function \`${funcName}()\`` };
      }
    }
    return { type: 'unknown', description: 'Unable to infer type' };
  }
}

// 関数の戻り値型を取得
function getFunctionReturnType(funcName: string, text: string): string | undefined {
  const funcPattern = new RegExp(`fn\\s+${funcName}\\s*\\([^)]*\\)\\s*:\\s*([\\w\\[\\]<>,\\s]+)`, 'g');
  const match = funcPattern.exec(text);
  return match ? match[1].trim() : undefined;
}

// 定義を検索する関数
function findDefinition(word: string, document: vscode.TextDocument): vscode.Location[] {
  const text = document.getText();
  const lines = text.split('\n');
  const locations: vscode.Location[] = [];

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
function getCompletionItems(document: vscode.TextDocument, position: vscode.Position): vscode.CompletionItem[] {
  const items: vscode.CompletionItem[] = [];

  // Zeno言語のキーワード
  const keywords = [
    { name: 'fn', detail: 'Function declaration', documentation: 'Declares a new function' },
    { name: 'let', detail: 'Variable declaration', documentation: 'Declares a new variable' },
    { name: 'if', detail: 'Conditional statement', documentation: 'Conditional execution' },
    { name: 'else', detail: 'Alternative branch', documentation: 'Alternative execution path' },
    { name: 'while', detail: 'Loop statement', documentation: 'Repeats code while condition is true' },
    { name: 'for', detail: 'Iteration statement', documentation: 'Iterates over a collection' },
    { name: 'return', detail: 'Return statement', documentation: 'Returns a value from function' },
    { name: 'import', detail: 'Import statement', documentation: 'Imports a module' },
    { name: 'pub', detail: 'Public modifier', documentation: 'Makes function/variable public' }
  ];

  keywords.forEach(keyword => {
    const item = new vscode.CompletionItem(keyword.name, vscode.CompletionItemKind.Keyword);
    item.insertText = keyword.name;
    item.detail = keyword.detail;
    item.documentation = new vscode.MarkdownString(keyword.documentation);
    items.push(item);
  });

  // 型の補完
  const types = [
    { name: 'int', description: 'Integer type' },
    { name: 'string', description: 'String type' },
    { name: 'bool', description: 'Boolean type' },
    { name: 'float', description: 'Floating point type' },
    { name: 'array', description: 'Array type' }
  ];

  types.forEach(type => {
    const item = new vscode.CompletionItem(type.name, vscode.CompletionItemKind.TypeParameter);
    item.insertText = type.name;
    item.detail = `Zeno type: ${type.name}`;
    item.documentation = new vscode.MarkdownString(type.description);
    items.push(item);
  });

  // 標準ライブラリ関数
  const stdFunctions = [
    {
      name: 'println',
      signature: 'fn println(...args: any): void',
      description: 'Outputs values to the console followed by a newline',
      detail: 'std function'
    },
    {
      name: 'print',
      signature: 'fn print(...args: any): void',
      description: 'Outputs values to the console without a trailing newline',
      detail: 'std function'
    },
    {
      name: 'read_line',
      signature: 'fn read_line(): string',
      description: 'Reads a line from standard input',
      detail: 'std function'
    },
    {
      name: 'write_file',
      signature: 'fn write_file(path: string, content: string): Result<(), Error>',
      description: 'Writes content to a file at the specified path',
      detail: 'std function'
    },
    {
      name: 'read_file',
      signature: 'fn read_file(path: string): Result<string, Error>',
      description: 'Reads the entire content of a file as a string',
      detail: 'std function'
    }
  ];

  stdFunctions.forEach(func => {
    const item = new vscode.CompletionItem(func.name, vscode.CompletionItemKind.Function);
    item.insertText = func.name;
    item.detail = func.detail;
    item.documentation = new vscode.MarkdownString(`\`\`\`zeno\n${func.signature}\n\`\`\`\n\n${func.description}`);
    items.push(item);
  });

  return items;
}
