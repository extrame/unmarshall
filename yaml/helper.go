package yaml

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/extrame/unmarshall"
	"gopkg.in/yaml.v3"
)

var defaultTag = "yaml_config"

func fetch(node *yaml.Node) (fetched map[string][]string) {

	if node != nil {
		fetched = make(map[string][]string)
		execNode(node, fetched)
	}

	return
}

func execNode(content *yaml.Node, fetched map[string][]string) {
	for i := 0; i < len(content.Content); i++ {
		var c = content.Content[i]
		if c.Kind == yaml.ScalarNode {
			execScalarNode(content.Content, i, c.Value, fetched)
			i = i + 1
		} else if c.Kind == yaml.MappingNode {
			execNode(content.Content[0], fetched)
		}
	}
}

func execScalarNode(contents []*yaml.Node, i int, parent string, fetched map[string][]string) {
	var content = contents[i+1]
	if content.Kind == yaml.ScalarNode {
		fetched[parent] = []string{content.Value}
	} else if content.Kind == yaml.SequenceNode {
		for i, sub := range content.Content {
			if sub.Kind == yaml.ScalarNode {
				var arr = fetched[parent]
				for len(arr) <= i {
					arr = append(arr, "")
				}
				arr[i] = sub.Value
				fetched[fmt.Sprintf("%s[]", parent)] = arr
			} else {
				subFetched := fetch(sub)
				for j, subFetched := range subFetched {
					fetched[fmt.Sprintf("%s[%d].%s", parent, i, j)] = subFetched
				}
			}
		}
	}
}

func GetChildNode(parent *yaml.Node, name string) (*yaml.Node, error) {
	if parent.Kind == yaml.DocumentNode && parent.Anchor == "" {
		parent = parent.Content[0]
	}
	for n, m := range parent.Content {
		if m.Kind == yaml.ScalarNode && m.Value == name && len(parent.Content) > n+1 {
			return parent.Content[n+1], nil
		}
	}
	return nil, fmt.Errorf("there is no cfg node named:%s", name)
}

func UnmarshallFile(fileName string, obj interface{}, tagName ...string) error {

	f, err := os.Open(fileName)
	if err == nil {
		var node = new(yaml.Node)
		err = yaml.NewDecoder(f).Decode(node)
		if err == nil || err == io.EOF {
			return UnmarshalNode(node, obj, tagName...)
		}
	}
	return err
}

func UnmarshallReader(source io.Reader, obj interface{}, tagName ...string) error {

	var node = new(yaml.Node)
	err := yaml.NewDecoder(source).Decode(node)
	if err == nil || err == io.EOF {
		return UnmarshalNode(node, obj, tagName...)
	}
	return err
}

func UnmarshallChild(parent *yaml.Node, name string, obj interface{}) error {
	node, err := GetChildNode(parent, name)
	if err == nil {
		err = UnmarshalNode(node, obj)
	}
	return err
}

func UnmarshalNode(node *yaml.Node, obj interface{}, tagName ...string) error {
	var content = fetch(node)

	var tName = defaultTag
	var defaultValTag string
	if len(tagName) > 0 {
		tName = tagName[0]
	}
	if len(tagName) > 1 {
		defaultValTag = tagName[1]
	}

	var u = unmarshall.Unmarshaller{
		ValueGetter: func(tag string) []string {
			tag = strings.ToLower(tag)
			if content != nil {
				if c, ok := content[tag]; ok {
					return c
				}
			}
			return []string{}
		},
		ValuesGetter: nil,
		TagConcatter: func(prefix string, tag string) string {
			return prefix + "." + tag
		},
		AutoFill:   true,
		Tag:        tName,
		DefaultTag: defaultValTag,
	}
	return u.Unmarshall(obj)
}
