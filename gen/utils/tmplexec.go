package utils

import (
	"gufeijun/hustgen/config"
	"io"
	"os"
	"text/template"
)

type TmplExec struct {
	Conf *config.ComplileConfig
	W    io.Writer
	Err  error
	file *os.File
}

func NewTmplExec(conf *config.ComplileConfig, path string) (*TmplExec, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return nil, err
	}
	te := &TmplExec{
		Conf: conf,
		file: file,
	}
	te.W = &errWriter{Writer: file, te: te}
	return te, nil
}

func (te *TmplExec) Execute(tmpl *template.Template, data interface{}) {
	if te.Err != nil {
		return
	}
	if err := tmpl.Execute(te.W, data); err != nil {
		te.Err = err
	}
}

func (te *TmplExec) Close() error {
	return te.file.Close()
}

type errWriter struct {
	te *TmplExec
	io.Writer
}

func (ew *errWriter) Write(p []byte) (n int, err error) {
	if ew.te.Err != nil {
		return 0, ew.te.Err
	}
	n, ew.te.Err = ew.Writer.Write(p)
	return n, ew.te.Err
}
