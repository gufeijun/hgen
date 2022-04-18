package parse

// TokenKind
const (
	T_MESSAGE      = iota // message
	T_ID                  // identity
	T_LEFTBRACE           // {
	T_RIGHTBRACE          // }
	T_LEFTBRACKET         // (
	T_RIGHTBRACKET        // )
	T_SERVICE             // service
	T_CRLF                // \n or \r\n
	T_COMMA               // ,
	T_EOF
)

type Token struct {
	Kind   int
	Value  string
	Line   int // token所在的文件行
	Kth    int // 处于该行的第几个字符
	Length int // 该token的长度
}

func (t *Token) content(l *lexer) string {
	if t.Kind == T_CRLF {
		return "\\n"
	}
	return l.lines[t.Line][t.Kth : t.Kth+t.Length]
}
