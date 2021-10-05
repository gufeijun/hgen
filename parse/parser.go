package parse

import (
	"bufio"
	"bytes"
	"fmt"
	"gufeijun/hustgen/service"
	"io"
	"os"
)

const (
	inNone = iota
	inService
	inMessage
)

type Parser struct {
	bufr       *bufio.Reader
	filename   string
	curService *service.Service
	curMethod  *service.Method
	nowParsing int //inService or inMessage
	curMessage *service.Message
	curLine    int
	status     int
	automaton  map[int][]int
}

func NewParser(filename string) *Parser {
	return &Parser{
		filename:  filename,
		automaton: automaton,
		status:    start,
	}
}

func (p *Parser) Parse() error {
	file, err := os.Open(p.filename)
	if err != nil {
		return err
	}
	defer file.Close()
	p.bufr = bufio.NewReader(file)
	if err = p.parse(); err != nil {
		return err
	}
	return newChecker().check()
}

func isValidChars(p []byte) (int, bool) {
	for i, ch := range p {
		if !(ch <= 'z' && ch >= 'a' || ch <= 'Z' && ch >= 'A') &&
			ch != '{' && ch != '}' && ch != ' ' &&
			!(ch <= '9' && ch >= '0') {
			return i, false
		}
	}
	return 0, true
}

func (p *Parser) parse() (err error) {
	p.curLine = 0
	for {
		line, _, err := p.bufr.ReadLine()
		if err != nil {
			if err == io.EOF {
				err = nil
			}
			return err
		}
		p.curLine++
		if len(line) == 0 {
			continue
		}
		index := bytes.Index(line, []byte(`//`))
		if index != -1 {
			line = line[:index]
		}
		for i := 0; i < len(line); i++ {
			// we treat Parentheses and comma as white space to facilitate grammar analysis
			if line[i] == '(' || line[i] == ')' || line[i] == ',' || line[i] == '\t' {
				line[i] = ' '
			}
		}
		if i, ok := isValidChars(line); !ok {
			return fmt.Errorf("[line %d]invalid character: %c", p.curLine, line[i])
		}
		if err = p.parseLine(line); err != nil {
			return fmt.Errorf("[line %d]%v", p.curLine, err)
		}
	}
}

func nextWord(line *[]byte) (string, error) {
	i := 0
	for i < len(*line) && (*line)[i] == ' ' {
		i++
	}
	if i == len(*line) {
		return "", nil
	}
	*line = (*line)[i:]
	i = 0
	for i < len(*line) && (*line)[i] != ' ' {
		i++
	}
	str := string((*line)[:i])
	*line = (*line)[i:]
	return str, nil
}

func (p *Parser) parseLine(line []byte) error {
	for {
		word, err := nextWord(&line)
		if err != nil {
			return err
		}
		col := getCol(word)
		p.status = p.automaton[p.status][col]
		breaking, err := stateFuncMap[p.status](p, word, &line)
		if err != nil {
			return err
		}
		if breaking {
			return nil
		}
	}
}
