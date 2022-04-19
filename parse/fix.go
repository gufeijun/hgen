package parse

import (
	"fmt"
	"os"
)

// 检查以下错误：
// 1. 同一个message不能有相同的成员				√
// 2. 同一个service不能有相同的method			√
// 3. 不能有相同的message(saveMessage时检查)	√
// 4. 不能有相同的service(saveService时检查)	√
// 5. 一个方法的请求参数只能有一个stream		√
// 6. message成员不能是stream类型				√
// 7. 是否使用未定义的message类型				√

func fixSymbols(syms *Symbols) {
	for _, msg := range syms.Messages {
		checkMessage(msg, syms)
	}
	for _, svr := range syms.Services {
		checkService(svr, syms)
	}
}

func checkMessage(msg *Message, syms *Symbols) {
	m := make(map[string]struct{})
	for _, mem := range msg.Mems {
		checkRepeatedDefine(m, mem.Name, "member", msg.Name, "message")
		checkUndefine(syms.Messages, mem.Type.Name, "message", msg.Name)
		checkMemberType(mem.Type.Name, msg.Name)
		m[mem.Name] = struct{}{}
	}
}

func checkService(srv *Service, syms *Symbols) {
	m := make(map[string]struct{})
	for _, method := range srv.Methods {
		checkRepeatedDefine(m, method.Name, "method", srv.Name, "service")
		checkUndefine(syms.Messages, method.RetType.Name, "service", srv.Name)

		occurStream := isStream(method.RetType.Name)
		for _, t := range method.ReqTypes {
			if t.Name == "void" {
				method.ReqTypes = nil
				return
			}
			checkUndefine(syms.Messages, t.Name, "service", srv.Name)
			occurStream = checkAtMostOneStream(occurStream, t.Name, srv.Name, method.Name)
		}

		m[method.Name] = struct{}{}
	}
}

func checkAtMostOneStream(occurStream bool, tName, service, method string) bool {
	if !isStream(tName) {
		return occurStream
	}
	if occurStream {
		fmt.Printf("[%s.%s]: method must have at most one stream type in parameters\n", service, method)
		os.Exit(0)
	}
	return true
}

func checkRepeatedDefine(m map[string]struct{}, what, t1, of, t2 string) {
	if _, ok := m[what]; !ok {
		return
	}
	fmt.Printf("repeatedly defined %s \"%s\" of %s \"%s\"\n", t1, what, t2, of)
	os.Exit(0)
}

func checkUndefine(m map[string]*Message, what, t1, t2 string) {
	if isBuiltin(what) {
		return
	}
	if _, ok := m[what]; ok {
		return
	}
	fmt.Printf("undefined type \"%s\" in %s \"%s\"\n", what, t1, t2)
	os.Exit(0)
}

func checkMemberType(t, message string) {
	if !isStream(t) {
		return
	}
	fmt.Printf("invalid stream member in message \"%s\"\n", message)
	os.Exit(0)
}
