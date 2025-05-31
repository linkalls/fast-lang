// import * as vscode from 'vscode';

// export function activate(context: vscode.ExtensionContext) {
//   console.log('Zeno Language Features extension is now active!');

//   // より洗練されたホバープロバイダーを登録
//   const hoverProvider = vscode.languages.registerHoverProvider('zeno', {
//     provideHover(document, position, token) {
//       const range = document.getWordRangeAtPosition(position);
//       if (!range) return undefined;

//       const word = document.getText(range);
      
//       // TypeScript風の美しい型情報を提供
//       const hoverInfo = getEnhancedTypeInfo(word, document, position);
//       if (hoverInfo) {
//         const markdown = new vscode.MarkdownString(hoverInfo.content);
//         markdown.isTrusted = true;
//         markdown.supportHtml = true;
//         return new vscode.Hover(markdown, range);
//       }
//     }
//   });

//   // 定義プロバイダーを登録
//   const definitionProvider = vscode.languages.registerDefinitionProvider('zeno', {
//     provideDefinition(document, position, token) {
//       const range = document.getWordRangeAtPosition(position);
//       if (!range) return undefined;

//       const word = document.getText(range);

//       // 関数定義を検索
//       return findDefinition(word, document);
//     }
//   });

//   // 補完プロバイダーを登録
//   const completionProvider = vscode.languages.registerCompletionItemProvider('zeno', {
//     provideCompletionItems(document, position, token, context) {
//       return getCompletionItems(document, position);
//     }
//   }, '.', ':');

//   context.subscriptions.push(hoverProvider, definitionProvider, completionProvider);
// }

// export function deactivate(): Thenable<void> | undefined {
//   // 何もしない
//   return undefined;
// }

// // TypeScript風の洗練されたホバー情報を取得する関数
// function getEnhancedTypeInfo(word: string, document: vscode.TextDocument, position: vscode.Position): { content: string; category: string } | undefined {
//   const text = document.getText();
//   const lines = text.split('\n');
//   const line = lines[position.line];

//   // コンテキスト分析
//   const context = analyzeContext(line, position.character, word);

//   // 1. 標準ライブラリ関数の場合（TypeScript風の美しい表示）
//   const stdFunctions = {
//     'println': {
//       signature: 'fn println(...args: any): void',
//       description: 'Outputs values to the console followed by a newline',
//       module: 'std::io',
//       example: 'println("Hello, World!")\nprintln("Value:", x, 42, true)',
//       returns: 'void',
//       category: 'std function',
//       tags: ['io', 'output', 'console']
//     },
//     'print': {
//       signature: 'fn print(...args: any): void',
//       description: 'Outputs values to the console without a trailing newline',
//       module: 'std::io',
//       example: 'print("Hello, ")\nprint("Value: ", x)',
//       returns: 'void',
//       category: 'std function',
//       tags: ['io', 'output', 'console']
//     },
//     'read_line': {
//       signature: 'fn read_line(): string',
//       description: 'Reads a line from standard input',
//       module: 'std::io',
//       example: 'let input = read_line()\nlet name = read_line()',
//       returns: 'string',
//       category: 'std function',
//       tags: ['io', 'input', 'console']
//     },
//     'write_file': {
//       signature: 'fn write_file(path: string, content: string): Result<(), Error>',
//       description: 'Writes content to a file at the specified path',
//       module: 'std::fs',
//       example: 'write_file("output.txt", "Hello, World!")\nwrite_file("data.json", json_str)',
//       returns: 'Result<(), Error>',
//       category: 'std function',
//       tags: ['io', 'filesystem', 'write']
//     },
//     'read_file': {
//       signature: 'fn read_file(path: string): Result<string, Error>',
//       description: 'Reads the entire content of a file as a string',
//       module: 'std::fs',
//       example: 'let content = read_file("input.txt")\nlet config = read_file("config.json")',
//       returns: 'Result<string, Error>',
//       category: 'std function',
//       tags: ['io', 'filesystem', 'read']
//     },
//     'len': {
//       signature: 'fn len<T>(collection: T): int where T: Collection',
//       description: 'Returns the length of a collection (string, array, etc.)',
//       module: 'std::collections',
//       example: 'len("hello")  // 5\nlen([1, 2, 3])  // 3',
//       returns: 'int',
//       category: 'std function',
//       tags: ['collections', 'utility']
//     },
//     'push': {
//       signature: 'fn push<T>(array: &mut [T], item: T): void',
//       description: 'Adds an element to the end of an array',
//       module: 'std::collections',
//       example: 'let mut arr = [1, 2]\npush(&mut arr, 3)  // [1, 2, 3]',
//       returns: 'void',
//       category: 'std function',
//       tags: ['collections', 'array', 'mutating']
//     }
//   };

//   if (stdFunctions[word as keyof typeof stdFunctions]) {
//     const func = stdFunctions[word as keyof typeof stdFunctions];
    
//     let content = `\`\`\`zeno\n${func.signature}\n\`\`\`\n\n`;
//     content += `$(symbol-function) **${word}** *(${func.module})*\n\n`;
//     content += `${func.description}\n\n`;
    
//     // タグを美しく表示
//     if (func.tags.length > 0) {
//       content += `**Tags**: ${func.tags.map(tag => `\`${tag}\``).join(', ')}\n\n`;
//     }
    
//     content += `**Returns**: \`${func.returns}\`\n\n`;
//     content += `**Examples**:\n\`\`\`zeno\n${func.example}\n\`\`\``;
    
//     return { content, category: func.category };
//   }

//   // 2. 関数定義の場合（TypeScript風の洗練された表示）
//   const funcPattern = new RegExp(`(pub\\s+)?fn\\s+${word}\\s*\\(([^)]*)\\)\\s*:\\s*([\\w\\[\\]<>,\\s]+)`);
//   const funcMatch = text.match(funcPattern);
//   if (funcMatch) {
//     const visibility = funcMatch[1] ? 'pub ' : '';
//     const params = funcMatch[2].trim();
//     const returnType = funcMatch[3].trim();

//     let signature = `${visibility}fn ${word}(${params}): ${returnType}`;
//     let content = `\`\`\`zeno\n${signature}\n\`\`\`\n\n`;
    
//     content += `$(symbol-function) **Function** \`${word}\`\n\n`;
    
//     if (visibility) {
//       content += `**Visibility**: \`${visibility.trim()}\`\n\n`;
//     }
    
//     content += `**Returns**: \`${returnType}\`\n\n`;

//     if (params) {
//       content += `**Parameters**:\n`;
//       const paramList = params.split(',').map(p => p.trim()).filter(p => p);
//       paramList.forEach(param => {
//         const [name, type] = param.split(':').map(s => s.trim());
//         content += `- \`${name}\`: \`${type}\`\n`;
//       });
//       content += '\n';
//     }

//     // 関数のドキュメントコメントを探す（JSDocスタイル対応）
//     const docInfo = findJSDocComment(word, lines);
//     if (docInfo) {
//       content += `**Description**:\n${docInfo.description}\n\n`;
      
//       // パラメータ説明を追加
//       if (docInfo.params.length > 0) {
//         content += `**Parameters**:\n`;
//         docInfo.params.forEach(param => {
//           content += `- \`${param.name}\` (\`${param.type || 'any'}\`): ${param.description}\n`;
//         });
//         content += '\n';
//       }
      
//       // 戻り値説明を追加
//       if (docInfo.returns) {
//         content += `**Returns**: ${docInfo.returns}\n\n`;
//       }
      
//       // 例を追加
//       if (docInfo.example) {
//         content += `**Example**:\n\`\`\`zeno\n${docInfo.example}\n\`\`\`\n\n`;
//       }
      
//       // その他のタグを追加
//       if (docInfo.since) {
//         content += `**Since**: ${docInfo.since}\n\n`;
//       }
      
//       if (docInfo.deprecated) {
//         content += `⚠️ **Deprecated**: ${docInfo.deprecated}\n\n`;
//       }
      
//       if (docInfo.throws && docInfo.throws.length > 0) {
//         content += `**Throws**:\n`;
//         docInfo.throws.forEach(throwInfo => {
//           content += `- \`${throwInfo.type}\`: ${throwInfo.description}\n`;
//         });
//         content += '\n';
//       }
//     }

//     return { content, category: 'user function' };
//   }

//   // 3. 変数宣言の場合（型注釈あり）
//   const letTypePattern = new RegExp(`let\\s+${word}\\s*:\\s*([\\w\\[\\]<>,\\s]+)\\s*=\\s*([^\\n;]+)`);
//   const letTypeMatch = text.match(letTypePattern);
//   if (letTypeMatch) {
//     const type = letTypeMatch[1].trim();
//     const value = letTypeMatch[2].trim();

//     let content = `\`\`\`zeno\nlet ${word}: ${type}\n\`\`\`\n\n`;
//     content += `$(symbol-variable) **Variable** \`${word}\`\n\n`;
//     content += `**Type**: \`${type}\`\n\n`;
//     content += `**Value**: \`${value}\`\n\n`;
    
//     // 型の説明を追加
//     const typeDescription = getTypeDescription(type);
//     if (typeDescription) {
//       content += `**Type Info**: ${typeDescription}`;
//     }

//     return { content, category: 'variable' };
//   }

//   // 4. 変数宣言の場合（型推論）
//   const letInferPattern = new RegExp(`let\\s+${word}\\s*=\\s*([^\\n;]+)`);
//   const letInferMatch = text.match(letInferPattern);
//   if (letInferMatch) {
//     const value = letInferMatch[1].trim();
//     const inferredType = inferType(value, text);
    
//     let content = `\`\`\`zeno\nlet ${word}: ${inferredType.type} // inferred\n\`\`\`\n\n`;
//     content += `$(symbol-variable) **Variable** \`${word}\`\n\n`;
//     content += `**Type**: \`${inferredType.type}\` *(inferred)*\n\n`;
//     content += `**Value**: \`${value}\`\n\n`;
//     content += `**Inference**: ${inferredType.description}`;

//     return { content, category: 'variable' };
//   }

//   // 5. 関数パラメータの場合
//   for (let i = 0; i < lines.length; i++) {
//     const line = lines[i];
//     const funcDefMatch = line.match(/fn\s+(\w+)\s*\(([^)]*)\)/);
//     if (funcDefMatch) {
//       const funcName = funcDefMatch[1];
//       const params = funcDefMatch[2];
//       const paramPattern = new RegExp(`${word}\\s*:\\s*([\\w\\[\\]<>,\\s]+)`);
//       const paramMatch = params.match(paramPattern);
//       if (paramMatch) {
//         const type = paramMatch[1].trim();
        
//         let content = `\`\`\`zeno\n${word}: ${type}\n\`\`\`\n\n`;
//         content += `$(symbol-parameter) **Parameter** \`${word}\`\n\n`;
//         content += `**Function**: \`${funcName}\`\n\n`;
//         content += `**Type**: \`${type}\`\n\n`;
        
//         const typeDescription = getTypeDescription(type);
//         if (typeDescription) {
//           content += `**Type Info**: ${typeDescription}`;
//         }

//         return { content, category: 'parameter' };
//       }
//     }
//   }

//   // 6. キーワードの場合（TypeScript風のエレガントな表示）
//   const keywords = {
//     'fn': {
//       description: 'Declares a function',
//       syntax: 'fn name(params): return_type { body }',
//       example: 'fn add(a: int, b: int): int {\n    return a + b\n}'
//     },
//     'let': {
//       description: 'Declares a variable binding',
//       syntax: 'let name: type = value',
//       example: 'let x: int = 42\nlet name = "Zeno"  // inferred'
//     },
//     'if': {
//       description: 'Conditional execution',
//       syntax: 'if condition { body } else { alternative }',
//       example: 'if x > 0 {\n    println("positive")\n} else {\n    println("not positive")\n}'
//     },
//     'else': {
//       description: 'Alternative execution branch',
//       syntax: 'else { body }',
//       example: 'if condition {\n    // ...\n} else {\n    println("alternative")\n}'
//     },
//     'while': {
//       description: 'Loop with condition',
//       syntax: 'while condition { body }',
//       example: 'while i < 10 {\n    println(i)\n    i = i + 1\n}'
//     },
//     'for': {
//       description: 'Iteration over collections',
//       syntax: 'for item in collection { body }',
//       example: 'for i in 0..10 {\n    println(i)\n}\nfor item in array {\n    println(item)\n}'
//     },
//     'return': {
//       description: 'Returns a value from function',
//       syntax: 'return expression',
//       example: 'fn double(x: int): int {\n    return x * 2\n}'
//     },
//     'pub': {
//       description: 'Makes declarations public',
//       syntax: 'pub fn/let/type name',
//       example: 'pub fn public_function() {\n    // visible outside module\n}'
//     },
//     'import': {
//       description: 'Imports module or symbols',
//       syntax: 'import module_name\nimport { symbol } from module',
//       example: 'import std::io\nimport { println } from std::io'
//     },
//     'from': {
//       description: 'Specifies import source',
//       syntax: 'import { symbols } from module',
//       example: 'import { HashMap } from std::collections'
//     }
//   };

//   if (keywords[word as keyof typeof keywords]) {
//     const keyword = keywords[word as keyof typeof keywords];
    
//     let content = `\`\`\`zeno\n${keyword.syntax}\n\`\`\`\n\n`;
//     content += `$(symbol-keyword) **Keyword** \`${word}\`\n\n`;
//     content += `${keyword.description}\n\n`;
//     content += `**Example**:\n\`\`\`zeno\n${keyword.example}\n\`\`\``;

//     return { content, category: 'keyword' };
//   }

//   // 7. 型の場合
//   const types = {
//     'int': 'Signed integer type (32-bit)',
//     'string': 'UTF-8 string type',
//     'bool': 'Boolean type (true/false)',
//     'float': 'Floating point number (64-bit)',
//     'array': 'Dynamic array type',
//     'void': 'Unit type representing no value'
//   };

//   if (types[word as keyof typeof types]) {
//     const description = types[word as keyof typeof types];
    
//     let content = `\`\`\`zeno\n${word}\n\`\`\`\n\n`;
//     content += `$(symbol-type-parameter) **Type** \`${word}\`\n\n`;
//     content += `${description}\n\n`;
//     content += getTypeExamples(word);

//     return { content, category: 'type' };
//   }

//   return undefined;
// }
//     const value = letMatch[2].trim();

//     let hoverText = `\`\`\`zeno\nlet ${word}: ${type}\n\`\`\`\n\n`;
//     hoverText += `**Variable**: \`${word}\`\n\n`;
//     hoverText += `**Type**: \`${type}\`\n\n`;
//     hoverText += `**Value**: \`${value}\``;

//     return hoverText;
//   }

//   // 3. 関数パラメータの場合
//   for (let i = 0; i < lines.length; i++) {
//     const line = lines[i];
//     const funcDefMatch = line.match(/fn\s+\w+\s*\(([^)]*)\)/);
//     if (funcDefMatch) {
//       const params = funcDefMatch[1];
//       const paramPattern = new RegExp(`${word}\\s*:\\s*([\\w\\[\\]]+)`);
//       const paramMatch = params.match(paramPattern);
//       if (paramMatch) {
//         const type = paramMatch[1];
//         let hoverText = `\`\`\`zeno\n${word}: ${type}\n\`\`\`\n\n`;
//         hoverText += `**Parameter**: \`${word}\`\n\n`;
//         hoverText += `**Type**: \`${type}\``;
//         return hoverText;
//       }
//     }
//   }

//   // 4. 型推論による判定（より詳細）
//   const assignPattern = new RegExp(`let\\s+${word}\\s*=\\s*([^\\n;]+)`);
//   const assignMatch = text.match(assignPattern);
//   if (assignMatch) {
//     const value = assignMatch[1].trim();
//     let inferredType = '';
//     let description = '';

//     if (/^\d+$/.test(value)) {
//       inferredType = 'int';
//       description = 'Integer literal';
//     } else if (/^\d+\.\d+$/.test(value)) {
//       inferredType = 'float';
//       description = 'Floating point literal';
//     } else if (/^".*"$/.test(value)) {
//       inferredType = 'string';
//       description = 'String literal';
//     } else if (value === 'true' || value === 'false') {
//       inferredType = 'bool';
//       description = 'Boolean literal';
//     } else if (/^\[.*\]$/.test(value)) {
//       inferredType = 'array';
//       description = 'Array literal';
//     } else {
//       // 関数呼び出しの場合
//       const funcCallMatch = value.match(/^(\w+)\s*\(/);
//       if (funcCallMatch) {
//         const funcName = funcCallMatch[1];
//         const funcReturnType = getFunctionReturnType(funcName, text);
//         if (funcReturnType) {
//           inferredType = funcReturnType;
//           description = `Return value from function \`${funcName}()\``;
//         }
//       }
//     }

//     if (inferredType) {
//       let hoverText = `\`\`\`zeno\nlet ${word}: ${inferredType} // inferred\n\`\`\`\n\n`;
//       hoverText += `**Variable**: \`${word}\`\n\n`;
//       hoverText += `**Type**: \`${inferredType}\` *(inferred)*\n\n`;
//       hoverText += `**Value**: \`${value}\`\n\n`;
//       hoverText += `**Description**: ${description}`;
//       return hoverText;
//     }
//   }

//   // 5. 標準ライブラリ関数の場合（TypeScript風の詳細情報）
//   const stdFunctions = {
//     'println': {
//       signature: 'fn println(...args: any): void',
//       description: 'Prints arguments to standard output followed by a newline. Accepts any number of arguments of any type.',
//       module: 'std/fmt',
//       example: 'println("Hello", x, 42, true)',
//       returns: 'void'
//     },
//     'print': {
//       signature: 'fn print(...args: any): void',
//       description: 'Prints arguments to standard output without a trailing newline. Accepts any number of arguments of any type.',
//       module: 'std/fmt',
//       example: 'print("Value: ", x, ", Status: ", active)',
//       returns: 'void'
//     },
//     'read_line': {
//       signature: 'fn read_line(): string',
//       description: 'Reads a line from standard input',
//       module: 'std.io',
//       example: 'let input = read_line()',
//       returns: 'string'
//     },
//     'write_file': {
//       signature: 'fn write_file(path: string, content: string): bool',
//       description: 'Writes content to a file at the specified path',
//       module: 'std.io',
//       example: 'write_file("output.txt", "Hello")',
//       returns: 'bool'
//     }
//   };

//   if (stdFunctions[word as keyof typeof stdFunctions]) {
//     const func = stdFunctions[word as keyof typeof stdFunctions];
//     let hoverText = `\`\`\`zeno\n${func.signature}\n\`\`\`\n\n`;
//     hoverText += `**Function**: \`${word}\` *(${func.module})*\n\n`;
//     hoverText += `**Description**: ${func.description}\n\n`;
//     hoverText += `**Returns**: \`${func.returns}\`\n\n`;
//     hoverText += `**Example**:\n\`\`\`zeno\n${func.example}\n\`\`\``;
//     return hoverText;
//   }

//   // 6. キーワードの場合
//   const keywords = {
//     'fn': 'Function declaration keyword',
//     'let': 'Variable declaration keyword',
//     'if': 'Conditional statement keyword',
//     'else': 'Alternative branch keyword',
//     'while': 'Loop statement keyword',
//     'for': 'Iteration statement keyword',
//     'return': 'Function return keyword',
//     'pub': 'Public visibility modifier',
//     'import': 'Module import keyword',
//     'from': 'Import source specifier'
//   };

//   if (keywords[word as keyof typeof keywords]) {
//     let hoverText = `\`\`\`zeno\n${word}\n\`\`\`\n\n`;
//     hoverText += `**Keyword**: \`${word}\`\n\n`;
//     hoverText += `**Description**: ${keywords[word as keyof typeof keywords]}`;
//     return hoverText;
//   }

//   return undefined;
// }

// // 関数の戻り値型を取得
// function getFunctionReturnType(funcName: string, text: string): string | undefined {
//   const funcPattern = new RegExp(`fn\\s+${funcName}\\s*\\([^)]*\\)\\s*:\\s*([\\w\\[\\]]+)`, 'g');
//   const match = funcPattern.exec(text);
//   return match ? match[1] : undefined;
// }

// // 定義を検索する関数
// function findDefinition(word: string, document: vscode.TextDocument): vscode.Location[] {
//   const text = document.getText();
//   const lines = text.split('\n');
//   const locations: vscode.Location[] = [];

//   for (let i = 0; i < lines.length; i++) {
//     const line = lines[i];

//     // 関数定義をチェック
//     const funcRegex = new RegExp(`fn\\s+${word}\\s*\\(`);
//     if (funcRegex.test(line)) {
//       const position = new vscode.Position(i, line.indexOf(word));
//       locations.push(new vscode.Location(document.uri, position));
//     }

//     // 変数定義をチェック
//     const letRegex = new RegExp(`let\\s+${word}\\s*[=:]`);
//     if (letRegex.test(line)) {
//       const position = new vscode.Position(i, line.indexOf(word));
//       locations.push(new vscode.Location(document.uri, position));
//     }
//   }

//   return locations;
// }

// // 補完アイテムを提供する関数（TypeScript風の詳細な補完）
// function getCompletionItems(document: vscode.TextDocument, position: vscode.Position): vscode.CompletionItem[] {
//   const items: vscode.CompletionItem[] = [];

//   // Zeno言語のキーワード
//   const keywords = [
//     { name: 'fn', detail: 'Function declaration', documentation: 'Declares a new function' },
//     { name: 'let', detail: 'Variable declaration', documentation: 'Declares a new variable' },
//     { name: 'if', detail: 'Conditional statement', documentation: 'Conditional execution' },
//     { name: 'else', detail: 'Alternative branch', documentation: 'Alternative execution path' },
//     { name: 'while', detail: 'Loop statement', documentation: 'Repeats code while condition is true' },
//     { name: 'for', detail: 'Iteration statement', documentation: 'Iterates over a collection' },
//     { name: 'return', detail: 'Return statement', documentation: 'Returns a value from function' },
//     { name: 'import', detail: 'Import statement', documentation: 'Imports a module' },
//     { name: 'pub', detail: 'Public modifier', documentation: 'Makes function/variable public' }
//   ];

//   keywords.forEach(keyword => {
//     const item = new vscode.CompletionItem(keyword.name, vscode.CompletionItemKind.Keyword);
//     item.insertText = keyword.name;
//     item.detail = keyword.detail;
//     item.documentation = new vscode.MarkdownString(keyword.documentation);
//     items.push(item);
//   });

//   // 型の補完
//   const types = [
//     { name: 'int', description: 'Integer type' },
//     { name: 'string', description: 'String type' },
//     { name: 'bool', description: 'Boolean type' },
//     { name: 'float', description: 'Floating point type' },
//     { name: 'array', description: 'Array type' }
//   ];

//   types.forEach(type => {
//     const item = new vscode.CompletionItem(type.name, vscode.CompletionItemKind.TypeParameter);
//     item.insertText = type.name;
//     item.detail = `Zeno type: ${type.name}`;
//     item.documentation = new vscode.MarkdownString(type.description);
//     items.push(item);
//   });

//   // 標準ライブラリ関数（より詳細）
//   const stdFunctions = [
//     {
//       name: 'println',
//       detail: 'fn println(message: string): void',
//       doc: 'Prints a message to standard output followed by a newline\n\n**Example:**\n```zeno\nprintln("Hello, World!")\n```',
//       snippet: 'println($1)'
//     },
//     {
//       name: 'print',
//       detail: 'fn print(message: string): void',
//       doc: 'Prints a message to standard output without a trailing newline\n\n**Example:**\n```zeno\nprint("Hello, ")\n```',
//       snippet: 'print($1)'
//     },
//     {
//       name: 'read_line',
//       detail: 'fn read_line(): string',
//       doc: 'Reads a line from standard input\n\n**Example:**\n```zeno\nlet input = read_line()\n```',
//       snippet: 'read_line()'
//     },
//     {
//       name: 'write_file',
//       detail: 'fn write_file(path: string, content: string): bool',
//       doc: 'Writes content to a file at the specified path\n\n**Example:**\n```zeno\nwrite_file("output.txt", "Hello")\n```',
//       snippet: 'write_file($1, $2)'
//     }
//   ];

//   stdFunctions.forEach(func => {
//     const item = new vscode.CompletionItem(func.name, vscode.CompletionItemKind.Function);
//     item.detail = func.detail;
//     item.documentation = new vscode.MarkdownString(func.doc);
//     item.insertText = new vscode.SnippetString(func.snippet);
//     items.push(item);
//   });

//   return items;
// }
