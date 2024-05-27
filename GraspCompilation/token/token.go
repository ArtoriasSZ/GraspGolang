// token/token.go
package token

type TokenType string

const (
	ILLEGAL = "ILLEGAL" // 非法
	EOF     = "EOF"     // 文件结束
	// 标识符+字面量
	IDENT = "IDENT" // add, foobar, x, y, ...用户自定义标识符
	INT   = "INT"   // 1343456 整数
	// 运算符
	ASSIGN   = "="
	PLUS     = "+"
	MINUS    = "-"
	BANG     = "!"
	ASTERISK = "*"
	SLASH    = "/"
	LT       = "<"
	GT       = ">"
	// 分隔符
	COMMA     = "," // 逗号
	SEMICOLON = ";" // 分号
	LPAREN    = "(" // left paren
	RPAREN    = ")" // right paren
	LBRACE    = "{" // left brace
	RBRACE    = "}"
	// 关键字
	FUNCTION = "FUNCTION"
	LET      = "LET"
)

type Token struct {
	Type    TokenType
	Literal string
}

// 关键字map
var keywords = map[string]TokenType{
	"fn":  FUNCTION,
	"let": LET,
}

// LookupIdent 通过检查关键字表来判断给定的标识符是否是关键字。如果是，则返
// 回关键字的TokenType常量。如果不是，则返回token.IDENT，这个TokenType表示当
// 前是用户定义的标识符。
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}
