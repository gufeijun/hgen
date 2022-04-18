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
	T_SEMICOLON           // ;
	T_EOF
)

type Token struct {
	Kind  int
	Value string
	Line  int // token所在的文件行
}
