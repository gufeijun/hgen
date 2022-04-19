package config

var Version string = "v0.1.8"

type ComplileConfig struct {
	TargetLang   string
	OutDir       string
	SrcIDL       string
	PrintVersion bool
}
