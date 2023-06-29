package yaml

import (
	"fmt"
	"io"
	"net/url"
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
	if content.Kind == yaml.MappingNode {
		for i := 0; i < len(content.Content); i = i + 2 {
			var c = content.Content[i]
			execScalarNode(content.Content[i+1], c.Value, fetched)
		}
	} else if content.Kind == yaml.SequenceNode {
		for i, sub := range content.Content {
			execSequenceNode(sub, i, "", fetched)
		}
	} else if content.Kind == yaml.ScalarNode {
		execScalarNode(content, "", fetched)
	} else if content.Kind == yaml.DocumentNode {
		for i := 0; i < len(content.Content); i = i + 2 {
			var c = content.Content[i]
			execNode(c, fetched)
		}
	}
	//
	// 	i = i + 1
	// }

	// for i := 0; i < len(content.Content); i++ {
	// 	var c = content.Content[i]
	// 	fmt.Println(c)
	// 	 else if c.Kind == yaml.MappingNode {
	// 		execNode(content.Content[0], fetched)
	// 	} else if c.Kind == yaml.SequenceNode {
	// 		fmt.Println(fetched)
	// 		for j, sub := range c.Content {
	// 			execSequenceNode(sub, j, c.Value, fetched)
	// 		}
	// 	}
	// }
}

func execSequenceNode(sub *yaml.Node, i int, parent string, fetched map[string][]string) {
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

func execScalarNode(content *yaml.Node, parent string, fetched map[string][]string) {
	if content.Kind == yaml.ScalarNode {
		fetched[parent] = []string{content.Value}
	} else if content.Kind == yaml.SequenceNode {
		for i, sub := range content.Content {
			execSequenceNode(sub, i, parent, fetched)
		}
	} else if content.Kind == yaml.MappingNode {
		for i := 0; i < len(content.Content); i = i + 2 {
			var c = content.Content[i]
			var name string
			if parent != "" {
				name = parent + "." + c.Value
			} else {
				name = c.Value
			}
			execScalarNode(content.Content[i+1], name, fetched)
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
		ValuesGetter: func(tag string) url.Values {
			tag = strings.ToLower(tag)
			var values = make(url.Values)
			for k, v := range content {
				if strings.HasPrefix(k, tag) {
					values[k] = v
				}
			}
			return values
		},
		BaseName: func(path string, prefix string) string {
			prefix = strings.ToLower(prefix)
			return strings.Split(strings.TrimPrefix(path, prefix+"."), ".")[0]
		},
		TagConcatter: func(prefix string, tag string) string {
			return strings.ToLower(prefix + "." + tag)
		},
		AutoFill:              true,
		Tag:                   tName,
		DefaultTag:            defaultValTag,
		ArrayValueGotByOffset: true,
	}
	return u.Unmarshall(obj)
}
