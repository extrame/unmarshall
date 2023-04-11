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

func TestFromArray(t *testing.T) {
	var obj []TestObj1
	err := UnmarshallFile("./test_file/array.yaml", &obj, defaultTag, "default")
	if err != nil {
		t.Log(err)
		t.Fail()
	} else if len(obj) != 2 {
		t.Error("array length error")
		t.Log(obj)
		t.Fail()
	} else if err != nil || obj[0].Key != "test1" || obj[0].Value != "default1" || obj[1].Value != "default2" {
		t.Error("array value error")
		t.Log(obj)
		t.Fail()
	}
	t.Log(obj)
}
