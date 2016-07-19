package main

import (
	"gopkg.in/yaml.v2"
)

type TestributorYml map[string]interface{}

func NewTestributorYml(yml_contents string) (TestributorYml, error) {
	testributorYml := make(TestributorYml)

	err := yaml.Unmarshal([]byte(yml_contents), &testributorYml)
	if err != nil {
		return *new(TestributorYml), err
	} else {
		return testributorYml, nil
	}
}
