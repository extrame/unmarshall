package unmarshall

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"
)

type TestedType struct {
	Name       string
	Sex        string
	Definition TestedNestedType
}

type TestedType20191018 struct {
	Transport            string        `protobuf:"bytes,1,opt,name=Transport,proto3" json:"Transport,omitempty"`
	Protocol             string        `protobuf:"bytes,2,opt,name=Protocol,proto3" json:"Protocol,omitempty"`
	ProtocolUUID         string        `protobuf:"bytes,7,opt,name=ProtocolUUID,proto3" json:"ProtocolUUID,omitempty"`
	TimeOut              string        `protobuf:"bytes,3,opt,name=TimeOut,proto3" json:"TimeOut,omitempty"`
	Pull                 []*PullConfig `protobuf:"bytes,4,rep,name=Pull,proto3" json:"Pull,omitempty"`
	Transforms           []*Transforms `protobuf:"bytes,5,rep,name=Transforms,proto3" json:"Transforms,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-" xorm:"-" bson:"-"`
	XXX_unrecognized     []byte        `json:"-" xorm:"-" bson:"-"`
	XXX_sizecache        int32         `json:"-" xorm:"-" bson:"-"`
}

type PullConfig struct {
	Address              string   `protobuf:"bytes,1,opt,name=Address,proto3" json:"Address,omitempty"`
	Period               string   `protobuf:"bytes,2,opt,name=Period,proto3" json:"Period,omitempty"`
	Length               uint64   `protobuf:"varint,3,opt,name=Length,proto3" json:"Length,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" xorm:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" xorm:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" xorm:"-" bson:"-"`
}

type Transforms struct {
	Address              string       `protobuf:"bytes,2,opt,name=Address,proto3" json:"Address,omitempty"`
	Trans                []*Transform `protobuf:"bytes,1,rep,name=Trans,proto3" json:"Trans,omitempty"`
	XXX_NoUnkeyedLiteral struct{}     `json:"-" xorm:"-" bson:"-"`
	XXX_unrecognized     []byte       `json:"-" xorm:"-" bson:"-"`
	XXX_sizecache        int32        `json:"-" xorm:"-" bson:"-"`
}

type Transform struct {
	As                   string   `protobuf:"bytes,1,opt,name=As,proto3" json:"As,omitempty"`
	Type                 string   `protobuf:"bytes,2,opt,name=Type,proto3" json:"Type,omitempty"`
	Calculator           string   `protobuf:"bytes,3,opt,name=Calculator,proto3" json:"Calculator,omitempty"`
	CalculatorForExec    string   `protobuf:"bytes,7,opt,name=CalculatorForExec,proto3" json:"CalculatorForExec,omitempty"`
	For                  string   `protobuf:"bytes,4,opt,name=For,proto3" json:"For,omitempty"`
	Omit                 string   `protobuf:"bytes,5,opt,name=Omit,proto3" json:"Omit,omitempty"`
	Tag                  string   `protobuf:"bytes,6,opt,name=Tag,proto3" json:"Tag,omitempty"`
	Offset               uint32   `protobuf:"varint,8,opt,name=Offset,proto3" json:"Offset,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-" xorm:"-" bson:"-"`
	XXX_unrecognized     []byte   `json:"-" xorm:"-" bson:"-"`
	XXX_sizecache        int32    `json:"-" xorm:"-" bson:"-"`
}

type TestedMapType struct {
	Name       string
	Sex        string
	Definition map[string]TestedNestedType1
	Nest       map[string]int
}

type TestedMapType2 struct {
	Name       string
	Sex        string
	Definition map[string]map[string]*TestedNestedType1
}

type TestedNestedType struct {
	Name string
	Up   []TestedNestedType
	Down []TestedNestedType
}

type TestedNestedType1 struct {
	Name  string
	Value string
}

type TestedArrayStringType struct {
	Name []string
}

type TestedArrayComplicatedType struct {
	// Objs []TestedType1
}

func TestUnmarshalComplicatedObj(t *testing.T) {
	var obj TestedType20191018
	var form = make(url.Values)
	form.Set("Protocol", "tested")
	form.Set("Transport", "ddp")
	form.Set("Device", "")
	form.Set("Bund", "9600")
	form.Set("Port", "3232")
	form.Set("Pull[0][Address]", "13911111112:0x020000")
	form.Set("Pull[0][Length]", "80")
	form.Set("Pull[0][Period]", "10s")
	form.Set("Bind[0][Address]", "")
	form.Set("AddressSet[0][Address]", "")
	form.Set("AddressSet[0][OmitError]", "")
	form.Set("Transforms[0][Address]", "0x020000")
	form.Set("Transforms[0][Trans][0][For]", "")
	form.Set("Transforms[0][Trans][0][Tag]", "")
	form.Set("Transforms[0][Trans][0][Calculator]", "")
	form.Set("Transforms[0][Trans][0][CalculatorForExec]", "")
	form.Set("Transforms[0][Trans][0][Offset]", "")
	err := Unmarshal(&obj, form, true)
	fmt.Println(obj, err)
	fmt.Println(obj.Pull[0], err)
	fmt.Println(obj.Transforms[0], err)
}

func TestUnmarshal(t *testing.T) {
	var obj TestedType
	var form = make(url.Values)
	form.Set("Name", "tested")
	form.Set("Sex", "man")
	form.Set("Definition[Name]", "testedNested")
	form.Set("Definition[Name][Value]", "one")
	form.Set("Definition[Body][Name]", "two")
	form.Set("Definition[Body][Value]", "three")
	err := Unmarshal(&obj, form, true)
	fmt.Println(obj, err)
}

func TestUnmarshalMap(t *testing.T) {
	var obj TestedMapType
	var form = make(url.Values)
	form.Set("Name", "tested")
	form.Set("Sex", "man")
	form.Set("Definition[Name][Name]", "testedNested")
	form.Set("Definition[Name][Value]", "one")
	form.Set("Definition[Body][Name]", "two")
	form.Set("Definition[Body][Value]", "three")
	form.Set("Nest[Test1]", "45")
	form.Set("Nest[Test2]", "1")
	err := Unmarshal(&obj, form, true)
	fmt.Println(obj, err)
}

func TestUnmarshalMap2(t *testing.T) {
	var obj TestedMapType2
	var form = make(url.Values)
	form.Set("Name", "tested")
	form.Set("Sex", "man")
	form.Set("Definition[Name][Name2][Name]", "testedNested")
	form.Set("Definition[Name][Name2][Value]", "one")
	form.Set("Definition[Body][Body2][Name]", "two")
	form.Set("Definition[Body][Body2][Value]", "three")

	err := Unmarshal(&obj, form, true)
	fmt.Println(obj, err)
	fmt.Println(obj.Definition["Body"]["Body2"])
}

func TestUnmarshalArray(t *testing.T) {
	var obj TestedArrayStringType
	var form = make(url.Values)
	form.Add("Name[]", "tested1")
	form.Add("Name[]", "tested2")
	err := Unmarshal(&obj, form, true)
	if err != nil {
		t.Error(err)
		t.Fail()
	} else if obj.Name[0] != "tested1" || obj.Name[1] != "tested2" {
		t.Error("Unmarshal failed")
		t.Fail()
	}
	fmt.Println(obj, form, err)
}

func TestUnmarshalComplicatedArray(t *testing.T) {
	var obj TestedArrayComplicatedType
	var form = make(url.Values)
	form.Add("Objs[0][Name]", "tested1")
	form.Add("Objs[0][Sex]", "tested2")
	err := Unmarshal(&obj, form, true)
	fmt.Println(obj, form, err)
}

type TestedUser struct {
	Id            int64  `xorm:"'id' pk autoincr"`
	Name          string `xorm:"unique"`
	NickName      string
	Password      string `goblet:",md5"`
	Status        int
	StatusMessage string
	OrgId         int64
}

type TestedTypeForShort struct {
	Id   string
	Name string
	Sex  string
}

func TestUnmarshalEmptyPassword(t *testing.T) {
	var obj TestedUser
	var form = make(url.Values)
	form.Set("Name", "man")
	form.Set("Password", "")
	err := Unmarshal(&obj, form, true)
	fmt.Println(obj, err)
}

func TestUnmarshalShownFieldAfterUnshownLongNameFiled(t *testing.T) {
	var obj TestedTypeForShort
	var form = make(url.Values)
	form.Set("Sex", "man")
	err := Unmarshal(&obj, form, true)
	fmt.Println(obj, err)
}

func TestUnmarshalString(t *testing.T) {
	var str string
	var form = make(url.Values)
	form.Set("", "man")
	err := Unmarshal(&str, form, true)
	if str != "man" {
		t.Fatal("Unmarshal failed")
	}
	fmt.Println(str, err)
}

func concatPrefix(prefix, tag string) string {
	return prefix + "[" + tag + "]"
}

func Unmarshal(v interface{}, form url.Values, autofill bool) error {

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
		ValuesGetter: func(prefix string) url.Values {
			values := (*map[string][]string)(&form)
			var sub = make(url.Values)
			if values != nil {
				for k, v := range *values {
					if strings.HasPrefix(k, prefix+"[") {
						sub[k] = v
					}
				}
			}
			return sub
		},
		ValueGetter: func(tag string) []string {
			values := (*map[string][]string)(&form)
			if values != nil {
				var lower = strings.ToLower(tag)
				if results, ok := (*values)[tag]; ok {
					return results
				}
				if results, ok := (*values)[lower]; ok {
					return results
				}
				if results, ok := (*values)[tag+"[]"]; ok {
					return results
				}
				if results, ok := (*values)[lower+"[]"]; ok {
					return results
				}
			}
			return []string{}
		},
		TagConcatter: concatPrefix,
		BaseName: func(path string, prefix string) string {
			return strings.Split(strings.TrimPrefix(path, prefix+"["), "]")[0]
		},
		AutoFill: autofill,
	}

	return unmarshaller.Unmarshall(v)
}

type TestedTypeForTime struct {
	Start time.Time `goblet:",fillby(2006-01-02 15:04)"`
	End   time.Time `goblet:",fillby(2006-01-02 15:04)"`
}

func TestUnmarshalTime(t *testing.T) {
	var obj TestedTypeForTime
	var form = make(url.Values)
	form.Set("Start", "2019-05-09 15:35")
	form.Set("End", "2019-05-10 03:30")
	err := Unmarshal(&obj, form, true)
	fmt.Println(obj, err)
}
