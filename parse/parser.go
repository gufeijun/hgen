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

func (p *Parser) saveService(srv *Service) {
	if _, ok := p.Infos.Services[srv.Name]; ok {
		// TODO panic
	}
	p.Infos.Services[srv.Name] = srv
}

func (p *Parser) saveMessage(msg *Message) {
	if _, ok := p.Infos.Messages[msg.Name]; ok {
		// TODO panic
	}
	p.Infos.Messages[msg.Name] = msg
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
		p.procExtra()
	case T_CRLF:
		// 产生式2
		p.procExtra()
	default:
		// TODO
		panic(nil)
	}
}

// 非终结符Stmt对应的过程
func (p *Parser) procStmt() {
	switch p.token.Kind {
	case T_MESSAGE:
		// 产生式5
		msg := p.procMsgStmt()
		p.saveMessage(msg)
	case T_SERVICE:
		// 产生式6
		srv := p.procServiceStmt()
		p.saveService(srv)
	default:
		// TODO
		panic(nil)
	}
}

// 非终结符Extra对应的过程
func (p *Parser) procExtra() {
	switch p.token.Kind {
	case T_CRLF:
		// 产生式3
		p.nextToken()
		p.procStmt()
		p.procExtra()
	case T_EOF:
		// 产生式4
		return
	default:
		// TODO
		panic(nil)
	}
}

// 非终结符MsgStmt对应的过程
func (p *Parser) procMsgStmt() *Message {
	// 产生式7
	if p.token.Kind != T_MESSAGE {
		// TODO
		panic(nil)
	}
	p.nextToken()
	if p.token.Kind != T_ID {
		// TODO
		panic(nil)
	}
	msg := &Message{Name: p.token.Value}
	p.nextToken()
	if p.token.Kind != T_LEFTBRACE {
		// TODO
		panic(nil)
	}
	p.nextToken()
	if p.token.Kind != T_CRLF {
		// TODO
		panic(nil)
	}
	p.nextToken()
	member := p.procMember()
	msg.Mems = append(msg.Mems, member)
	if p.token.Kind != T_CRLF {
		// TODO
		panic(nil)
	}
	p.nextToken()
	mems := p.procMembers()
	msg.Mems = append(msg.Mems, mems...)
	if p.token.Kind != T_RIGHTBRACE {
		// TODO
		panic(nil)
	}
	p.nextToken()

	return msg
}

// 非终结符ServiceStmt对应的过程
func (p *Parser) procServiceStmt() *Service {
	// 产生式11
	if p.token.Kind != T_SERVICE {
		// TODO
		panic(nil)
	}
	p.nextToken()
	if p.token.Kind != T_ID {
		// TODO
		panic(nil)
	}
	srv := &Service{Name: p.token.Value}
	p.nextToken()
	if p.token.Kind != T_LEFTBRACE {
		// TODO
		panic(nil)
	}
	p.nextToken()
	if p.token.Kind != T_CRLF {
		// TODO
		panic(nil)
	}
	p.nextToken()
	method := p.procFunc()
	srv.Methods = append(srv.Methods, method)
	if p.token.Kind != T_CRLF {
		// TODO
		panic(nil)
	}
	p.nextToken()
	methods := p.procFuncs()
	srv.Methods = append(srv.Methods, methods...)
	if p.token.Kind != T_RIGHTBRACE {
		// TODO
		panic(nil)
	}
	for _, method := range srv.Methods {
		method.Service = srv
	}
	p.nextToken()
	return srv
}

// 非终结符Members对应的过程
func (p *Parser) procMembers() []*Member {
	var members []*Member
	switch p.token.Kind {
	case T_ID:
		// 产生式8
		mem := p.procMember()
		members = append(members, mem)
		if p.token.Kind != T_CRLF {
			// TODO
			panic(nil)
		}
		p.nextToken()
		mems := p.procMembers()
		members = append(members, mems...)
	case T_RIGHTBRACE:
		// 产生式9
		return nil
	default:
		// TODO
		panic(nil)
	}
	return members
}

// 非终结符Member对应的过程
func (p *Parser) procMember() *Member {
	// 产生式10
	if p.token.Kind != T_ID {
		// TODO
		panic(nil)
	}
	t := newType(p.token.Value)
	p.nextToken()
	if p.token.Kind != T_ID {
		// TODO
		panic(nil)
	}
	p.nextToken()
	return &Member{
		MemType: t,
		MemName: p.token.Value,
	}
}

// 非终结符Funcs对应的过程
func (p *Parser) procFuncs() []*Method {
	var methods []*Method
	switch p.token.Kind {
	case T_ID:
		// 产生式12
		method := p.procFunc()
		methods = append(methods, method)
		if p.token.Kind != T_CRLF {
			// TODO
			panic(nil)
		}
		p.nextToken()
		ms := p.procFuncs()
		methods = append(methods, ms...)
	case T_RIGHTBRACE:
		// 产生式13
		return nil
	default:
		// TODO
		panic(nil)
	}
	return methods
}

// 非终结符Func对应的过程
func (p *Parser) procFunc() *Method {
	method := new(Method)
	// 产生式14
	if p.token.Kind != T_ID {
		// TODO
		panic(nil)
	}
	method.RetType = newType(p.token.Value)
	p.nextToken()
	if p.token.Kind != T_ID {
		// TODO
		panic(nil)
	}
	method.Name = p.token.Value
	p.nextToken()
	if p.token.Kind != T_LEFTBRACKET {
		// TODO
		panic(nil)
	}
	p.nextToken()
	method.ReqTypes = p.procArgList()
	if p.token.Kind != T_RIGHTBRACKET {
		// TODO
		panic(nil)
	}
	p.nextToken()
	return method
}

// 非终结符ArgList对应的过程
func (p *Parser) procArgList() []*Type {
	switch p.token.Kind {
	case T_ID:
		// 产生式15
		return p.procArgs()
	case T_RIGHTBRACKET:
		// 产生式16
		return nil
	default:
		// TODO
		panic(nil)
	}
	return nil
}

// 非终结符Args对应的过程
func (p *Parser) procArgs() []*Type {
	var args []*Type
	// 产生式17
	if p.token.Kind != T_ID {
		// TODO
		panic(nil)
	}
	args = append(args, newType(p.token.Value))

	p.nextToken()
	args = append(args, p.procArgs_()...)
	return args
}

// 非终结符Args_对应的过程
func (p *Parser) procArgs_() []*Type {
	var args []*Type
	switch p.token.Kind {
	case T_COMMA:
		// 产生式18
		p.nextToken()
		if p.token.Kind != T_ID {
			// TODO
			panic(nil)
		}
		args = append(args, newType(p.token.Value))
		p.nextToken()
		args = append(args, p.procArgs_()...)
	case T_RIGHTBRACKET:
		// 产生式19
		return nil
	default:
		// TODO
		panic(nil)
	}
	return args
}
