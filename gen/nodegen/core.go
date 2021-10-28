package nodegen

import (
	"fmt"
	"gufeijun/hustgen/config"
	"gufeijun/hustgen/gen/utils"
	"gufeijun/hustgen/service"
	"path"
)

func Gen(conf *config.ComplileConfig) error {
	te, err := utils.NewTmplExec(conf, utils.GenFilePath(conf.SrcIDL, conf.OutDir, ".rpch.js"))
	if err != nil {
		return err
	}
	defer te.Close()
	genStatement(te)
	genUseStrict(te)
	genServiceInterfaces(te)
	genHandlers(te)
	genCheckImplementsFunc(te)
	genRegisterFunc(te)
	genExports(te)
	return te.Err
}

func genExports(te *utils.TmplExec) {
	services := make([]string, 0, len(service.GlobalAsset.Services))
	for _, s := range service.GlobalAsset.Services {
		services = append(services, s.Name)
	}
	te.Execute(moduleExportsTmpl, services)
}

func genRegisterFunc(te *utils.TmplExec) {
	for _, s := range service.GlobalAsset.Services {
		data := &struct {
			Name        string
			MethodsName string
			Methods     []string
		}{Name: s.Name}
		data.MethodsName, data.Methods = buildMethodsName(s)
		te.Execute(registerServiceTmpl, data)
	}
}

func genCheckImplementsFunc(te *utils.TmplExec) {
	fmt.Fprint(te.W, checkImplementsTmpl)
}

func genHandlers(te *utils.TmplExec) {
}

type methodDesc struct {
	Desc      string
	Signature string
}

func genServiceInterfaces(te *utils.TmplExec) {
	for _, s := range service.GlobalAsset.Services {
		data := &struct {
			Name    string
			Methods []*methodDesc
		}{Name: s.Name}
		for _, method := range s.Methods {
			data.Methods = append(data.Methods, buildNodeMethod(method))
		}
		te.Execute(serviceInterfaceTmpl, data)
	}
}

func genUseStrict(te *utils.TmplExec) {
	fmt.Fprintln(te.W, `'use strict';`)
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
