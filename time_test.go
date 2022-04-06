package unmarshall

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestParseRFC3339InLocal(t *testing.T) {
	var parsed = "2019-10-29T03:42:08.076Z"
	ti, e := time.ParseInLocation(time.RFC3339, parsed, time.Local)
	fmt.Println(ti, e)
	parsed = "2019-10-29 03:42"
	ti, e = time.ParseInLocation("2006-01-02 15:04", parsed, time.Local)
	fmt.Println(ti, e)
	ti, e = time.Parse("2006-01-02 15:04", parsed)
	fmt.Println(ti, e)
}

func TestReflectOfSliceByte(t *testing.T) {
	var obj []byte
	ro := reflect.ValueOf(obj)
	rt := ro.Type()
	if rt.Kind() != reflect.Slice {
		t.Fatal("出错啦")
	} else {
		t.Log(rt.Elem().Kind())
	}
}
