package gogen

import (
	"errors"
	"fmt"
	"gufeijun/hustgen/config"
	"gufeijun/hustgen/service"
	"io"
	"os"
	"path"
	"strings"
	"text/template"
)

const (
	RPCH = `rpch "github.com/gufeijun/rpch-go"`
	IO   = `"io"`
	JSON = `"encoding/json"`
)

type asset struct {
	conf *config.ComplileConfig
	file io.Writer
	err  error
}

type errWriter struct {
	asset *asset
	io.Writer
}

func (ew *errWriter) Write(p []byte) (n int, err error) {
	if ew.asset.err != nil {
		return 0, ew.asset.err
	}
	n, ew.asset.err = ew.Writer.Write(p)
	return n, ew.asset.err
}

func genFilepath(srcIDL string, outdir string) string {
	index := strings.Index(srcIDL, ".")
	if index != -1 {
		srcIDL = srcIDL[:index]
	}
	return path.Join(outdir, path.Base(srcIDL)+".rpch.go")
}

func Gen(conf *config.ComplileConfig) error {
	file, err := os.OpenFile(genFilepath(conf.SrcIDL, conf.OutDir), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	asset := &asset{
		conf: conf,
	}
	asset.file = &errWriter{Writer: file, asset: asset}
	return asset.gen()
}

func (asset *asset) gen() error {
	asset.genStatement()
	asset.genPackage()
	asset.genImports()
	asset.genMessages()
	asset.genServiceInterfaces()
	asset.genServiceRegisterFunc()
	asset.genInit()
	asset.genClientStruct()
	asset.genClientMethods()

	return asset.err
}

type CallArg struct {
	TypeKind uint16
	TypeName string
	Data     string
}

func (asset *asset) genStatement() {
	asset.execute(statementTmpl, struct {
		Version string
		Source  string
	}{
		Version: config.Version,
		Source:  path.Base(asset.conf.SrcIDL),
	})
}

func (asset *asset) genClientMethods() {
	for _, s := range service.GlobalAsset.Services {
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
				MethodName:  method.MethodName,
				RequestArg:  buildRequestArgs(method.ReqTypes),
				ResponseArg: buildResponseArg(method.RetType),
				Return:      buildReturn(method.RetType),
				CallArgs:    buildCallArgs(method.ReqTypes),
			}
			asset.execute(clientMethodTmpl, data)
		}
	}
}

func (asset *asset) genClientStruct() {
	for _, s := range service.GlobalAsset.Services {
		asset.execute(clientStructTmpl, s.Name)
	}
}

func (asset *asset) genInit() {
	messages := service.GlobalAsset.Messages
	if len(messages) == 0 {
		return
	}
	var msgs []string
	for _, m := range messages {
		msgs = append(msgs, m.Name)
	}
	asset.execute(initTmpl, msgs)
}

func (asset *asset) genServiceRegisterFunc() {
	type MethodDesc struct {
		MethodName  string
		RetTypeName string
	}
	for _, s := range service.GlobalAsset.Services {
		var descs []*MethodDesc
		for _, method := range s.Methods {
			tn := method.RetType.TypeName
			if tn == "void" {
				tn = ""
			}
			descs = append(descs, &MethodDesc{MethodName: method.MethodName, RetTypeName: tn})
		}
		data := &struct {
			Name        string
			MethodDescs []*MethodDesc
		}{
			Name:        s.Name,
			MethodDescs: descs,
		}
		asset.execute(serviceRegisterTmpl, data)
	}
}

func (asset *asset) genServiceInterfaces() {
	for _, s := range service.GlobalAsset.Services {
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
		asset.execute(serviceInterfaceTmpl, data)
	}
}

func (asset *asset) genPackage() {
	packageName := path.Base(asset.conf.OutDir)
	if packageName == "/" {
		asset.err = errors.New("outDir can not be /")
	}
	io.WriteString(asset.file, fmt.Sprintf("package %s\n", packageName))
}

func (asset *asset) genMessages() {
	for _, message := range service.GlobalAsset.Messages {
		asset.execute(structTmpl, message)
	}
}

func (asset *asset) genImports() {
	if len(service.GlobalAsset.Services) == 0 {
		return
	}
	var addIO bool
	var addJson bool
	traverseMethod(func(method *service.Method) bool {
		if method.RetType.TypeKind == service.TypeKindMessage {
			addJson = true
		}
		for _, t := range append(method.ReqTypes, method.RetType) {
			if t.TypeKind == service.TypeKindStream {
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
	asset.execute(importTmpl, imports)
}

func traverseMethod(callback func(method *service.Method) (end bool)) {
	for _, s := range service.GlobalAsset.Services {
		for _, method := range s.Methods {
			if end := callback(method); end {
				return
			}
		}
	}
}

func (asset *asset) execute(tmpl *template.Template, data interface{}) {
	if asset.err != nil {
		return
	}
	if err := tmpl.Execute(asset.file, data); err != nil {
		asset.err = err
	}
}
