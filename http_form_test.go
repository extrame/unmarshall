package unmarshall

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
)

type TestedType struct {
	Name       string
	Definition TestedNestedType
}

type TestedNestedType struct {
	Name string
	Up   []TestedNestedType
	Down []TestedNestedType
}

func TestUnmarshal(t *testing.T) {
	var obj TestedType
	err := Unmarshal(&obj, true)
	fmt.Println(obj, err)
}

func concatPrefix(prefix, tag string) string {
	return prefix + "[" + tag + "]"
}

func Unmarshal(v interface{}, autofill bool) error {
	var form = make(url.Values)
	form.Set("Name", "tested")
	form.Set("Definition[Name]", "testedNested")
	form.Set("Definition[Up][0][Name]", "testedNestedA")
	form.Set("Definition[Up][0][Up][0][Name]", "testedNestedA-A")

	var maxlength = 0
	for k, _ := range form {
		if len(k) > maxlength {
			maxlength = len(k)
		}
	}

	var unmarshaller = Unmarshaller{
		Values: func() map[string][]string {
			return form
		},
		MaxLength: maxlength,
		ValueGetter: func(tag string) []string {
			values := (*map[string][]string)(&form)
			fmt.Println(tag)
			if values != nil {
				if results, ok := (*values)[tag]; !ok {
					//get the value of [Tag] from [tag](lower case), it maybe a problem TODO
					return (*values)[strings.ToLower(tag)]
				} else {
					return results
				}
			}
			return []string{}
		},
		TagConcatter: concatPrefix,
		AutoFill:     autofill,
	}

	return unmarshaller.Unmarshall(v)
}
