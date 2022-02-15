package yaml

import (
	"testing"
)

type TestObj struct {
	Key   string
	Value string `yaml_config:",default"`
}

type TestObj1 struct {
	Key   string
	Value string `default:"default"`
}

type TestChildType struct {
	TestObj1
	Value2 string
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
	var obj TestObj1
	err := UnmarshallFile("./test_file/novalue.yaml", &obj, defaultTag, "default")
	if err != nil || obj.Key != "test" || obj.Value != "default" {
		t.Log(err)
		t.Log(obj)
		t.Fail()
	}
}

func TestFromWithDefaultValueFile(t *testing.T) {
	var obj TestObj
	err := UnmarshallFile("./test_file/novalue.yaml", &obj)
	if err != nil || obj.Key != "test" || obj.Value != "default" {
		t.Log(err)
		t.Log(obj)
		t.Fail()
	}
}

func TestFromWithChildFile(t *testing.T) {
	var obj TestChildType
	err := UnmarshallFile("./test_file/novalue.yaml", &obj, defaultTag, "default")
	if err != nil || obj.Key != "test" || obj.Value != "default" || obj.Value2 != "test" {
		t.Log(err)
		t.Log(obj)
		t.Fail()
	}
}
