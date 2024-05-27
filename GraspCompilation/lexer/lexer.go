// lexer/lexer.go
package lexer

import "jieshiqi/token"

// 词法分析器
type Lexer struct {
	input        string // 当前分析的字符串
	position     int    // 所输入字符串中的当前位置（指向当前字符）
	readPosition int    // 所输入字符串中的当前读取位置（指向当前字符之后的一个字符）
	ch           byte   // 当前正在查看的字符
}

// New 使用readChar， 初始化 l.ch、 l.position 和
// l.readPosition，以便在调用NextToken()之前让*Lexer完全就绪：
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

// readChar的目的是读取input中的下一个字符，并前移其在input中的位置。这个
// 过程的第一件事就是检查是否已经到达input的末尾。如果是，则将l.ch设置为0，这是
// NUL字符的ASCII编码，用来表示“尚未读取任何内容”或“文件结尾”。如果还没有到达
// input的末尾，则将l.ch设置为下一个字符，即l.input[l.readPosition]指向的字
// 符。
// 之 后， 将 l.position 更 新 为 刚 用 过 的 l.readPosition， 然 后 将
// l.readPosition加1。这样一来，l.readPosition就始终指向下一个将读取的字符位
// 置，而l.position始终指向刚刚读取的位置。
// 该词法分析器仅支持ASCII字符，不能支持所有
// 的Unicode字符。这么做也是为了让事情保持简单，让我们能够专注于解释器的基础部分。
// 如果要完全支持Unicode和UTF-8，就要将l.ch的类型从byte改为rune，同时还要修改读
// 取下一个字符的方式。因为字符此时可能为多字节，所以l.input[l.readPosition]
// 将无法工作。除此之外，还需要修改其他一些后面会介绍的方法和函数。
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) { // 读到了最后，直接清零
		l.ch = 0
	} else { // 没有读到最后，让ch为输入字符串的下一个字符
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition += 1
}

// NextToken 它首先检查了当前正在查看的字符l.ch，根
// 据具体的字符来返回对应的词法单元。在返回词法单元之前，位于所输入字符串中的指针
// 会前移，所以之后再次调用NextToken()时，l.ch字段就已经更新过了。最后，名为
// newToken的小型函数可以帮助初始化这些词法单元。
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()
	switch l.ch {
	case '=':
		tok = newToken(token.ASSIGN, l.ch)
	case '+':
		tok = newToken(token.PLUS, l.ch)
	case '-':
		tok = newToken(token.MINUS, l.ch)
	case '!':
		tok = newToken(token.BANG, l.ch)
	case '/':
		tok = newToken(token.SLASH, l.ch)
	case '*':
		tok = newToken(token.ASTERISK, l.ch)
	case '<':
		tok = newToken(token.LT, l.ch)
	case '>':
		tok = newToken(token.GT, l.ch)
	case ';':
		tok = newToken(token.SEMICOLON, l.ch)
	case ',':
		tok = newToken(token.COMMA, l.ch)
	case '(':
		tok = newToken(token.LPAREN, l.ch)
	case ')':
		tok = newToken(token.RPAREN, l.ch)
	case '{':
		tok = newToken(token.LBRACE, l.ch)
	case '}':
		tok = newToken(token.RBRACE, l.ch)
	case 0:
		tok.Literal = ""
		tok.Type = token.EOF
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdent(tok.Literal)
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			return tok
		} else {
			tok = newToken(token.ILLEGAL, l.ch)
		}
	}
	l.readChar()
	return tok
}

func newToken(tokenType token.TokenType, ch byte) token.Token {
	return token.Token{Type: tokenType, Literal: string(ch)}
}

// readIdentifier 函数顾名思义，就是读入一个标识符并前移词法分析器的扫描位
// 置，直到遇见非字母字符。
// 在switch语句的default分支中，使用readIdentifier()设置了当前词法单元的
// Literal字段
func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) { //如果读到的一直是字母或者下划线就持续往下读，然后然会一个字符串，这个字符串就是整个变量名
		l.readChar()
	}
	return l.input[position:l.position]
}

// isLetter 辅助函数用来判断给定的参数是否为字母。值得注意的是，这个函数虽然
// 看起来简短，但意义重大，其决定了解释器所能处理的语言形式。比如示例中包含ch
// =='_'，这意味着下划线_会被视为字母，允许在标识符和关键字中使用。因此可以使用
// 诸如foo_bar之类的变量名。其他编程语言甚至允许在标识符中使用问号和感叹号。如果
// 读者也想这么做，那么可以修改这个isLetter函数。
func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

// skipWhitespace 很多分析器中有这个简单的辅助函数，它有时称为eatWhitespace，有时称为
// consumeWhiteSpace，还有时是完全不同的名称。这个函数实际跳过的字符根据具体分
// 析的语言会有所不同。例如在某些语言的实现中，会为换行符创建词法单元，如果它们不
// 在词法单元流中的正确位置，就会抛出解析错误。不过这里没有处理换行符，是为了简化
// 后面的语法分析步骤。
func (l *Lexer) skipWhitespace() { //eatWhitespace，consumeWhiteSpace
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func (l *Lexer) readNumber() string {
	position := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[position:l.position]
}
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
