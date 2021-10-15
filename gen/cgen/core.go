package cgen

import (
	"gufeijun/hustgen/config"
	"gufeijun/hustgen/gen/utils"
	"path"
	"strings"
)

func genFilepath(srcIDL string, outdir string, suffix string) string {
	index := strings.Index(srcIDL, ".")
	if index != -1 {
		srcIDL = srcIDL[:index]
	}
	return path.Join(outdir, path.Base(srcIDL)+suffix)
}

func Gen(conf *config.ComplileConfig) error {
	hte, err := utils.NewTmplExec(conf, utils.GenFilePath(conf.SrcIDL, conf.OutDir, "rpch.h"))
	if err != nil {
		return err
	}
	defer hte.Close()

	return hte.Err
}
