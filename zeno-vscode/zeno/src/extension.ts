import * as vscode from 'vscode';

export function activate(context: vscode.ExtensionContext) {
  console.log('Zeno Language Features extension is now active!');

  // ホバープロバイダーを直接登録
  const hoverProvider = vscode.languages.registerHoverProvider('zeno', {
    provideHover(document, position, token) {
      const range = document.getWordRangeAtPosition(position);
      if (!range) return undefined;

      const word = document.getText(range);

      // 簡単な型情報の提供
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
  // 何もしない
  return undefined;
}

// 型情報を取得する関数（TypeScript風の美しいホバー表示）
function getTypeInfo(word: string, document: vscode.TextDocument, position: vscode.Position): string | undefined {
  const text = document.getText();
  const lines = text.split('\n');

  // 1. 関数定義の場合
  const funcPattern = new RegExp(`(pub\\s+)?fn\\s+${word}\\s*\\(([^)]*)\\)\\s*:\\s*([\\w\\[\\]]+)`);
  const funcMatch = text.match(funcPattern);
  if (funcMatch) {
    const visibility = funcMatch[1] ? 'pub ' : '';
    const params = funcMatch[2].trim();
    const returnType = funcMatch[3];

    let signature = `${visibility}fn ${word}(${params}): ${returnType}`;
    let hoverText = `\`\`\`zeno\n${signature}\n\`\`\`\n\n`;

    // 関数の説明を追加
    hoverText += `**Function**: \`${word}\`\n\n`;
    hoverText += `**Returns**: \`${returnType}\`\n\n`;

    if (params) {
      hoverText += `**Parameters**:\n`;
      const paramList = params.split(',').map(p => p.trim());
      paramList.forEach(param => {
        if (param) {
          const [name, type] = param.split(':').map(s => s.trim());
          hoverText += `- \`${name}\`: \`${type}\`\n`;
        }
      });
    }

    return hoverText;
  }

  // 2. 変数宣言の場合（明示的な型注釈）
  const letPattern = new RegExp(`let\\s+${word}\\s*:\\s*([\\w\\[\\]]+)\\s*=\\s*([^\\n;]+)`);
  const letMatch = text.match(letPattern);
  if (letMatch) {
    const type = letMatch[1];
    const value = letMatch[2].trim();

    let hoverText = `\`\`\`zeno\nlet ${word}: ${type}\n\`\`\`\n\n`;
    hoverText += `**Variable**: \`${word}\`\n\n`;
    hoverText += `**Type**: \`${type}\`\n\n`;
    hoverText += `**Value**: \`${value}\``;

    return hoverText;
  }

  // 3. 関数パラメータの場合
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const funcDefMatch = line.match(/fn\s+\w+\s*\(([^)]*)\)/);
    if (funcDefMatch) {
      const params = funcDefMatch[1];
      const paramPattern = new RegExp(`${word}\\s*:\\s*([\\w\\[\\]]+)`);
      const paramMatch = params.match(paramPattern);
      if (paramMatch) {
        const type = paramMatch[1];
        let hoverText = `\`\`\`zeno\n${word}: ${type}\n\`\`\`\n\n`;
        hoverText += `**Parameter**: \`${word}\`\n\n`;
        hoverText += `**Type**: \`${type}\``;
        return hoverText;
      }
    }
  }

  // 4. 型推論による判定（より詳細）
  const assignPattern = new RegExp(`let\\s+${word}\\s*=\\s*([^\\n;]+)`);
  const assignMatch = text.match(assignPattern);
  if (assignMatch) {
    const value = assignMatch[1].trim();
    let inferredType = '';
    let description = '';

    if (/^\d+$/.test(value)) {
      inferredType = 'int';
      description = 'Integer literal';
    } else if (/^\d+\.\d+$/.test(value)) {
      inferredType = 'float';
      description = 'Floating point literal';
    } else if (/^".*"$/.test(value)) {
      inferredType = 'string';
      description = 'String literal';
    } else if (value === 'true' || value === 'false') {
      inferredType = 'bool';
      description = 'Boolean literal';
    } else if (/^\[.*\]$/.test(value)) {
      inferredType = 'array';
      description = 'Array literal';
    } else {
      // 関数呼び出しの場合
      const funcCallMatch = value.match(/^(\w+)\s*\(/);
      if (funcCallMatch) {
        const funcName = funcCallMatch[1];
        const funcReturnType = getFunctionReturnType(funcName, text);
        if (funcReturnType) {
          inferredType = funcReturnType;
          description = `Return value from function \`${funcName}()\``;
        }
      }
    }

    if (inferredType) {
      let hoverText = `\`\`\`zeno\nlet ${word}: ${inferredType} // inferred\n\`\`\`\n\n`;
      hoverText += `**Variable**: \`${word}\`\n\n`;
      hoverText += `**Type**: \`${inferredType}\` *(inferred)*\n\n`;
      hoverText += `**Value**: \`${value}\`\n\n`;
      hoverText += `**Description**: ${description}`;
      return hoverText;
    }
  }

  // 5. 標準ライブラリ関数の場合（TypeScript風の詳細情報）
  const stdFunctions = {
    'println': {
      signature: 'fn println(message: string): void',
      description: 'Prints a message to standard output followed by a newline',
      module: 'built-in',
      example: 'println("Hello, World!")',
      returns: 'void'
    },
    'print': {
      signature: 'fn print(message: string): void',
      description: 'Prints a message to standard output without a trailing newline',
      module: 'built-in',
      example: 'print("Hello, ")',
      returns: 'void'
    },
    'read_line': {
      signature: 'fn read_line(): string',
      description: 'Reads a line from standard input',
      module: 'std.io',
      example: 'let input = read_line()',
      returns: 'string'
    },
    'write_file': {
      signature: 'fn write_file(path: string, content: string): bool',
      description: 'Writes content to a file at the specified path',
      module: 'std.io',
      example: 'write_file("output.txt", "Hello")',
      returns: 'bool'
    }
  };

  if (stdFunctions[word as keyof typeof stdFunctions]) {
    const func = stdFunctions[word as keyof typeof stdFunctions];
    let hoverText = `\`\`\`zeno\n${func.signature}\n\`\`\`\n\n`;
    hoverText += `**Function**: \`${word}\` *(${func.module})*\n\n`;
    hoverText += `**Description**: ${func.description}\n\n`;
    hoverText += `**Returns**: \`${func.returns}\`\n\n`;
    hoverText += `**Example**:\n\`\`\`zeno\n${func.example}\n\`\`\``;
    return hoverText;
  }

  // 6. キーワードの場合
  const keywords = {
    'fn': 'Function declaration keyword',
    'let': 'Variable declaration keyword',
    'if': 'Conditional statement keyword',
    'else': 'Alternative branch keyword',
    'while': 'Loop statement keyword',
    'for': 'Iteration statement keyword',
    'return': 'Function return keyword',
    'pub': 'Public visibility modifier',
    'import': 'Module import keyword',
    'from': 'Import source specifier'
  };

  if (keywords[word as keyof typeof keywords]) {
    let hoverText = `\`\`\`zeno\n${word}\n\`\`\`\n\n`;
    hoverText += `**Keyword**: \`${word}\`\n\n`;
    hoverText += `**Description**: ${keywords[word as keyof typeof keywords]}`;
    return hoverText;
  }

  return undefined;
}

// 関数の戻り値型を取得
function getFunctionReturnType(funcName: string, text: string): string | undefined {
  const funcPattern = new RegExp(`fn\\s+${funcName}\\s*\\([^)]*\\)\\s*:\\s*([\\w\\[\\]]+)`, 'g');
  const match = funcPattern.exec(text);
  return match ? match[1] : undefined;
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

// 補完アイテムを提供する関数（TypeScript風の詳細な補完）
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

  // 標準ライブラリ関数（より詳細）
  const stdFunctions = [
    {
      name: 'println',
      detail: 'fn println(message: string): void',
      doc: 'Prints a message to standard output followed by a newline\n\n**Example:**\n```zeno\nprintln("Hello, World!")\n```',
      snippet: 'println($1)'
    },
    {
      name: 'print',
      detail: 'fn print(message: string): void',
      doc: 'Prints a message to standard output without a trailing newline\n\n**Example:**\n```zeno\nprint("Hello, ")\n```',
      snippet: 'print($1)'
    },
    {
      name: 'read_line',
      detail: 'fn read_line(): string',
      doc: 'Reads a line from standard input\n\n**Example:**\n```zeno\nlet input = read_line()\n```',
      snippet: 'read_line()'
    },
    {
      name: 'write_file',
      detail: 'fn write_file(path: string, content: string): bool',
      doc: 'Writes content to a file at the specified path\n\n**Example:**\n```zeno\nwrite_file("output.txt", "Hello")\n```',
      snippet: 'write_file($1, $2)'
    }
  ];

  stdFunctions.forEach(func => {
    const item = new vscode.CompletionItem(func.name, vscode.CompletionItemKind.Function);
    item.detail = func.detail;
    item.documentation = new vscode.MarkdownString(func.doc);
    item.insertText = new vscode.SnippetString(func.snippet);
    items.push(item);
  });

  return items;
}
