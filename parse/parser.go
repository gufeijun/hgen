// 产生式见文法：bnt.txt
package parse

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
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
		return fmt.Errorf("[ERROR] size of %s cannot exceed 10MB!", p.filepath)
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

// TODO delete
func (p *Parser) Test() *lexer {
	if err := p.initLexer(); err != nil {
		panic(err)
	}
	return p.lexer
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
		p.logError(fmt.Sprintf("expect message or service, but got \"%s\"", p.token.content(p.lexer)), *p.token)
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
		p.logError(fmt.Sprintf("expect message or service, but got \"%s\"", p.token.content(p.lexer)), *p.token)
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
		p.logError(fmt.Sprintf("expect \"\\n\" after \"}\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
}

// 非终结符MsgStmt对应的过程
func (p *Parser) procMsgStmt() (*Message, Token) {
	var token Token
	// 产生式7
	if p.token.Kind != T_MESSAGE {
		p.logError(fmt.Sprintf("expect message, but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	p.nextToken()
	if p.token.Kind != T_ID {
		p.logError(fmt.Sprintf("unrecognized identity \"%s\" after message", p.token.content(p.lexer)), *p.token)
	}
	msg := &Message{Name: p.token.Value}
	token = *p.token
	tmp := *p.token
	p.nextToken()
	if p.token.Kind != T_LEFTBRACE {
		p.logError(fmt.Sprintf("expect \"{\" after \"%s\"", msg.Name), tmp)
	}
	tmp = *p.token
	p.nextToken()
	if p.token.Kind != T_CRLF {
		p.logError(fmt.Sprintf("expect \"\\n\" after \"{\", but got \"%s\"", p.token.content(p.lexer)), tmp)
	}
	p.nextToken()
	member := p.procMember()
	msg.Mems = append(msg.Mems, member)
	if p.token.Kind != T_CRLF {
		p.logError(fmt.Sprintf("expect \"\\n\" after \"%s\", but got \"%s\"", member.MemName, p.token.content(p.lexer)), *p.token)
	}
	p.nextToken()
	mems := p.procMembers()
	msg.Mems = append(msg.Mems, mems...)
	if p.token.Kind != T_RIGHTBRACE {
		p.logError(fmt.Sprintf("expect \"}\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	p.nextToken()

	return msg, token
}

// 非终结符ServiceStmt对应的过程
func (p *Parser) procServiceStmt() (*Service, Token) {
	var token Token
	// 产生式11
	if p.token.Kind != T_SERVICE {
		p.logError(fmt.Sprintf("expect \"service\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	p.nextToken()
	if p.token.Kind != T_ID {
		p.logError(fmt.Sprintf("expect identity after service, but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	token = *p.token
	srv := &Service{Name: p.token.Value}
	p.nextToken()
	if p.token.Kind != T_LEFTBRACE {
		p.logError(fmt.Sprintf("expect \"{\" after service, but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	p.nextToken()
	if p.token.Kind != T_CRLF {
		p.logError(fmt.Sprintf("expect \"\\n\" after \"{\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	p.nextToken()
	method := p.procFunc()
	srv.Methods = append(srv.Methods, method)
	if p.token.Kind != T_CRLF {
		p.logError(fmt.Sprintf("expect \"\\n\" after \")\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	p.nextToken()
	methods := p.procFuncs()
	srv.Methods = append(srv.Methods, methods...)
	if p.token.Kind != T_RIGHTBRACE {
		p.logError(fmt.Sprintf("expect \"}\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
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
			p.logError(fmt.Sprintf("expect \"\\n\" after \"%s\", but got \"%s\"", mem.MemName, p.token.content(p.lexer)), *p.token)
		}
		p.nextToken()
		mems := p.procMembers()
		members = append(members, mems...)
	case T_RIGHTBRACE:
		// 产生式9
		return nil
	default:
		p.logError(fmt.Sprintf("expect \"}\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	return members
}

// 非终结符Member对应的过程
func (p *Parser) procMember() *Member {
	// 产生式10
	if p.token.Kind != T_ID {
		p.logError(fmt.Sprintf("expect identity, but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	t := newType(p.token.Value)
	p.nextToken()
	if p.token.Kind != T_ID {
		p.logError(fmt.Sprintf("expect identity after \"%s\", but got \"%s\"", t.Name, p.token.content(p.lexer)), *p.token)
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
			p.logError(fmt.Sprintf("expect \"\\n\" after \")\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
		}
		p.nextToken()
		ms := p.procFuncs()
		methods = append(methods, ms...)
	case T_RIGHTBRACE:
		// 产生式13
		return nil
	default:
		p.logError(fmt.Sprintf("expect \"}\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	return methods
}

// 非终结符Func对应的过程
func (p *Parser) procFunc() *Method {
	method := new(Method)
	// 产生式14
	if p.token.Kind != T_ID {
		p.logError(fmt.Sprintf("expect identity, but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	method.RetType = newType(p.token.Value)
	p.nextToken()
	if p.token.Kind != T_ID {
		p.logError(fmt.Sprintf("expect identity after \"%s\", but got \"%s\"", method.RetType.Name, p.token.content(p.lexer)), *p.token)
	}
	method.Name = p.token.Value
	p.nextToken()
	if p.token.Kind != T_LEFTBRACKET {
		p.logError(fmt.Sprintf("expect \"(\" after \"%s\", but got \"%s\"", method.Name, p.token.content(p.lexer)), *p.token)
	}
	p.nextToken()
	method.ReqTypes = p.procArgList()
	if p.token.Kind != T_RIGHTBRACKET {
		p.logError(fmt.Sprintf("expect \")\" after \"%s\", but got \"%s\"", method.ReqTypes[len(method.ReqTypes)-1].Name, p.token.content(p.lexer)), *p.token)
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
		p.logError(fmt.Sprintf("expect identity or \")\" after \"(\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	return nil
}

// 非终结符Args对应的过程
func (p *Parser) procArgs() []*Type {
	var args []*Type
	// 产生式17
	if p.token.Kind != T_ID {
		p.logError(fmt.Sprintf("expect identity after \"(\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
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
			p.logError(fmt.Sprintf("expect type after \",\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
		}
		args = append(args, newType(p.token.Value))
		p.nextToken()
		args = append(args, p.procArgs_()...)
	case T_RIGHTBRACKET:
		// 产生式19
		return nil
	default:
		p.logError(fmt.Sprintf("expect \",\" or \")\", but got \"%s\"", p.token.content(p.lexer)), *p.token)
	}
	return args
}

func (p *Parser) logError(msg string, token Token) {
	fmt.Printf("%s:\n", msg)
	line := p.lexer.lines[token.Line]
	if token.Kind == T_CRLF {
		fmt.Printf("[%s:%d] %s\n", path.Base(p.filepath), p.lexer.locationMap[token.Line]+1, line)
	} else {
		fmt.Printf("[%s:%d:%d] %s", p.filepath, p.lexer.locationMap[token.Line]+1, token.Kth, line[:token.Kth])
		fmt.Printf("\033[1;31;40m%s\033[0m", line[token.Kth:token.Kth+token.Length])
		fmt.Printf("%s\n", line[token.Kth+token.Length:])
	}
	fmt.Println("compile failed! ")
	os.Exit(0)
}
