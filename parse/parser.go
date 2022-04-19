// 见文法：bnf.txt
package parse

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type Parser struct {
	filepath string
	lexer    *lexer
	token    *Token

	tmpToken *Token

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

func (p *Parser) saveService(srv *Service, token Token) {
	if _, ok := p.Infos.Services[srv.Name]; ok {
		p.logError(fmt.Sprintf("repeated service %s", srv.Name), token)
	}
	p.Infos.Services[srv.Name] = srv
}

func (p *Parser) saveMessage(msg *Message, token Token) {
	if _, ok := p.Infos.Messages[msg.Name]; ok {
		p.logError(fmt.Sprintf("repeated message %s", msg.Name), token)
	}
	p.Infos.Messages[msg.Name] = msg
}

func (p *Parser) initLexer() error {
	if p.lexer != nil {
		return nil
	}
	info, err := os.Stat(p.filepath)
	if err != nil {
		return err
	}
	if size := info.Size(); size >= (10 << 20) {
		return fmt.Errorf("# size of %s cannot exceed 10MB!", p.filepath)
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
		p.Panic1("message|service", "")
	}
}

// 非终结符Stmt对应的过程
func (p *Parser) procStmt() {
	switch p.token.Kind {
	case T_MESSAGE:
		// 产生式5
		msg, token := p.procMsgStmt()
		p.saveMessage(msg, token)
	case T_SERVICE:
		// 产生式6
		srv, token := p.procServiceStmt()
		p.saveService(srv, token)
	default:
		p.Panic1("message|service", "")
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
		p.Panic1(`\n`, "}")
	}
}

// 非终结符MsgStmt对应的过程
func (p *Parser) procMsgStmt() (*Message, Token) {
	var token Token
	// 产生式7
	if p.token.Kind != T_MESSAGE {
		p.Panic1(`\n`, "")
	}
	p.nextToken()
	if p.token.Kind != T_ID {
		p.Panic1("message name", "message")
	}
	msg := &Message{Name: p.token.Value}
	token = *p.token
	tmp1 := *p.token
	p.tmpToken = &tmp1 // 暂存此token，方便后面的错误处理
	p.nextToken()
	if p.token.Kind != T_LEFTBRACE {
		p.Panic2("{", msg.Name, tmp1)
	}
	tmp2 := *p.token
	p.nextToken()
	if p.token.Kind != T_CRLF {
		if p.token.Kind == T_RIGHTBRACE {
			p.logError(fmt.Sprintf("message \"%s\" should have at least one member", tmp1.Value), tmp1)
		}
		p.Panic2(`\n`, "{", tmp2)
	}
	p.nextToken()
	member := p.procMember()
	msg.Mems = append(msg.Mems, member)
	if p.token.Kind != T_CRLF {
		p.Panic1(`\n`, member.MemName)
	}
	p.nextToken()
	mems := p.procMembers()
	msg.Mems = append(msg.Mems, mems...)
	if p.token.Kind != T_RIGHTBRACE {
		p.Panic1(`}`, "")
	}
	p.nextToken()

	return msg, token
}

// 非终结符ServiceStmt对应的过程
func (p *Parser) procServiceStmt() (*Service, Token) {
	var token Token
	// 产生式11
	if p.token.Kind != T_SERVICE {
		p.Panic1("service", "")
	}
	p.nextToken()
	if p.token.Kind != T_ID {
		p.Panic1("service name", "service")
	}
	token = *p.token
	p.tmpToken = &token
	srv := &Service{Name: p.token.Value}
	p.nextToken()
	if p.token.Kind != T_LEFTBRACE {
		p.Panic1("{", srv.Name)
	}
	p.nextToken()
	if p.token.Kind != T_CRLF {
		if p.token.Kind == T_RIGHTBRACE {
			p.logError(fmt.Sprintf("service \"%s\" should have at least one method", token.Value), token)
		}
		p.Panic1(`\n`, "{")
	}
	p.nextToken()
	method := p.procFunc()
	srv.Methods = append(srv.Methods, method)
	if p.token.Kind != T_CRLF {
		p.Panic1(`\n`, ")")
	}
	p.nextToken()
	methods := p.procFuncs()
	srv.Methods = append(srv.Methods, methods...)
	if p.token.Kind != T_RIGHTBRACE {
		p.Panic1("}", "")
	}
	for _, method := range srv.Methods {
		method.Service = srv
	}
	p.nextToken()
	return srv, token
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
			p.Panic1(`\n`, mem.MemName)
		}
		p.nextToken()
		mems := p.procMembers()
		members = append(members, mems...)
	case T_RIGHTBRACE:
		// 产生式9
		return nil
	default:
		p.Panic1("}", "")
	}
	return members
}

// 非终结符Member对应的过程
func (p *Parser) procMember() *Member {
	// 产生式10
	if p.token.Kind != T_ID {
		p.logError(fmt.Sprintf("message \"%s\" should have at least one member", p.tmpToken.Value), *p.tmpToken)
	}
	t := newType(p.token.Value)
	p.nextToken()
	if p.token.Kind != T_ID {
		p.Panic1("member name", t.Name)
	}
	name := p.token.Value
	p.nextToken()
	return &Member{
		MemType: t,
		MemName: name,
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
			p.Panic1(`\n`, ")")
		}
		p.nextToken()
		ms := p.procFuncs()
		methods = append(methods, ms...)
	case T_RIGHTBRACE:
		// 产生式13
		return nil
	default:
		p.Panic1("}", "")
	}
	return methods
}

// 非终结符Func对应的过程
func (p *Parser) procFunc() *Method {
	method := new(Method)
	// 产生式14
	if p.token.Kind != T_ID {
		p.logError(fmt.Sprintf("service \"%s\" should have at least one method", p.tmpToken.Value), *p.tmpToken)
	}
	method.RetType = newType(p.token.Value)
	p.nextToken()
	if p.token.Kind != T_ID {
		p.Panic1("function name", method.RetType.Name)
	}
	method.Name = p.token.Value
	p.nextToken()
	if p.token.Kind != T_LEFTBRACKET {
		p.Panic1("(", method.Name)
	}
	p.nextToken()
	method.ReqTypes = p.procArgList()
	if p.token.Kind != T_RIGHTBRACKET {
		p.Panic1(")", method.ReqTypes[len(method.ReqTypes)-1].Name)
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
		p.Panic1("type or )", "(")
	}
	return nil
}

// 非终结符Args对应的过程
func (p *Parser) procArgs() []*Type {
	var args []*Type
	// 产生式17
	if p.token.Kind != T_ID {
		p.Panic1("type", "(")
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
			p.Panic1("type", ",")
		}
		args = append(args, newType(p.token.Value))
		p.nextToken()
		args = append(args, p.procArgs_()...)
	case T_RIGHTBRACKET:
		// 产生式19
		return nil
	default:
		p.Panic1(", or )", "")
	}
	return args
}

func (p *Parser) logError(msg string, token Token) {
	fmt.Printf("%s:\n", msg)
	line := p.lexer.lines[token.Line]
	lineNum := p.lexer.locationMap[token.Line] + 1
	filepath := path.Base(p.filepath)
	if token.Kind == T_CRLF {
		fmt.Printf("[%s:%d] %s\n", filepath, lineNum, line)
	} else {
		fmt.Printf("[%s:%d:%d] %s", filepath, lineNum, token.Kth, line[:token.Kth])
		// 标红错误的token
		fmt.Printf("\033[1;37;41m%s\033[0m", line[token.Kth:token.Kth+token.Length])
		fmt.Printf("%s\n", line[token.Kth+token.Length:])
	}
	fmt.Println("compile failed! ")
	os.Exit(0)
}

func (p *Parser) Panic(expect, after, got string, token Token) {
	var b strings.Builder
	fmt.Fprintf(&b, "expect \"%s\"", expect)
	if len(after) != 0 {
		fmt.Fprintf(&b, " after \"%s\", ", after)
	} else {
		fmt.Fprintf(&b, ", ")
	}
	fmt.Fprintf(&b, "but got \"%s\"", got)
	p.logError(b.String(), token)
}

func (p *Parser) Panic1(expect, after string) {
	p.Panic(expect, after, p.token.content(p.lexer), *p.token)
}

func (p *Parser) Panic2(expect, after string, token Token) {
	p.Panic(expect, after, p.token.content(p.lexer), token)
}
