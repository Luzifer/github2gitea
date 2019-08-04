package main

import (
	"os"
	"regexp"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

type mapFile struct {
	Mappings []mapping `yaml:"mappings"`
}

type mapping struct {
	SourceExpression string `yaml:"source_expression"`
	TargetUser       int64  `yaml:"target_user"`
	TargetUserName   string `yaml:"target_user_name"`
}

func loadMapFile(fileName string) (*mapFile, error) {
	if _, err := os.Stat(fileName); err != nil {
		return nil, errors.Wrap(err, "Mapping file not available")
	}

	f, err := os.Open(fileName)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open mapping file")
	}
	defer f.Close()

	var out = &mapFile{}
	return out, errors.Wrap(yaml.NewDecoder(f).Decode(out), "Unable to decode mapping file")
}

func newMapFile() *mapFile {
	return &mapFile{}
}

func (m mapFile) GetMapping(repoName string) *mapping {
	for _, me := range m.Mappings {
		if regexp.MustCompile(me.SourceExpression).MatchString(repoName) {
			return &me
		}
	}

	return nil
}

func (m mapFile) MappingAvailable(repoName string) bool {
	for _, me := range m.Mappings {
		if regexp.MustCompile(me.SourceExpression).MatchString(repoName) {
			return true
		}
	}

	return false
}
