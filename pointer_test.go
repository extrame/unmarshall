package unmarshall

import (
	"fmt"
	"net/url"
	"testing"
)

type Tested struct {
	A *TestedA
	B *TestedB
}

type TestedA struct {
	A1 string
	A2 string
}

type TestedB struct {
	B1 string
	B2 string
}

func TestObjectWithPointeredChild(t *testing.T) {
	var obj Tested
	var form = make(url.Values)
	form.Set("A[A1]", "a1")
	form.Set("A[A2]", "a2")
	err := Unmarshal(&obj, form, true)
	fmt.Println(obj, err)
	fmt.Println(obj.A)
}
