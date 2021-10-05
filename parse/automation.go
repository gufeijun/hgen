package parse

import (
	"errors"
	"fmt"
	"gufeijun/hustgen/service"
)

const (
	start = iota
	in_service
	read_method
	in_retType
	method_name
	in_acpType
	in_message
	read_mem
	msg_mem_typ
	msg_mem_name
	bad
)

/* following is our automaton
--------------------------------------------------------------------------
            |     ""    | "service" | "message" |    "}"    |   others   |       description
--------------------------------------------------------------------------
start       |   start   |in_service |in_message |   bad     |   bad      |  0   should read next line
--------------------------------------------------------------------------
in_service  |read_method|   bad     |   bad     |   bad     |   bad      |  1   read service name
--------------------------------------------------------------------------
read_method |read_method|   bad     |   bad     |   start   |in_retType  |  2   ready to read a method
--------------------------------------------------------------------------
in_retType  |   bad     |   bad     |   bad     |   bad     |method_name |  3   read return type of current method
--------------------------------------------------------------------------
method_name |   bad     |   bad     |   bad     |   bad     |in_acpType  |  4   read method name
--------------------------------------------------------------------------
in_acpType  |read_method|   bad     |   bad     |   bad     |in_acpType  |  5   read parameters of current method
--------------------------------------------------------------------------
in_message  |read_mem   |   bad     |   bad     |   bad     |   bad      |  6   read message name
--------------------------------------------------------------------------
read_mem    |read_mem   |   bad     |   bad     |   start   |msg_mem_typ |  7   ready to read members of current message
--------------------------------------------------------------------------
msg_mem_typ |   bad     |   bad     |   bad     |   bad     |msg_mem_name|  8   read type of the member
--------------------------------------------------------------------------
msg_mem_name|read_mem   |   bad     |   bad     |   bad     |   bad      |  9   read name of the member
--------------------------------------------------------------------------
bad         |   bad     |   bad     |   bad     |   bad     |   bad      |  10  error happens
--------------------------------------------------------------------------
*/

// this tells how to change our state when encountering diffrent situations
var automaton = map[int][]int{
	start:        []int{start, in_service, in_message, bad, bad},
	in_service:   []int{read_method, bad, bad, bad, bad},
	read_method:  []int{read_method, bad, bad, start, in_retType},
	in_retType:   []int{bad, bad, bad, bad, method_name},
	method_name:  []int{bad, bad, bad, bad, in_acpType},
	in_acpType:   []int{read_method, bad, bad, bad, in_acpType},
	in_message:   []int{read_mem, bad, bad, bad, bad},
	read_mem:     []int{read_mem, bad, bad, start, msg_mem_typ},
	msg_mem_typ:  []int{bad, bad, bad, bad, msg_mem_name},
	msg_mem_name: []int{read_mem, bad, bad, bad, bad},
	bad:          []int{bad, bad, bad, bad, bad},
}

// each state corresponds to a processing function which contains three arguments and two return values.
// parser: parser is the recorder of state.
// word: we analyze our grammar with each word as the smallest unit.
// line: we read the file line by line and extract words on it. we might change the length of line so it's a pointer of []byte.
// return values: bool indicates if we need to read next line. error indicates if any error happends.
// whenever the state changes, the function will be called automatically.
var stateFuncMap = map[int]func(parser *Parser, word string, line *[]byte) (bool, error){
	start:        stateStart,
	in_service:   stateInService,
	read_method:  stateReadMethod,
	in_retType:   stateInRetType,
	method_name:  stateMethodName,
	in_acpType:   stateInAcpType,
	in_message:   stateInMessage,
	read_mem:     stateReadMem,
	msg_mem_typ:  stateMsgMemType,
	msg_mem_name: stateMsgMemName,
	bad:          stateBad,
}

func stateStart(parser *Parser, word string, line *[]byte) (bool, error) {
	// if current state is tranformed from read_method or read_mem, it indicates
	// that a message or service is completely read and we need to save it
	if word == "}" {
		if parser.nowParsing == inMessage {
			service.SaveMessage(parser.curMessage)
		} else {
			service.SaveService(parser.curService)
		}
	}
	parser.nowParsing = inNone
	return true, nil
}

func stateInService(parser *Parser, word string, line *[]byte) (breaking bool, err error) {
	parser.nowParsing = inService
	w, err := nextWord(line)
	if err != nil {
		return
	}
	if len(w) <= 1 {
		err = errors.New("invalid service name")
		return
	}
	end := len(w) - 1
	if w[len(w)-1] != '{' {
		var ww string
		ww, err = nextWord(line)
		if err != nil {
			return
		}
		if ww != "{" {
			err = fmt.Errorf("can not find { after service:%s", w)
			return
		}
		end++
	}
	parser.curService = service.NewService(w[:end])
	return
}

func stateReadMethod(parser *Parser, word string, line *[]byte) (bool, error) {
	if parser.curMethod != nil {
		parser.curService.AddMethod(parser.curMethod)
		parser.curMethod = nil
	}
	return true, nil
}

func stateInRetType(parser *Parser, word string, line *[]byte) (breaking bool, err error) {
	_type := new(service.Type)
	parser.curMethod = &service.Method{
		Service: parser.curService,
		RetType: _type,
	}
	_type.TypeName = word
	_type.TypeKind = GetTypeKind(word)
	return false, nil
}

func stateMethodName(parser *Parser, word string, line *[]byte) (bool, error) {
	parser.curMethod.MethodName = word
	return false, nil
}

func stateInAcpType(parser *Parser, word string, line *[]byte) (breaking bool, err error) {
	_type := new(service.Type)
	parser.curMethod.ReqTypes = append(parser.curMethod.ReqTypes, _type)
	_type.TypeName = word
	return false, nil
}

func stateInMessage(parser *Parser, word string, line *[]byte) (breaking bool, err error) {
	parser.nowParsing = inMessage
	w, err := nextWord(line)
	if err != nil {
		return
	}
	if len(w) <= 1 {
		err = errors.New("invalid message name")
		return
	}
	if w[len(w)-1] != '{' {
		err = fmt.Errorf("can not find { after message:%s", w)
		return
	}
	parser.curMessage = &service.Message{
		Name: w[:len(w)-1],
	}
	return
}

func stateReadMem(parser *Parser, word string, line *[]byte) (breaking bool, err error) {
	return true, nil
}

func stateMsgMemType(parser *Parser, word string, line *[]byte) (bool, error) {
	_type := new(service.Type)
	_type.TypeName = word
	_type.TypeKind = GetTypeKind(word)
	parser.curMessage.Mems = append(parser.curMessage.Mems, &service.Member{MemType: _type})
	return false, nil
}

func stateMsgMemName(parser *Parser, word string, line *[]byte) (bool, error) {
	l := len(parser.curMessage.Mems)
	parser.curMessage.Mems[l-1].MemName = word
	return false, nil
}

func stateBad(parser *Parser, word string, line *[]byte) (bool, error) {
	return false, fmt.Errorf("unexpected string %s", word)
}

func getCol(word string) int {
	var ret int
	switch word {
	case "":
		ret = 0
	case "service":
		ret = 1
	case "message":
		ret = 2
	case "}":
		ret = 3
	default:
		ret = 4
	}
	return ret
}

func GetTypeKind(t string) uint16 {
	var tk uint16
	if t == "stream" || t == "istream" || t == "ostream" {
		tk = service.TypeKindStream
	} else if service.IsBuiltinType(t) {
		tk = service.TypeKindNormal
	} else {
		tk = service.TypeKindMessage
	}
	return tk
}
