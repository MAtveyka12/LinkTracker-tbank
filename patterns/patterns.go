package patterns

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Patterns struct {
	GitHub        string `yaml:"github"`
	StackOverflow string `yaml:"stackoverflow"`
}

type Cfg struct {
	Patterns Patterns `yaml:"patterns"`
}

func LoadPatterns(path string) (*Cfg, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	var cfg Cfg

	err = yaml.Unmarshal(data, &cfg)

	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
