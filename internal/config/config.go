package config

import (
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"log"
)

var Global = koanf.New(".")

func init() {
	err := Global.Load(file.Provider("./config.toml"), toml.Parser())
	if err != nil {
		log.Fatalln(err)
	}
}
