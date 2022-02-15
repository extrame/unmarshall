package yaml

import (
	"testing"
)

type TestObj struct {
	Key   string
	Value string `yaml_config:",default"`
}

func TestFromFile(t *testing.T) {
	var obj TestObj
	err := UnmarshallFile("./test_file/default.yaml", &obj)
	if err != nil || obj.Key != "test" || obj.Value != "test" {
		t.Log(err)
		t.Log(obj)
		t.Fail()
	}
}

func TestFromWithNoValueFile(t *testing.T) {
	var obj TestObj
	err := UnmarshallFile("./test_file/novalue.yaml", &obj)
	if err != nil || obj.Key != "test" || obj.Value != "default" {
		t.Log(err)
		t.Log(obj)
		t.Fail()
	}
}
