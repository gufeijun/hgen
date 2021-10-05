package config

var Version string = "v0.1.0"

type ComplileConfig struct {
	TargetLang string
	OutDir     string
	SrcIDL     string
}
