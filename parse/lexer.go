package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

// TODO 文件一行行处理和Token处理同步
type lexer struct {
	srcCode  []byte
	cursor   int
	curChar  byte
	curToken Token
	curLine  int
	curKth   int

	locationMap []int
	lines       []string
}

// TODO delete
func (l *lexer) GetNextToken() Token {
	l.getNextToken()
	return l.curToken
}

func newLexer(code []byte) (*lexer, error) {
	var i int
	for i = len(code) - 1; i >= 0 && (code[i] == ' ' || code[i] == '\r' || code[i] == '\n' || code[i] == '\t'); i-- {
	}
	// 去除结尾的空白
	code = code[:i+1]
	l := &lexer{
		srcCode: code,
		curKth:  -1,
	}
	// 去除注释和多余的空白
	if err := l.preHandleCode(); err != nil {
		return nil, err
	}
	return l, nil
}

// 代码预处理
// 去除注释和多余的换行
func (l *lexer) preHandleCode() error {
	buff := bufio.NewReader(bytes.NewBuffer(l.srcCode))
	var writeTo bytes.Buffer
	for curLine := 0; ; curLine++ {
		line, err := readLine(buff)
		if err != nil && err != io.EOF {
			return err
		}
		index := bytes.Index(line, []byte("//"))
		if index != -1 {
			// 去除注释
			line = line[:index]
		}
		if index := bytes.Index(line, []byte("/")); index != -1 {
			return fmt.Errorf("syntax error: expect // at %dth line", curLine+1)
		}
		// 如果为空行，则跳过
		if len(line) == 0 || emptyLine(line) {
			if err == io.EOF {
				break
			}
			continue
		}
		writeTo.Write(line)
		writeTo.Write([]byte("\n"))
		// 记录新代码行号到旧代码行号的映射
		l.locationMap = append(l.locationMap, curLine)
		l.lines = append(l.lines, string(line))
		if err == io.EOF {
			break
		}
	}
	l.srcCode = writeTo.Bytes()
	if len(l.srcCode) != 0 && l.srcCode[len(l.srcCode)-1] == '\n' {
		l.srcCode = l.srcCode[:len(l.srcCode)-1]
	}
	return nil
}

func readLine(buff *bufio.Reader) ([]byte, error) {
	line, isPrefix, err := buff.ReadLine()
	if err != nil {
		return line, err
	}
	var p []byte
	for isPrefix {
		p, isPrefix, err = buff.ReadLine()
		if err != nil {
			if err == io.EOF {
				line = append(line, p...)
				return line, err
			}
			return nil, err
		}
		line = append(line, p...)
	}
	return line, nil
}

func emptyLine(line []byte) bool {
	for i := 0; i < len(line); i++ {
		if line[i] != ' ' && line[i] != '\t' {
			return false
		}
	}
	return true
}

func (l *lexer) getNextToken() {
	for l.curChar == 0 || l.curChar == ' ' || l.curChar == '\t' {
		if l.cursor >= len(l.srcCode) {
			l.curToken.Kind = T_EOF
			return
		}
		l.getNextChar()
	}
	// start:
	ch := l.curChar
	l.curToken.Kth = l.curKth
	switch ch {
	case '0':
		l.curToken.Kind = T_EOF
	case '(':
		l.curToken.Kind = T_LEFTBRACKET
		l.curToken.Length = 1
	case ')':
		l.curToken.Kind = T_RIGHTBRACKET
		l.curToken.Length = 1
	case '{':
		l.curToken.Kind = T_LEFTBRACE
		l.curToken.Length = 1
	case '}':
		l.curToken.Kind = T_RIGHTBRACE
		l.curToken.Length = 1
	// case '\r':
	// 	l.getNextChar()
	// 	if l.curChar != '\n' {
	// 		fmt.Printf("expect \\n at end of %dth line\n", l.curLine)
	// 		os.Exit(1)
	// 	}
	// 	fallthrough
	case '\n':
		l.curToken.Kind = T_CRLF
		defer func() { l.curLine++ }()
	case ',':
		l.curToken.Kind = T_COMMA
		l.curToken.Length = 1
	// case '/':
	// 	// 处理注释
	// 	l.getNextChar()
	// 	if l.curChar != '/' {
	// 		fmt.Printf("expect // in %dth line\n", l.curLine)
	// 		os.Exit(1)
	// 	}
	// 	for l.curChar != '\r' && l.curChar != '\n' {
	// 		l.getNextChar()
	// 	}
	// 	goto start
	default:
		if !isLetter_(ch) {
			l.logError()
		}
		id := string(ch)
		for {
			l.getNextChar()
			ch = l.curChar
			if !isLetter_(ch) && !isNumber(ch) {
				break
			}
			id += string(ch)
		}
		l.curToken.Length = len(id)
		if id == "message" {
			l.curToken.Kind = T_MESSAGE
		} else if id == "service" {
			l.curToken.Kind = T_SERVICE
		} else {
			l.curToken.Kind = T_ID
			l.curToken.Value = id
		}
		goto end
	}
	l.getNextChar()
end:
	l.curToken.Line = l.curLine
}

func (l *lexer) logError() {
	fmt.Printf("%dth line lexer failed: invalid character %c\n", l.curLine, l.curChar)
	os.Exit(0)
}

func isNumber(ch byte) bool {
	return ch <= '9' && ch >= '0'
}

func isLetter_(ch byte) bool {
	return ch <= 'z' && ch >= 'a' || ch >= 'A' && ch <= 'Z' || ch == '_'
}

func (l *lexer) getNextChar() {
	if l.cursor >= len(l.srcCode) {
		l.curChar = 0
		return
	}
	l.curChar = l.srcCode[l.cursor]
	l.cursor++
	if l.curChar == '\n' {
		l.curKth = -1
	} else {
		l.curKth++
	}
}
