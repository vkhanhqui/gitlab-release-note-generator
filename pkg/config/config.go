package config

import (
	"log"
	"path"
	"reflect"
	"runtime"
	"strings"

	"gitLab-rls-note/pkg/errors"

	"github.com/spf13/viper"
)

// LoadEnvConfig loads environment variables using viper global instance
func LoadEnvConfig() {
	_, f, _, _ := runtime.Caller(1)
	pwd := path.Dir(f)
	file := path.Join(pwd, ".env")

	viper.SetConfigFile(file)
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("Not loading env from file: %s", err.Error())
	}

	viper.AutomaticEnv()
}

// UnmarshalEnvConfig stores environment variables into a defined struct.
// Supports int, string, bool, []string, map[string]string.
//
// An array env must be defined as: "elem1;elem2;elem3"
//
// A map env must be defined as: "key1:value1;key2:value2;key3:value3"
//
// Note: cannot use v.Unmarshal() because of this https://github.com/spf13/viper/issues/584
func UnmarshalEnvConfig(o interface{}) (err error) {
	if o == nil {
		return
	}

	const tag string = "mapstructure"
	typ := reflect.TypeOf(o).Elem()
	val := reflect.ValueOf(o).Elem()

	if typ.Kind() != reflect.Struct {
		return errors.New("Output must be a struct")
	}

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		if !structField.CanSet() {
			continue
		}

		fieldKind := structField.Kind()
		fieldName := typeField.Tag.Get(tag)

		switch fieldKind {
		case reflect.Struct:
			err = UnmarshalEnvConfig(structField.Addr().Interface())
			if err != nil {
				return
			}

		case reflect.String:
			structField.SetString(viper.GetString(fieldName))
		case reflect.Bool:
			structField.SetBool(viper.GetBool(fieldName))
		case reflect.Int:
			structField.SetInt(int64(viper.GetInt(fieldName)))

		case reflect.Slice:
			_, ok := structField.Interface().([]string)
			if !ok {
				return errors.New("Only support string slice.")
			}

			elems := strings.Split(viper.GetString(fieldName), ";")
			s := make([]string, 0)
			s = append(s, elems...)
			structField.Set(reflect.ValueOf(s))

		case reflect.Map:
			_, ok := structField.Interface().(map[string]string)
			if !ok {
				return errors.New("Only support map[string]string map.")
			}

			m := make(map[string]string)
			elems := strings.Split(viper.GetString(fieldName), ";")
			for _, elem := range elems {
				kv := strings.Split(elem, ":")
				m[kv[0]] = kv[1]
			}
			structField.Set(reflect.ValueOf(m))

		default:
			return errors.Errorf("Unsupported field kind: %s", fieldKind)
		}
	}

	return
}
