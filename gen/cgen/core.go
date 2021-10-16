package cgen

import (
	"fmt"
	"gufeijun/hustgen/config"
	"gufeijun/hustgen/gen/utils"
	"gufeijun/hustgen/service"
	"path"
	"text/template"
)

func genHeaderFile(conf *config.ComplileConfig) error {
	hte, err := utils.NewTmplExec(conf, utils.GenFilePath(conf.SrcIDL, conf.OutDir, ".rpch.h"))
	if err != nil {
		return err
	}
	defer hte.Close()
	genStatement(hte)
	genHeaderFileIncludes(hte)
	genStructs(hte)
	genServiceMethod(hte)
	return hte.Err
}

func genSourceFile(conf *config.ComplileConfig) error {
	cte, err := utils.NewTmplExec(conf, utils.GenFilePath(conf.SrcIDL, conf.OutDir, ".rpch.c"))
	if err != nil {
		return err
	}
	defer cte.Close()
	genStatement(cte)
	genSourceFileIncludes(cte)
	genArgumentInitAndDestroy(cte)
	genMashalFunc(cte)
	genUnmarshalFunc(cte)
	//TODO
	genHandlers(cte)
	genRegisterService(cte)

	return cte.Err
}

func Gen(conf *config.ComplileConfig) error {
	if err := genHeaderFile(conf); err != nil {
		return err
	}
	return genSourceFile(conf)
}

func genHandlers(te *utils.TmplExec) {
	utils.TraverseMethod(func(method *service.Method) (end bool) {

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

func genArgumentInitAndDestroy(te *utils.TmplExec) {
	te.W.Write([]byte{'\n'})
	for _, m := range service.GlobalAsset.Messages {
		fmt.Fprintf(te.W, "static inline __attribute__((always_inline)) void %s_init(struct %s*);\n", m.Name, m.Name)
		fmt.Fprintf(te.W, "static inline __attribute__((always_inline)) void %s_destroy(struct %s*);\n", m.Name, m.Name)
	}

	for _, m := range service.GlobalAsset.Messages {
		data := &struct {
			Name        string
			MessageMems []*service.Member
		}{Name: m.Name}
		for _, mem := range m.Mems {
			if mem.MemType.TypeKind != service.TypeKindMessage {
				continue
			}
			data.MessageMems = append(data.MessageMems, mem)
		}
		te.Execute(argumentInitAndDestroyTmpl, data)
	}
}

func genSourceFileIncludes(te *utils.TmplExec) {
	data := &struct {
		Header   string
		Stdlib   []string
		Includes []string
	}{Stdlib: []string{"stdint.h", "stdlib.h", "string.h"},
		Includes: []string{"argument.h", "cJSON.h", "error.h", "request.h", "server.h"}}
	data.Header = path.Base(utils.GenFilePath(te.Conf.SrcIDL, te.Conf.OutDir, ".rpch.h"))
	te.Execute(sourceFileIncludesTmpl, data)
}

func genServiceMethod(te *utils.TmplExec) {
	for _, s := range service.GlobalAsset.Services {
		data := &struct {
			ServiceName string
			Methods     []string
		}{ServiceName: s.Name}
		for _, method := range s.Methods {
			data.Methods = append(data.Methods, buildMethod(method))
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
			s.Members[i] = fmt.Sprintf("%s %s", toClangType(mem.MemType), mem.MemName)
		}
		te.Execute(structTmpl, s)
	}
}

func genHeaderFileIncludes(te *utils.TmplExec) {
	var includes = []string{`<stdint.h>`, `"error.h"`, `"server.h"`}
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
