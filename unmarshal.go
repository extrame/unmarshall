package unmarshall

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Unmarshaller for specified struct, ValueGetter for how to get the value, it get the tag as input and return the value
type Unmarshaller struct {
	Values func() map[string][]string
	//return values and if the result is array and should be filled in the struct by offset
	ValueGetter  func(string) []string
	ValuesGetter func(prefix string) url.Values
	TagConcatter func(string, string) string
	BaseName     func(string, string) string
	// FileGetter            func(string) (multipart.File, *multipart.FileHeader, error)
	FillForSpecifiledType map[string]func(string) (reflect.Value, error)
	AutoFill              bool
	MaxLength             int
	Tag                   string //the tag name of action control,the value is seperated by ',', first value is the overrided key, second is the default value
	DefaultTag            string //the tag name to get the default value, if DefaultTag is not empty, the default value will got by the tag, or get from the second value of 'Tag' marked tag
	ArrayValueGotByOffset bool   //if the value got from ValueGetter is array, it will be filled in the struct by offset
}

func (u *Unmarshaller) Unmarshall(v interface{}) error {
	if u.MaxLength == 0 {
		u.MaxLength = 100
	}
	if u.Tag == "" {
		u.Tag = "unmarshal"
	}
	// check v is valid
	rv := reflect.ValueOf(v).Elem()
	// dereference pointer
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		// for each struct field on v
		u.unmarshalStructInForm("", rv, 0, 0, false)
	} else if rv.Kind() == reflect.Map {
		kType := rv.Type().Key()
		vType := rv.Type().Elem()
		if kType.Kind() == reflect.String && vType.Kind() == reflect.Interface {
			values := u.Values()
			for key, value := range values {
				vValue := reflect.ValueOf(value)
				if vValue.Kind() == reflect.Slice || vValue.Kind() == reflect.Array {
					if vValue.Len() == 1 {
						rv.SetMapIndex(reflect.ValueOf(key), vValue.Index(0))
						continue
					}
				}
				rv.SetMapIndex(reflect.ValueOf(key), reflect.ValueOf(value))
			}
		}
	} else if rv.Kind() == reflect.Slice {
		id, form_values, _, extraTags, err := u.getFormField("", func(key string) string { return "" }, "", 0, false)
		if err == nil {
			if err = u.unmarshallSlice("", rv.Type(), 0, form_values, rv, id, 0, extraTags); err != nil {
				return err
			}
		}
	} else if rv.Kind() == reflect.String {
		tempVal := reflect.New(rv.Type())
		_, form_values, _, extraTags, err := u.getFormField("", func(key string) string { return "" }, "", 0, false)
		if err == nil {
			if len(form_values) > 0 {
				u.unmarshalField("", tempVal.Elem(), form_values[0], extraTags, false)
			}
			if tempVal.Elem().String() != "" {
				rv.Set(tempVal.Elem())
			}
		}
	} else {
		return fmt.Errorf("v must point to a struct type")
	}
	return nil
}

func (u *Unmarshaller) unmarshalStructInForm(context string,
	rvalue reflect.Value,
	offset int,
	deep int,
	inarray bool) (thisObjectIsNotEmpty bool, err error) {

	if rvalue.Type().Kind() == reflect.Ptr {
		rvalue = rvalue.Elem()
	}
	rtype := rvalue.Type()

	for i := 0; i < rtype.NumField() && err == nil; i++ {
		var id, defaultVal string
		var form_values, extraTags []string
		var rField = rtype.Field(i)
		if rField.Anonymous {
			u.unmarshalStructInForm(id, rvalue.Field(i), offset, deep, false)
			continue
		}
		id, form_values, defaultVal, extraTags, err = u.getFormField(context, rField.Tag.Get, rField.Name, offset, inarray)
		if err == TooDeepErr {
			err = nil
			continue
		}
		if err != nil {
			return
		}
		thisObjectIsNotEmpty = thisObjectIsNotEmpty || len(form_values) > 0
		increaseOffset := !(context != "" && inarray)
		var used_offset = 0
		if increaseOffset && !u.ArrayValueGotByOffset {
			used_offset = offset
		}
		if rvalue.Field(i).CanSet() {
			switch rField.Type.Kind() {
			case reflect.Ptr: //TODO if the ptr point to a basic data, it will crash
				val := rvalue.Field(i)
				typ := rField.Type.Elem()
				tempVal := reflect.New(typ)
				switch typ.Kind() {
				case reflect.Struct:
					var childIsNotEmpty bool
					if childIsNotEmpty, err = u.fill_struct(typ,
						tempVal.Elem(),
						id, form_values, extraTags, used_offset, deep+1); err == nil && childIsNotEmpty {
						// 	return false, err
						// } else {
						val.Set(tempVal)
						thisObjectIsNotEmpty = thisObjectIsNotEmpty || childIsNotEmpty
					}
				case reflect.String, reflect.Bool:
					if len(form_values) > 0 && used_offset < len(form_values) {
						u.unmarshalField(context, tempVal.Elem(), form_values[used_offset], extraTags, false)
						thisObjectIsNotEmpty = true
					} else if defaultVal != "" {
						u.unmarshalField(context, tempVal.Elem(), defaultVal, extraTags, true)
						if !inarray { // if in array, the rule to judge if the object is empty is exclude the default value
							thisObjectIsNotEmpty = true
						}
					}
					if thisObjectIsNotEmpty && tempVal.Elem().String() != "" {
						val.Set(tempVal)
					}
				}

				//忽略可能的设置错误，进行到下一个
				err = nil
				continue
			case reflect.Struct:
				var childIsNotEmpty bool
				if childIsNotEmpty, err = u.fill_struct(rField.Type, rvalue.Field(i), id, form_values, extraTags, used_offset, deep+1); childIsNotEmpty && err != nil {
					return childIsNotEmpty, err
				} else {
					thisObjectIsNotEmpty = thisObjectIsNotEmpty || childIsNotEmpty
					continue
				}
			case reflect.Interface:
				//ask the parent to tell me how to unmarshal it
				mtd := rvalue.MethodByName("UnmarshallForm")
				if mtd.IsValid() {
					values := mtd.Call([]reflect.Value{reflect.ValueOf(rField.Name)})
					if len(values) == 2 && values[1].Interface() == nil {
						res := values[0].Interface()
						resValue := reflect.ValueOf(res)
						resType := reflect.TypeOf(res)
						if thisObjectIsNotEmpty, err = u.fill_struct(resType, resValue, id, form_values, extraTags, used_offset, deep+1); err != nil {
							rvalue.Field(i).Set(resValue)
							return thisObjectIsNotEmpty, err
						} else {
							continue
						}
					}
				} else {
					return false, fmt.Errorf("try to use UnmarshallForm to unmarshall interface type(%T) fail", rvalue.Interface())
				}
			case reflect.Slice:
				if err := u.unmarshallSlice(context, rField.Type, used_offset, form_values, rvalue.Field(i), id, deep, extraTags); err == nil {
					continue
				} else {
					return thisObjectIsNotEmpty, err
				}
			case reflect.Map:
				err := u.unmarshallMap(id, rvalue.Field(i), extraTags, deep)
				if err != nil {
					return thisObjectIsNotEmpty, errors.Wrap(err, "in unmarshall map")
				}
			default:
				if len(form_values) > 0 && used_offset < len(form_values) {
					u.unmarshalField(context, rvalue.Field(i), form_values[used_offset], extraTags, false)
					thisObjectIsNotEmpty = true
				} else if defaultVal != "" && !inarray {
					u.unmarshalField(context, rvalue.Field(i), defaultVal, extraTags, true)
					thisObjectIsNotEmpty = true
				}
			}
		} else if rField.IsExported() {
			return thisObjectIsNotEmpty, fmt.Errorf("cannot set value of (%s,%s) in fill", rField.Name, rField.Type.Name())
		} // else ignore the unexported field
	}
	return
}

func (u *Unmarshaller) unmarshallSlice(context string, fType reflect.Type, used_offset int,
	form_values []string, target reflect.Value, id string, deep int, extraTags []string) error {
	subRType := fType.Elem()
	//net.IP alias of []byte
	if used_offset < len(form_values) {
		if fType.PkgPath() == "net" && fType.Name() == "IP" {
			target.Set(reflect.ValueOf(net.ParseIP(form_values[used_offset])))
			return nil
		} else if subRType.Kind() == reflect.Uint8 {
			target.SetBytes([]byte(form_values[used_offset]))
		}
	}

	switch subRType.Kind() {
	case reflect.Struct:
		// if lastDeep, ok := parents[subRType.PkgPath()+"/"+subRType.Name()]; !ok || lastDeep == deep {
		rvalueTemp := reflect.MakeSlice(fType, 0, 0)
		offset := 0
		for {
			subRValue := reflect.New(subRType)
			isNotEmpty, _ := u.unmarshalStructInForm(id, subRValue, offset, deep+1, true)
			if !isNotEmpty {
				break
			}
			offset++
			rvalueTemp = reflect.Append(rvalueTemp, subRValue.Elem())
		}
		if rvalueTemp.Len() > 0 {
			target.Set(rvalueTemp)
		}
		// } else {
		// 	err = fmt.Errorf("Too deep of type reuse %v", parents)
		// }
	case reflect.Ptr:
		if subRType.Elem().Kind() == reflect.Struct {
			var elemType = subRType.Elem()
			// if lastDeep, ok := parents[elemType.PkgPath()+"/"+elemType.Name()]; !ok || lastDeep == deep {
			rvalueTemp := reflect.MakeSlice(fType, 0, 0)
			offset := 0
			for {
				subRValue := reflect.New(elemType)
				//依靠下层返回进行终止
				isNotEmpty, err := u.unmarshalStructInForm(id, subRValue, offset, deep+1, true)
				if !isNotEmpty {
					break
				}
				if err != nil {
					return errors.Wrap(err, "unmarshall []*struct err ")
				}
				offset++
				rvalueTemp = reflect.Append(rvalueTemp, subRValue)
			}
			if rvalueTemp.Len() > 0 {
				target.Set(rvalueTemp)
			}
			// } else {
			// 	err = fmt.Errorf("Too deep of type reuse %v,%T,%d", parents, elemType.PkgPath()+"/"+elemType.Name(), deep)
			// }
		}
	default:
		lenFv := len(form_values)
		rvnew := reflect.MakeSlice(fType, lenFv, lenFv)
		for j := 0; j < lenFv; j++ {
			u.unmarshalField(context, rvnew.Index(j), form_values[j], extraTags, false)
		}
		target.Set(rvnew)
	}
	return nil
}

var TooDeepErr = errors.New("too deep")

func (u *Unmarshaller) getFormField(prefix string, tagGetter func(key string) string, name string, offset int, inarray bool) (string, []string, string, []string, error) {

	tag, tags := u.getTag(prefix, tagGetter, name, offset, inarray)

	if len(tag) > u.MaxLength {
		return "", nil, "", nil, TooDeepErr
	}
	var defaultVal string
	if u.DefaultTag != "" {
		defaultVal = tagGetter(u.DefaultTag)
	} else if len(tags) > 1 {
		defaultVal = tags[1]
	}

	values := u.ValueGetter(tag)

	return tag, values, defaultVal, tags[1:], nil
}

func (u *Unmarshaller) getTag(prefix string,
	t func(key string) string, name string, offset int, inarray bool) (string, []string) {
	tags := []string{""}
	tag := t(u.Tag)
	if tag != "" {
		tags = strings.Split(tag, ",")
		tag = tags[0]
	}
	if tag == "" {
		tag = name
	}

	if inarray {
		tag = u.TagConcatter(fmt.Sprintf(prefix+"[%d]", offset), tag)
	} else if prefix != "" {
		tag = u.TagConcatter(prefix, tag)
	}
	return tag, tags
}

func (u *Unmarshaller) unmarshalField(contex string, v reflect.Value, form_value string, tags []string, forFill bool) error {

	if fn, ok := u.FillForSpecifiledType[v.Type().PkgPath()+"."+v.Type().Name()]; ok {
		var err error
		var nv reflect.Value
		if nv, err = fn(form_value); err == nil {
			v.Set(nv)
		}
		return err
	}

	switch v.Kind() {
	case reflect.Int64:
		if i, err := strconv.ParseInt(form_value, 10, 64); err == nil {
			v.SetInt(i)
		}
	case reflect.Uint64:
		if i, err := strconv.ParseUint(form_value, 10, 64); err == nil {
			v.SetUint(i)
		}
	case reflect.Int, reflect.Int32:
		if i, err := strconv.ParseInt(form_value, 10, 32); err == nil {
			v.SetInt(i)
		}
	case reflect.Uint32:
		if i, err := strconv.ParseUint(form_value, 10, 32); err == nil {
			v.SetUint(i)
		}
	case reflect.Int16:
		if i, err := strconv.ParseInt(form_value, 10, 16); err == nil {
			v.SetInt(i)
		}
	case reflect.Uint16:
		if i, err := strconv.ParseUint(form_value, 10, 16); err == nil {
			v.SetUint(i)
		}
	case reflect.Int8:
		if i, err := strconv.ParseInt(form_value, 10, 8); err == nil {
			v.SetInt(i)
		}
	case reflect.Uint8:
		if i, err := strconv.ParseUint(form_value, 10, 8); err == nil {
			v.SetUint(i)
		}
	case reflect.String:
		// copy string
		if len(tags) > 0 && tags[len(tags)-1] == "md5" && form_value != "" {
			if !(forFill && len(tags) == 1) {
				h := md5.New()
				h.Write([]byte(form_value))
				v.SetString(hex.EncodeToString(h.Sum(nil)))
			}
		} else {
			v.SetString(form_value)
		}
	case reflect.Float64:
		if f, err := strconv.ParseFloat(form_value, 64); err == nil {
			v.SetFloat(f)
		}
	case reflect.Float32:
		if f, err := strconv.ParseFloat(form_value, 32); err == nil {
			v.SetFloat(f)
		}
	case reflect.Bool:
		// the following strings convert to true
		// 1,true,True,on,yes
		fv := form_value
		if fv == "1" || fv == "true" || fv == "True" || fv == "on" || fv == "yes" {
			v.SetBool(true)
		} else if fv == "0" || fv == "false" || fv == "False" || fv == "off" || fv == "no" {
			v.SetBool(false)
		}
	default:
		fmt.Println("unknown type", v.Kind())
	}
	return nil
}

func (u *Unmarshaller) fill_struct(typ reflect.Type,
	val reflect.Value, id string, form_values []string, tag []string, used_offset int, deep int) (bool, error) {
	if typ.PkgPath() == "time" && typ.Name() == "Time" {
		var fillby string
		var fillby_valid = regexp.MustCompile(`^\s*fillby\((.*)\)\s*$`)
		for _, v := range tag {
			matched := fillby_valid.FindStringSubmatch(v)
			if len(matched) == 2 {
				fillby = matched[1]
			}
		}
		fillby = strings.TrimSpace(fillby)
		var value string
		if len(form_values) > used_offset {
			value = form_values[used_offset]
		}
		switch fillby {
		case "now":
			val.Set(reflect.ValueOf(time.Now()))
		case "timestamp":
			if unix, err := strconv.ParseInt(value, 10, 64); err == nil {
				val.Set(reflect.ValueOf(time.Unix(unix, 0)))
			} else {
				return false, err
			}
		default:
			if fillby == "" {
				fillby = time.RFC3339
			}
			if value != "" {
				time, err := time.ParseInLocation(fillby, value, time.Local)
				if err == nil {
					val.Set(reflect.ValueOf(time))
				} else {
					return false, err
				}
			}
		}
	} else {
		for k, fn := range u.FillForSpecifiledType {
			if typ.PkgPath()+"."+typ.Name() == k {
				if v, err := fn(id); err == nil {
					val.Set(v)
					return true, nil
				} else {
					return false, err
				}
				//if has, set,if not ,return err and the upper will do nothing
			}
		}
		isNotEmpty, err := u.unmarshalStructInForm(id, val, 0, deep, false)
		// if isNotEmpty && err != nil {
		return isNotEmpty, err
		// }

	}
	return true, nil
}

func (u *Unmarshaller) unmarshallMap(id string, mapValue reflect.Value, tag []string, deep int) error {
	var maps = make(map[string]bool)
	var sub url.Values
	if u.ValuesGetter != nil {
		sub = u.ValuesGetter(id)
	}
	if len(sub) == 0 {
		return nil
	}
	for k := range sub {
		//
		subName := u.BaseName(k, id)
		if _, ok := maps[subName]; !ok {
			var newKValue = reflect.New(mapValue.Type().Key())
			var fullName = u.TagConcatter(id, subName)
			err := u.unmarshalField(fullName, newKValue.Elem(), subName, tag, false)
			if err == nil {
				subRType := mapValue.Type().Elem()
				subRValue := reflect.New(subRType)
				switch subRType.Kind() {
				case reflect.Struct:
					isNotEmpty, err := u.unmarshalStructInForm(fullName, subRValue, 0, deep+1, false)
					if isNotEmpty && err != nil { //非空还出错了
						return err
					} else if !isNotEmpty {
						continue
					}
				case reflect.Ptr:
					if subRType.Elem().Kind() == reflect.Struct {
						var elemType = subRType.Elem()
						// if lastDeep, ok := parents[elemType.PkgPath()+"/"+elemType.Name()]; !ok || lastDeep == deep {
						subElemValue := reflect.New(elemType)
						//依靠下层返回进行终止
						isNotEmpty, err := u.unmarshalStructInForm(fullName, subElemValue, 0, deep+1, false)
						if isNotEmpty && err != nil { //非空还出错了
							return err
						} else if !isNotEmpty {
							continue
						}
						subRValue.Elem().Set(subElemValue)
					}
				case reflect.Map:
					err := u.unmarshallMap(fullName, subRValue.Elem(), tag, deep+1)
					if err != nil {
						return err
					}
				default:
					form_values := u.ValueGetter(fullName)
					if len(form_values) > 0 {
						u.unmarshalField(fullName, subRValue.Elem(), form_values[0], tag, false)
					} else {
						return fmt.Errorf("%s[%s]has no value", id, subName)
					}
				}
				if mapValue.IsNil() {
					mapValue.Set(reflect.MakeMap(mapValue.Type()))
				}
				mapValue.SetMapIndex(newKValue.Elem(), subRValue.Elem())
			}
			maps[subName] = true
		}
	}
	return nil
}
