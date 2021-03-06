package gogen

import (
	"errors"
	"fmt"
	"gufeijun/hustgen/config"
	"gufeijun/hustgen/gen/utils"
	"gufeijun/hustgen/parse"
	"io"
	"path"
)

const (
	RPCH = `rpch "github.com/gufeijun/rpch-go"`
	IO   = `"io"`
	JSON = `"encoding/json"`
)

var infos *parse.Symbols

func Gen(_infos *parse.Symbols, conf *config.ComplileConfig) error {
	infos = _infos
	te, err := utils.NewTmplExec(conf, utils.GenFilePath(conf.SrcIDL, conf.OutDir, ".rpch.go"))
	if err != nil {
		return err
	}
	defer te.Close()
	genStatement(te)
	genPackage(te)
	genImports(te)
	genMessages(te)
	genServiceInterfaces(te)
	genServiceRegisterFunc(te)
	genInit(te)
	genClientStruct(te)
	genClientMethods(te)
	return te.Err
}

type CallArg struct {
	TypeKind uint16
	TypeName string
	Data     string
}

func genStatement(te *utils.TmplExec) {
	te.Execute(statementTmpl, struct {
		Version string
		Source  string
	}{
		Version: config.Version,
		Source:  path.Base(te.Conf.SrcIDL),
	})
}

func genClientMethods(te *utils.TmplExec) {
	for _, s := range infos.Services {
		for _, method := range s.Methods {
			data := &struct {
				ServiceName string
				MethodName  string
				RequestArg  string
				ResponseArg string
				Return      string
				CallArgs    []*CallArg
			}{
				ServiceName: s.Name,
				MethodName:  method.Name,
				RequestArg:  buildRequestArgs(method.ReqTypes),
				ResponseArg: buildResponseArg(method.RetType),
				Return:      buildReturn(method.RetType),
				CallArgs:    buildCallArgs(method.ReqTypes),
			}
			te.Execute(clientMethodTmpl, data)
		}
	}
}

func genClientStruct(te *utils.TmplExec) {
	for _, s := range infos.Services {
		te.Execute(clientStructTmpl, s.Name)
	}
}

func genInit(te *utils.TmplExec) {
	messages := infos.Messages
	if len(messages) == 0 {
		return
	}
	var msgs []string
	for _, m := range messages {
		msgs = append(msgs, m.Name)
	}
	te.Execute(initTmpl, msgs)
}

func genServiceRegisterFunc(te *utils.TmplExec) {
	type MethodDesc struct {
		MethodName  string
		RetTypeName string
	}
	for _, s := range infos.Services {
		var descs []*MethodDesc
		for _, method := range s.Methods {
			tn := method.RetType.Name
			if tn == "void" {
				tn = ""
			}
			descs = append(descs, &MethodDesc{MethodName: method.Name, RetTypeName: tn})
		}
		data := &struct {
			Name        string
			MethodDescs []*MethodDesc
		}{
			Name:        s.Name,
			MethodDescs: descs,
		}
		te.Execute(serviceRegisterTmpl, data)
	}
}

func genServiceInterfaces(te *utils.TmplExec) {
	for _, s := range infos.Services {
		var methods []string
		for _, method := range s.Methods {
			methods = append(methods, toGolangMethod(method))
		}
		data := &struct {
			Name    string
			Methods []string
		}{
			Name:    s.Name,
			Methods: methods,
		}
		te.Execute(serviceInterfaceTmpl, data)
	}
}

func genPackage(te *utils.TmplExec) {
	packageName := path.Base(te.Conf.OutDir)
	if packageName == "/" {
		te.Err = errors.New("outDir can not be /")
	}
	io.WriteString(te.W, fmt.Sprintf("package %s\n", packageName))
}

func genMessages(te *utils.TmplExec) {
	for _, message := range infos.Messages {
		te.Execute(structTmpl, message)
	}
}

func genImports(te *utils.TmplExec) {
	if len(infos.Services) == 0 {
		return
	}
	var addIO bool
	var addJson bool
	utils.TraverseMethod(infos, func(method *parse.Method) bool {
		if method.RetType.Kind == parse.TypeKindMessage {
			addJson = true
		}
		for _, t := range append(method.ReqTypes, method.RetType) {
			if t.Kind == parse.TypeKindStream {
				addIO = true
			}
		}
		return addIO && addJson
	})
	var imports []string
	if addIO {
		imports = append(imports, IO)
	}
	if addJson {
		imports = append(imports, JSON)
	}
	imports = append(imports, RPCH)
	te.Execute(importTmpl, imports)
}
