package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// WriteConfig는 현재 GlobalConfig를 지정 경로에 YAML로 저장합니다.
// 부모 디렉터리가 없으면 생성하며, 존재하는 파일은 덮어씁니다.
func WriteConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := marshalYAML(GlobalConfig)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// DumpConfig는 현재 GlobalConfig를 사람이 읽을 수 있는 YAML 문자열로 반환합니다.
func DumpConfig() (string, error) {
	data, err := marshalYAML(GlobalConfig)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// SetConfig는 wconf 태그 키로 식별되는 설정 항목의 값을 변경합니다.
// 값은 필드 타입에 맞게 파싱되며, 알 수 없는 키나 잘못된 값이면 에러를 반환합니다.
func SetConfig(key, value string) error {
	v := reflect.ValueOf(GlobalConfig).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := t.Field(i)

		tag := ft.Tag.Get("wconf")
		if tag == "" {
			tag = ft.Name
		}
		if tag != key {
			continue
		}
		if !f.CanSet() {
			return fmt.Errorf("field %s is not settable", key)
		}
		if err := setFieldByType(f, ft, value); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("unknown config key: %s", key)
}

// setFieldByType는 문자열 값을 필드의 타입에 맞게 파싱해 설정합니다.
func setFieldByType(f reflect.Value, ft reflect.StructField, value string) error {
	switch f.Kind() {
	case reflect.String:
		f.SetString(value)
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid value %q for %s: %w", value, ft.Name, err)
		}
		f.SetFloat(parsed)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if ft.Type == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration %q for %s: %w", value, ft.Name, err)
			}
			f.SetInt(int64(d))
			return nil
		}
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid value %q for %s: %w", value, ft.Name, err)
		}
		f.SetInt(parsed)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid value %q for %s: %w", value, ft.Name, err)
		}
		f.SetUint(parsed)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid bool %q for %s: %w", value, ft.Name, err)
		}
		f.SetBool(parsed)
	default:
		return fmt.Errorf("unsupported type %s for field %s", f.Kind(), ft.Name)
	}
	return nil
}

// marshalYAML은 설정 구조체를 YAML 형태의 바이트로 직렬화합니다.
// 값이 비어 있는(빈 문자열) 필드는 유효하지 않은 값으로 보고 생략합니다.
func marshalYAML(target any) ([]byte, error) {
	v := reflect.ValueOf(target)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("config target must be a struct")
	}

	var sb strings.Builder
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		ft := t.Field(i)

		key := ft.Tag.Get("wconf")
		if key == "" {
			key = ft.Name
		}
		val, ok := fieldValueString(f, ft)
		if !ok || val == "" {
			continue
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", key, val))
	}
	return []byte(sb.String()), nil
}

// fieldValueString은 필드 값을 설정 파일 표현 문자열로 변환합니다.
// 표현할 수 없는 타입이면 ok=false를 반환합니다.
func fieldValueString(f reflect.Value, ft reflect.StructField) (string, bool) {
	switch f.Kind() {
	case reflect.String:
		return f.String(), true
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(f.Float(), 'f', -1, 64), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if ft.Type == reflect.TypeOf(time.Duration(0)) {
			return time.Duration(f.Int()).String(), true
		}
		return strconv.FormatInt(f.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(f.Uint(), 10), true
	case reflect.Bool:
		return strconv.FormatBool(f.Bool()), true
	default:
		return "", false
	}
}
