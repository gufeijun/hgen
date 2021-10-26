package cgen

import (
	"fmt"
	"gufeijun/hustgen/config"
	"gufeijun/hustgen/gen/utils"
	"gufeijun/hustgen/service"
	"io"
	"path"
	"strings"
	"text/template"
)

func Gen(conf *config.ComplileConfig) error {
	if err := genServerHeaderFile(conf); err != nil {
		return err
	}
	if err := genServerSourceFile(conf); err != nil {
		return err
	}
	if err := genClientHeaderFile(conf); err != nil {
		return err
	}
	return genClientSourceFile(conf)
}

func genDef(w io.Writer, srcIDL string, side string) {
	index := strings.Index(srcIDL, ".")
	if index != -1 {
		srcIDL = srcIDL[:index]
	}
	macro := fmt.Sprintf("__%s_RPCH_%s_H_", srcIDL, side)
	fmt.Fprintf(w, "#ifndef %s\n", macro)
	fmt.Fprintf(w, "#define %s\n\n", macro)
}

func genServerHeaderFile(conf *config.ComplileConfig) error {
	hte, err := utils.NewTmplExec(conf, utils.GenFilePath(conf.SrcIDL, conf.OutDir, ".rpch.server.h"))
	if err != nil {
		return err
	}
	defer hte.Close()
	genStatement(hte)
	genDef(hte.W, conf.SrcIDL, "SERVER")
	genHeaderFileIncludes(hte, []string{`<stdint.h>`, `"error.h"`, `"server.h"`})
	genStructs(hte)
	genStructCreate(hte)
	genServiceMethod(hte)
	fmt.Fprint(hte.W, "#endif")
	return hte.Err
}

func genClientHeaderFile(conf *config.ComplileConfig) error {
	cte, err := utils.NewTmplExec(conf, utils.GenFilePath(conf.SrcIDL, conf.OutDir, ".rpch.client.h"))
	if err != nil {
		return err
	}
	defer cte.Close()
	genStatement(cte)
	genDef(cte.W, conf.SrcIDL, "CLIENT")
	genHeaderFileIncludes(cte, []string{`<stdint.h>`, `"client.h"`})
	genStructs(cte)
	genStructDelete(cte)
	genClientMethod(cte)
	fmt.Fprint(cte.W, "\n#endif")
	return cte.Err
}

func genServerSourceFile(conf *config.ComplileConfig) error {
	cte, err := utils.NewTmplExec(conf, utils.GenFilePath(conf.SrcIDL, conf.OutDir, ".rpch.server.c"))
	if err != nil {
		return err
	}
	defer cte.Close()
	genStatement(cte)
	genSourceFileIncludes(cte, []string{"stdint.h", "stdlib.h", "string.h"}, []string{"argument.h", "cJSON.h", "error.h", "request.h", "server.h"}, "server")
	genArgumentInitAndDestroy(cte, true)
	genErrorMacro(cte, "return")
	genMashalFunc(cte)
	genUnmarshalFunc(cte)
	genHandlers(cte)
	genRegisterService(cte)
	return cte.Err
}

func genClientSourceFile(conf *config.ComplileConfig) error {
	cte, err := utils.NewTmplExec(conf, utils.GenFilePath(conf.SrcIDL, conf.OutDir, ".rpch.client.c"))
	if err != nil {
		return err
	}
	defer cte.Close()
	genStatement(cte)
	genSourceFileIncludes(cte, []string{"stdint.h", "string.h", "stdlib.h"}, []string{"argument.h", "cJSON.h", "error.h", "client.h"}, "client")
	genArgumentInitAndDestroy(cte, false)
	genErrorMacro(cte, "goto end")
	genMashalFunc(cte)
	genUnmarshalFunc(cte)
	genCallFuncs(cte)
	return cte.Err
}

func genStructDelete(te *utils.TmplExec) {
	var data []string
	utils.TraverseRespArgs(func(t *service.Type) bool {
		if t.TypeKind != service.TypeKindMessage {
			return false
		}
		data = append(data, t.TypeName)
		return false
	})
	te.Execute(structDeleteTmpl, data)
}

func genStructCreate(te *utils.TmplExec) {
	var data []string
	utils.TraverseRespArgs(func(t *service.Type) bool {
		if t.TypeKind != service.TypeKindMessage {
			return false
		}
		data = append(data, t.TypeName)
		return false
	})
	te.Execute(structCreateTmpl, data)
}

func genCallFuncs(te *utils.TmplExec) {
	type Data struct {
		HasRtn        bool
		MessageArgs   []string
		FuncSignature string
		RespDefine    string
		RequestInit   string
		ArgInits      []string
		RespCheck     string
		RespUnmarshal string
	}
	utils.TraverseMethod(func(method *service.Method) bool {
		data := &Data{
			FuncSignature: buildCallFuncSignature(method),
			RespDefine:    buildRespArgDefine(method),
			RequestInit:   buildCallRequestInit(method),
			ArgInits:      buildCallArgInits(method),
			RespCheck:     buildRespCheck(method),
			RespUnmarshal: buildRespUnmarshal(method),
			HasRtn:        method.RetType.TypeName != "void",
		}
		var i int
		for _, t := range method.ReqTypes {
			if t.TypeKind == service.TypeKindMessage {
				i++
				data.MessageArgs = append(data.MessageArgs, fmt.Sprintf("node%d", i))
			}
		}
		te.Execute(clientCallTmpl, data)
		return false
	})
}

func genDestroyResponse(te *utils.TmplExec) {
	utils.TraverseRespArgs(func(t *service.Type) bool {
		if t.TypeKind == service.TypeKindNormal {
			return false
		}
		fmt.Fprintf(te.W, "\nvoid %s_destroy(struct %s*);", t.TypeName, t.TypeName)
		return false
	})
}

func genClientMethod(te *utils.TmplExec) {
	for _, s := range service.GlobalAsset.Services {
		var methods []string
		for _, method := range s.Methods {
			methods = append(methods, buildMethod(method, "client_t*"))
		}
		te.Execute(clientMethodTmpl, methods)
	}
}

func genHandlers(te *utils.TmplExec) {
	type Data struct {
		MessageResp   bool
		NoResp        bool
		FuncName      string   //函数名
		Defines       []string //变量定义
		ArgChecks     string   //参数合法性检查
		ArgInits      []string //传参初始化
		ArgUnmarshals []string //传参反序列化
		CallArgs      string
		Resp          string //返回值序列化
		End           string //资源释放
	}
	utils.TraverseMethod(func(method *service.Method) (end bool) {
		data := new(Data)
		data.NoResp = method.RetType.TypeName == "void"
		data.MessageResp = method.RetType.TypeKind == service.TypeKindMessage
		data.FuncName = fmt.Sprintf("%s_%s", method.Service.Name, method.MethodName)
		data.Defines = buildArgDefines(method)
		data.ArgChecks = buildArgChecks(method)
		data.ArgInits = buildArgInits(method)
		data.ArgUnmarshals = buildArgUnmarshals(method)
		data.CallArgs = buildCallArgs(method)
		data.Resp = buildResp(method)
		data.End = buildEnd(method)
		te.Execute(handlerTmpl, data)
		return
	})
}

func common(te *utils.TmplExec, tmpl *template.Template) {
	for _, message := range service.GlobalAsset.Messages {
		data := &struct {
			TypeName   string
			Message    *service.Message
			MessageMem bool
			IDL2CType  map[string]string
		}{TypeName: message.Name, Message: message, IDL2CType: IDLtoCType}
		for _, mem := range message.Mems {
			if mem.MemType.TypeKind == service.TypeKindMessage {
				data.MessageMem = true
				break
			}
		}
		te.Execute(tmpl, data)
	}
}

func genUnmarshalFunc(te *utils.TmplExec) {
	for _, message := range service.GlobalAsset.Messages {
		fmt.Fprintf(te.W, "static void %s_unmarshal(struct %s* dst, char* data, error_t* err);\n",
			message.Name, message.Name)
	}
	common(te, unmarshalFuncTmpl)
}

func genMashalFunc(te *utils.TmplExec) {
	fmt.Fprint(te.W, "\n\n")
	for _, message := range service.GlobalAsset.Messages {
		fmt.Fprintf(te.W, "static cJSON* %s_marshal(struct %s* arg, error_t* err);\n",
			message.Name, message.Name)
	}
	common(te, marshalFuncTmpl)
}

func genRegisterService(te *utils.TmplExec) {
	for _, s := range service.GlobalAsset.Services {
		te.Execute(registerServiceTmpl, s)
	}
}

func genArgumentInitAndDestroy(te *utils.TmplExec, serverSide bool) {
	te.W.Write([]byte{'\n'})
	for _, m := range service.GlobalAsset.Messages {
		fmt.Fprintf(te.W, "static inline __attribute__((always_inline)) void %s_init(struct %s*);\n", m.Name, m.Name)
		fmt.Fprintf(te.W, "static inline __attribute__((always_inline)) void %s_destroy(struct %s*);\n", m.Name, m.Name)
	}

	for _, m := range service.GlobalAsset.Messages {
		data := &struct {
			ServerSide  bool
			Name        string
			MessageMems []*service.Member
			StringMems  []*service.Member
		}{Name: m.Name, ServerSide: serverSide}
		for _, mem := range m.Mems {
			if mem.MemType.TypeKind != service.TypeKindMessage {
				if mem.MemType.TypeName == "string" {
					data.StringMems = append(data.StringMems, mem)
				}
				continue
			}
			data.MessageMems = append(data.MessageMems, mem)
		}
		te.Execute(argumentInitAndDestroyTmpl, data)
	}
}

func genSourceFileIncludes(te *utils.TmplExec, stdlib []string, includes []string, side string) {
	data := &struct {
		Header   string
		Stdlib   []string
		Includes []string
	}{}
	data.Includes = includes
	data.Stdlib = stdlib
	data.Header = path.Base(utils.GenFilePath(te.Conf.SrcIDL, te.Conf.OutDir, ".rpch."+side+".h"))
	te.Execute(sourceFileIncludesTmpl, data)
}

func genServiceMethod(te *utils.TmplExec) {
	for _, s := range service.GlobalAsset.Services {
		data := &struct {
			ServiceName string
			Methods     []string
		}{ServiceName: s.Name}
		for _, method := range s.Methods {
			data.Methods = append(data.Methods, buildMethod(method, "error_t*"))
		}
		te.Execute(serviceMethodTmpl, data)
	}
}

func genStructs(te *utils.TmplExec) {
	te.Execute(structStateTmpl, service.GlobalAsset.Messages)
	for _, message := range service.GlobalAsset.Messages {
		s := &struct {
			Name    string
			Members []string
		}{}
		s.Name = message.Name
		s.Members = make([]string, len(message.Mems))
		for i, mem := range message.Mems {
			s.Members[i] = fmt.Sprintf("%s %s", toClangType(mem.MemType, true), mem.MemName)
		}
		te.Execute(structTmpl, s)
	}
}

func genHeaderFileIncludes(te *utils.TmplExec, includes []string) {
	te.Execute(includesTmpl, includes)
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

func genErrorMacro(te *utils.TmplExec, action string) {
	te.W.Write([]byte{'\n'})
	te.Execute(macroTmpl, action)
}
