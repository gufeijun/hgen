// 产生式见文法：bnt.txt
package parse

import (
	"fmt"
	"io/ioutil"
)

type Parser struct {
	filepath string
	lexer    *lexer
	token    *Token

	Infos *Symbols
}

func NewParser(filepath string) *Parser {
	return &Parser{
		filepath: filepath,
		Infos: &Symbols{
			Services: make(map[string]*Service),
			Messages: make(map[string]*Message),
		},
	}
}

func (p *Parser) initLexer() error {
	if p.lexer != nil {
		return nil
	}
	data, err := ioutil.ReadFile(p.filepath)
	if err != nil {
		return err
	}
	p.lexer, err = newLexer(data)
	return err
}

func (p *Parser) nextToken() {
	p.lexer.getNextToken()
}

func (p *Parser) Parse() error {
	if err := p.initLexer(); err != nil {
		return err
	}
	p.token = &p.lexer.curToken
	p.nextToken()
	p.procCode()
	p.nextToken()
	if p.token.Kind != T_EOF {
		return fmt.Errorf("syntax error: want eof at end")
	}
	return nil
}

// 非终结符Code对应的过程
func (p *Parser) procCode() {
	switch p.token.Kind {
	case T_EOF:
		return
	case T_MESSAGE:
		fallthrough
	case T_SERVICE:
		// 产生式1
		p.procStmt()
		p.nextToken()
		p.procExtra()
	case T_CRLF:
		// 产生式2
		p.procExtra()
	default:
		// TODO
	}
}

// 非终结符Stmt对应的过程
func (p *Parser) procStmt() {
	switch p.token.Kind {
	case T_MESSAGE:
		// 产生式5
		p.procMsgStmt()
	case T_SERVICE:
		// 产生式6
		p.procServiceStmt()
	default:
		// TODO
	}
}

// 非终结符Extra对应的过程
func (p *Parser) procExtra() {
	switch p.token.Kind {
	case T_CRLF:
		// 产生式3
		p.nextToken()
		p.procStmt()
		p.nextToken()
		p.procExtra()
	case T_EOF:
		// 产生式4
		return
	default:
		// TODO
	}
}

// 非终结符MsgStmt对应的过程
func (p *Parser) procMsgStmt() {
	// 产生式7
	if p.token.Kind != T_MESSAGE {
		// TODO
	}
	p.nextToken()
	if p.token.Kind != T_ID {
		// TODO
	}
	// TODO 处理ID
	p.nextToken()
	if p.token.Kind != T_LEFTBRACE {
		// TODO
	}
	p.nextToken()
	if p.token.Kind != T_CRLF {
		// TODO
	}
	p.nextToken()
	p.procMember()
	p.nextToken()
	if p.token.Kind != T_CRLF {
		// TODO
	}
	p.nextToken()
	p.procMembers()
	p.nextToken()
	if p.token.Kind != T_RIGHTBRACE {
		// TODO
	}
}

// 非终结符ServiceStmt对应的过程
func (p *Parser) procServiceStmt() {
	// 产生式11
	if p.token.Kind != T_SERVICE {
		// TODO
	}
	p.nextToken()
	if p.token.Kind != T_ID {
		// TODO
	}
	// TODO 处理ID
	p.nextToken()
	if p.token.Kind != T_LEFTBRACE {
		// TODO
	}
	p.nextToken()
	if p.token.Kind != T_CRLF {
		// TODO
	}
	p.nextToken()
	p.procFunc()
	p.nextToken()
	if p.token.Kind != T_CRLF {
		// TODO
	}
	p.nextToken()
	p.procFuncs()
	p.nextToken()
	if p.token.Kind != T_RIGHTBRACE {
		// TODO
	}
}

// 非终结符Members对应的过程
func (p *Parser) procMembers() {
	switch p.token.Kind {
	case T_ID:
		// 产生式8
		p.procMember()
		p.nextToken()
		if p.token.Kind != T_CRLF {
			// TODO
		}
		p.nextToken()
		p.procMembers()
	case T_RIGHTBRACE:
		// 产生式9
		return
	default:
		// TODO
	}
}

// 非终结符Member对应的过程
func (p *Parser) procMember() {
	// 产生式10
	if p.token.Kind != T_ID {
		// TODO
	}
	// TODO 处理ID
	p.nextToken()
	if p.token.Kind != T_ID {
		// TODO
	}
	// TODO 处理ID
}

// 非终结符Funcs对应的过程
func (p *Parser) procFuncs() {
	switch p.token.Kind {
	case T_ID:
		// 产生式12
		p.procFunc()
		p.nextToken()
		if p.token.Kind != T_CRLF {
			// TODO
		}
		p.nextToken()
		p.procFuncs()
	case T_RIGHTBRACE:
		// 产生式13
		return
	default:
		// TODO
	}
}

// 非终结符Func对应的过程
func (p *Parser) procFunc() {
	// 产生式14
	if p.token.Kind != T_ID {
		// TODO
	}
	// TODO 处理ID
	p.nextToken()
	if p.token.Kind != T_ID {
		// TODO
	}
	// TODO 处理ID
	p.nextToken()
	if p.token.Kind != T_LEFTBRACE {
		// TODO
	}
	p.nextToken()
	p.procArgList()
	p.nextToken()
	if p.token.Kind != T_RIGHTBRACKET {
		// TODO
	}
}

// 非终结符ArgList对应的过程
func (p *Parser) procArgList() {
	switch p.token.Kind {
	case T_ID:
		// 产生式15
		p.procArgs()
	case T_RIGHTBRACKET:
		// 产生式16
		return
	default:
		// TODO
	}
}

// 非终结符Args对应的过程
func (p *Parser) procArgs() {
	// 产生式17
	if p.token.Kind != T_ID {
		// TODO
	}
	// TODO 处理id
	p.nextToken()
	p.procArgs_()
}

// 非终结符Args_对应的过程
func (p *Parser) procArgs_() {
	switch p.token.Kind {
	case T_COMMA:
		// 产生式18
		p.nextToken()
		if p.token.Kind != T_ID {
			// TODO
		}
		// TODO 处理ID
		p.nextToken()
		p.procArgs_()
	case T_RIGHTBRACKET:
		// 产生式19
		return
	default:
		// TODO
	}
}
