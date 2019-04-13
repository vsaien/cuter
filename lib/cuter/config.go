package cuter

import (
	"fmt"
	"io/ioutil"
	"log"
	"path"

	"github.com/vsaien/cuter/lib/mapping"
	"github.com/vsaien/cuter/lib/service"
)

type (
	PrivateKeyConf struct {
		Fingerprint string
		KeyFile     string
	}

	SignatureConf struct {
		Strict      bool `json:",default=false"`
		PrivateKeys []PrivateKeyConf
	}
	ServerConfig struct {
		service.ServiceConf
		Host     string
		Port     int
		Verbose  bool `json:",optional"`
		MaxConns int  `json:",optional"`
		// milliseconds
		Timeout   int64         `json:",optional"`
		Signature SignatureConf `json:",optional"`
	}
)

var loaders = map[string]func([]byte, interface{}) error{
	".json": LoadConfigFromJsonBytes,
	".yaml": LoadConfigFromYamlBytes,
	".yml":  LoadConfigFromYamlBytes,
}

func LoadConfig(file string, v interface{}) error {
	if content, err := ioutil.ReadFile(file); err != nil {
		return err
	} else if loader, ok := loaders[path.Ext(file)]; ok {
		return loader(content, v)
	} else {
		return fmt.Errorf("unrecoginized file type: %s", file)
	}
}

func LoadConfigFromJsonBytes(content []byte, v interface{}) error {
	return mapping.UnmarshalJsonBytes(content, v)
}

func LoadConfigFromYamlBytes(content []byte, v interface{}) error {
	return mapping.UnmarshalYamlBytes(content, v)
}

func MustLoadConfig(path string, v interface{}) {
	if err := LoadConfig(path, v); err != nil {
		log.Fatalf("error: config file %s, %s", path, err.Error())
	}
}
