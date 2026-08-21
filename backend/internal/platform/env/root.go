package env

import (
	"errors"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	seconds = "s"
	minute  = "m"
	hour    = "h"
	day     = "d"
)

type env struct {
	key   string
	value string
}

func search(filename string) (bool, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return false, err
	}

	dirs, err := os.ReadDir(currentDir)
	if err != nil {
		return false, err
	}

	for _, data := range dirs {
		if !data.IsDir() && data.Name() == filename {
			return true, nil
		}
	}

	return false, errors.New("filename not found")
}

func validates(data string) []env {
	envs := strings.Split(data, "\n")
	valids := make([]env, 0, len(envs))

	for _, data := range envs {
		data = strings.TrimSpace(data)
		key, value, found := strings.Cut(data, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" || value == "" {
			continue
		}

		valueLen := len(value) - 1

		if value[0] == '"' && value[valueLen] == '"' {
			value = value[1:valueLen]
		}

		key = strings.ToUpper(key)

		valids = append(valids, env{key, value})
	}

	return valids
}

func Load(filenames ...string) error {
	filename := ".env"

	if len(filenames) != 0 {
		filename = filenames[0]
	}

	exist, err := search(filename)
	if err != nil {
		return err
	}

	if !exist {
		return errors.New("file not exist.")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	envs := validates(strings.TrimSpace(string(data)))

	for _, env := range envs {
		if err := os.Setenv(env.key, env.value); err != nil {
			return err
		}
	}

	return nil
}

type FieldParams struct {
	KeyName         string
	Key             string
	DefaultValue    string
	HasDefaultValue bool
	Required        bool
}

func getKeyTags(v string) (string, []string) {
	opts := strings.Split(v, ",")

	return opts[0], opts[1:]
}

func getFieldParams(f reflect.StructField) (FieldParams, error) {
	value, ok := f.Tag.Lookup("env")
	if !ok {
		return FieldParams{}, errors.New("tag env no exist")
	}

	defaultValue, hasDefaultValue := f.Tag.Lookup("default")

	name, tags := getKeyTags(value)

	field := FieldParams{
		Key:             f.Name,
		KeyName:         name,
		DefaultValue:    defaultValue,
		HasDefaultValue: hasDefaultValue,
	}

	for _, tag := range tags {
		switch tag {
		case "required":
			field.Required = true
		}
	}

	return field, nil
}

func getString(key, defaultValue string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value != "" {
		return value
	}

	return defaultValue
}

func convertDuration(value string) time.Duration {
	value = strings.TrimSpace(value)

	switch {
	case strings.Contains(value, seconds):
		strNS, _ := strings.CutSuffix(value, seconds)
		ns, _ := strconv.ParseInt(strNS, 10, 64)

		return time.Second * time.Duration(ns)
	case strings.Contains(value, minute):
		strNM, _ := strings.CutSuffix(value, minute)
		nm, _ := strconv.ParseInt(strNM, 10, 64)

		return time.Minute * time.Duration(nm)
	case strings.Contains(value, hour):
		strNH, _ := strings.CutSuffix(value, hour)
		nh, _ := strconv.ParseInt(strNH, 10, 64)

		return time.Hour * time.Duration(nh)
	case strings.Contains(value, day):
		strND, _ := strings.CutSuffix(value, day)
		nd, _ := strconv.ParseInt(strND, 10, 64)

		return time.Hour * 24 * time.Duration(nd)
	}

	return time.Duration(0)
}

func getSliceString(key, defaultValue string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return strings.Split(defaultValue, ",")
	}

	return strings.Split(value, ",")
}

func getDuration(key string, defaultValue string) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return convertDuration(defaultValue)
	}

	return convertDuration(value)
}

func getInt64(key string, defaultValue string) int64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		dv, _ := strconv.ParseInt(defaultValue, 10, 64)

		return dv
	}

	int64Value, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		dv, _ := strconv.ParseInt(defaultValue, 10, 64)

		return dv
	}

	return int64Value
}

func getBool(key string, defaultValue string) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		dv, _ := strconv.ParseBool(defaultValue)

		return dv
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		dv, _ := strconv.ParseBool(defaultValue)

		return dv
	}

	return boolValue
}

func Parse(v any) error {
	ptrRef := reflect.ValueOf(v)
	if ptrRef.Kind() != reflect.Pointer {
		return errors.New("no pointer")
	}

	s := ptrRef.Elem()
	if s.Kind() != reflect.Struct {
		return errors.New("no struct")
	}

	t := s.Type()

	for i := range t.NumField() {
		fieldParams, err := getFieldParams(t.Field(i))
		if err != nil {
			return err
		}

		f := s.Field(i)

		switch f.Type() {
		case reflect.TypeFor[string]():
			f.SetString(getString(fieldParams.KeyName, fieldParams.DefaultValue))
		case reflect.TypeFor[int64]():
			f.SetInt(getInt64(fieldParams.KeyName, fieldParams.DefaultValue))
		case reflect.TypeFor[bool]():
			f.SetBool(getBool(fieldParams.KeyName, fieldParams.DefaultValue))
		case reflect.TypeFor[time.Duration]():
			f.SetInt(int64(getDuration(fieldParams.KeyName, fieldParams.DefaultValue)))
		case reflect.TypeFor[[]string]():
			f.Set(reflect.ValueOf(getSliceString(fieldParams.KeyName, fieldParams.DefaultValue)))
		}
	}

	return nil
}
