package yaml

import (
	"fmt"
	"os"

	"github.com/extrame/unmarshall"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func fetch(node *yaml.Node) map[string]string {

	var fetched = make(map[string]string)
	for i := 0; i < len(node.Content); i++ {
		var c = node.Content[i]
		if c.Kind == yaml.ScalarNode {
			var content = node.Content[i+1]
			if content.Kind == yaml.ScalarNode {
				fetched[c.Value] = content.Value
				i = i + 1
			} else if content.Kind == yaml.SequenceNode {
				for i, sub := range content.Content {
					if sub.Kind == yaml.ScalarNode {
						fetched[fmt.Sprintf("%s[%d]", c.Value, i)] = sub.Value
					} else {
						subFetched := fetch(sub)
						for j, subFetched := range subFetched {
							fetched[fmt.Sprintf("%s[%d].%s", c.Value, i, j)] = subFetched
						}
					}
				}
				i = i + 1
			}
		}
	}

	return fetched
}

func GetChildNode(parent *yaml.Node, name string) *yaml.Node {
	if parent.Kind == yaml.DocumentNode && parent.Anchor == "" {
		parent = parent.Content[0]
	}
	for n, m := range parent.Content {
		if m.Kind == yaml.ScalarNode && m.Value == name && len(parent.Content) > n+1 {
			return parent.Content[n+1]
		}
	}
	logrus.Info("there is no cfg node named:", name)
	return new(yaml.Node)
}

func UnmarshallFile(fileName string, obj interface{}) error {
	f, err := os.Open(fileName)
	if err == nil {
		err = yaml.NewDecoder(f).Decode(obj)
	}
	return err
}

func UnmarshallChild(parent *yaml.Node, name string, obj interface{}) error {
	var node = GetChildNode(parent, name)
	return UnmarshalNode(node, obj)
}

func UnmarshalNode(node *yaml.Node, obj interface{}) error {
	var content = fetch(node)

	var u = unmarshall.Unmarshaller{
		ValueGetter: func(tag string) []string {
			if c, ok := content[tag]; ok {
				return []string{c}
			} else {
				return []string{}
			}

		},
		ValuesGetter: nil,
		TagConcatter: func(prefix string, tag string) string {
			return prefix + "." + tag
		},
		AutoFill: true,
		Tag:      "goblet",
	}
	return u.Unmarshall(obj)
}
