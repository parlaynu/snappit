package main

import (
	"bytes"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Archives []struct {
		Label  string
		Source string
	}

	Arena string

	SkipDirs []string `yaml:"skip_dirs"`

	DeepScan bool
	Prune bool
}

func LoadConfig(path string) (*Config, error) {
	path = expand_tilde(path)

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var data bytes.Buffer
	io.Copy(&data, f)

	var config Config
	err = yaml.Unmarshal(data.Bytes(), &config)
	if err != nil {
		return nil, err
	}

	for i, archive := range config.Archives {
		config.Archives[i].Source = expand_tilde(archive.Source)
	}
	config.Arena = expand_tilde(config.Arena)

	return &config, nil
}

func expand_tilde(path string) string {
	user, err := user.Current()
	if err != nil {
		return path
	}
	npath, err := filepath.Abs(strings.Replace(path, "~", user.HomeDir, 1))
	if err != nil {
		return path
	}
	return npath
}
