package parse

import (
	"fmt"
	"io/ioutil"
)

type Parser struct {
	filepath string
	lexer    *lexer
	token    *Token
}

func NewParser(filepath string) *Parser {
	return &Parser{
		filepath: filepath,
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
	p.lexer, err = NewLexer(data)
	return err
}

func (p *Parser) nextToken() {
	p.lexer.GetNextToken()
}

func (p *Parser) Parse() error {
	if err := p.initLexer(); err != nil {
		return err
	}
	p.token = &p.lexer.curToken
	p.nextToken()
	// TODO parser

	p.nextToken()
	if p.token.Kind != T_EOF {
		return fmt.Errorf("syntax error: want eof at end")
	}
	return nil
}
