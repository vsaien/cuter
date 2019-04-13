package mapping

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/vsaien/cuter/common/stringx"
)

const (
	defaultOption   = "default"
	stringOption    = "string"
	optionalOption  = "optional"
	optionsOption   = "options"
	optionSeparator = "|"
)

var (
	errUnsupportedType = errors.New("unsupported type on setting field value")
	errNotSettable     = errors.New("target not settable")
	optionsCache       = make(map[string]*optionsCacheValue)
	cacheLock          sync.RWMutex
)

type optionsCacheValue struct {
	key     string
	options *FieldOptions
	err     error
}

func Deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t
}

func ParseKeyAndOptions(tagName string, field reflect.StructField) (string, *FieldOptions, error) {
	value := field.Tag.Get(tagName)
	if len(value) == 0 {
		return field.Name, nil, nil
	} else {
		cacheLock.RLock()
		cache, ok := optionsCache[value]
		cacheLock.RUnlock()

		if ok {
			return stringx.TakeOne(cache.key, field.Name), cache.options, cache.err
		} else {
			key, options, err := doParseKeyAndOptions(field, value)
			cacheLock.Lock()
			optionsCache[value] = &optionsCacheValue{
				key:     key,
				options: options,
				err:     err,
			}
			cacheLock.Unlock()
			return stringx.TakeOne(key, field.Name), options, err
		}
	}
}

func SetValue(kind reflect.Kind, value reflect.Value, str string) error {
	if !value.CanSet() {
		return errNotSettable
	}

	switch kind {
	case reflect.Bool:
		value.SetBool(str == "1" || strings.ToLower(str) == "true")
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if intValue, err := strconv.ParseInt(str, 10, 64); err != nil {
			return fmt.Errorf("error: the value %q cannot parsed as int", str)
		} else {
			value.SetInt(intValue)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if uintValue, err := strconv.ParseUint(str, 10, 64); err != nil {
			return fmt.Errorf("error: the value %q cannot parsed as uint", str)
		} else {
			value.SetUint(uintValue)
		}
	case reflect.Float32, reflect.Float64:
		if floatValue, err := strconv.ParseFloat(str, 64); err != nil {
			return fmt.Errorf("error: the value %q cannot parsed as float", str)
		} else {
			value.SetFloat(floatValue)
		}
	case reflect.String:
		value.SetString(str)
	default:
		return errUnsupportedType
	}

	return nil
}

func ValidatePtr(v *reflect.Value) error {
	// sequence is very important, IsNil must be called after checking Kind() with reflect.Ptr,
	// panic otherwise
	if !v.IsValid() || v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("error: not a valid pointer: %v", v)
	}

	return nil
}

func doParseKeyAndOptions(field reflect.StructField, value string) (string, *FieldOptions, error) {
	segments := strings.Split(value, ",")
	key := strings.TrimSpace(segments[0])
	options := segments[1:]

	if len(options) > 0 {
		var fieldOptions FieldOptions

		for _, segment := range options {
			option := strings.TrimSpace(segment)
			switch {
			case option == stringOption:
				fieldOptions.FromString = true
			case option == optionalOption:
				fieldOptions.Optional = true
			case strings.Index(option, optionsOption) == 0:
				segs := strings.Split(option, "=")
				if len(segs) != 2 {
					return "", nil, fmt.Errorf("field %s has wrong options", field.Name)
				} else {
					fieldOptions.Options = strings.Split(segs[1], optionSeparator)
				}
			case strings.Index(option, defaultOption) == 0:
				segs := strings.Split(option, "=")
				if len(segs) != 2 {
					return "", nil, fmt.Errorf("field %s has wrong default option", field.Name)
				} else {
					fieldOptions.Default = strings.TrimSpace(segs[1])
				}
			}
		}

		return key, &fieldOptions, nil
	}

	return key, nil, nil
}

func maybeNewValue(field reflect.StructField, value reflect.Value) {
	if field.Type.Kind() == reflect.Ptr && value.IsNil() {
		value.Set(reflect.New(value.Type().Elem()))
	}
}

func usingDifferentKeys(field reflect.StructField, key string) bool {
	if len(field.Tag) > 0 {
		if _, ok := field.Tag.Lookup(key); !ok {
			return true
		}
	}

	return false
}
