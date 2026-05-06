package structx

import (
	"fmt"
	"reflect"
)

// DiffStruct 比较 dst 与 src，返回值不相同的字段 map 和字段名列表。
// 支持指针、嵌套 struct。
func DiffStruct(dst, src any) (map[string]any, []string, error) {
	_, dstVal, err := validateStructInput(dst)
	if err != nil {
		return nil, nil, fmt.Errorf("dst: %w", err)
	}
	_, srcVal, err := validateStructInput(src)
	if err != nil {
		return nil, nil, fmt.Errorf("src: %w", err)
	}

	// 校验类型一致
	dstType := reflect.TypeOf(dst)
	srcType := reflect.TypeOf(src)
	if dstType.Kind() == reflect.Ptr {
		dstType = dstType.Elem()
	}
	if srcType.Kind() == reflect.Ptr {
		srcType = srcType.Elem()
	}
	if dstType != srcType {
		return nil, nil, fmt.Errorf("dst and src must be same type")
	}

	result := make(map[string]any)
	fields := make([]string, 0)

	for i := 0; i < dstType.NumField(); i++ {
		dstField := dstVal.Field(i)
		srcField := srcVal.Field(i)

		if !dstField.CanInterface() || !srcField.CanInterface() {
			continue
		}

		dstConverted := fieldToMapValue(dstField, 0)
		srcConverted := fieldToMapValue(srcField, 0)

		if !reflect.DeepEqual(dstConverted, srcConverted) {
			result[dstType.Field(i).Name] = dstConverted
			fields = append(fields, dstType.Field(i).Name)
		}
	}

	return result, fields, nil
}

// Assign 将 src 中非零值的字段赋值给 dst。dst 必须是指针。
// 嵌套 struct 整体替换，不做字段级合并。
func Assign(dst, src any) error {
	dstT := reflect.TypeOf(dst)
	dstV := reflect.ValueOf(dst)

	if dstT == nil || dstT.Kind() != reflect.Ptr {
		return fmt.Errorf("dst must be a pointer, got %T", dst)
	}
	if dstV.IsNil() {
		return fmt.Errorf("dst is nil pointer")
	}

	dstT = dstT.Elem()
	dstV = dstV.Elem()

	if dstT.Kind() != reflect.Struct {
		return fmt.Errorf("dst elem must be a struct, got %q", dstT.Kind().String())
	}

	srcT := reflect.TypeOf(src)
	srcV := reflect.ValueOf(src)

	if srcT == nil {
		return fmt.Errorf("src is nil")
	}

	if srcT.Kind() == reflect.Ptr {
		if srcV.IsNil() {
			return fmt.Errorf("src is nil pointer")
		}
		srcT = srcT.Elem()
		srcV = srcV.Elem()
	}

	if srcT.Kind() != reflect.Struct {
		return fmt.Errorf("src elem must be a struct, got %q", srcT.Kind().String())
	}

	if dstT != srcT {
		return fmt.Errorf("dst and src must be same type")
	}

	for i := 0; i < dstT.NumField(); i++ {
		dstField := dstV.Field(i)
		srcField := srcV.Field(i)

		if !dstField.CanSet() {
			continue
		}

		if srcField.IsZero() {
			continue
		}

		dstField.Set(srcField)
	}

	return nil
}

const maxDepth = 10

// dereference 安全解引用指针。如果是指针且为 nil，返回零值 + true。
func dereference(v reflect.Value) (reflect.Value, bool) {
	if !v.IsValid() {
		return reflect.Value{}, false
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Zero(v.Type().Elem()), true
		}
		return v.Elem(), false
	}
	return v, false
}

// safeInterface 将 reflect.Value 安全转为 interface{}。
func safeInterface(v reflect.Value) interface{} {
	if !v.CanInterface() {
		return nil
	}
	derefV, isNil := dereference(v)
	if isNil {
		return nil
	}
	return derefV.Interface()
}

// fieldToMapValue 将单个字段值转为 map 可存储格式，递归处理复杂类型。
func fieldToMapValue(v reflect.Value, depth int) interface{} {
	if depth > maxDepth {
		return nil
	}

	derefV, isNil := dereference(v)
	if isNil {
		return nil
	}

	switch derefV.Kind() {
	case reflect.Struct:
		return structToMapValue(derefV, depth+1)
	case reflect.Slice, reflect.Array:
		if derefV.IsNil() {
			return nil
		}
		result := make([]any, derefV.Len())
		for i := 0; i < derefV.Len(); i++ {
			elem := derefV.Index(i)
			result[i] = fieldToMapValue(elem, depth+1)
		}
		return result
	case reflect.Map:
		if derefV.IsNil() {
			return nil
		}
		result := make(map[string]any)
		for _, key := range derefV.MapKeys() {
			val := derefV.MapIndex(key)
			keyStr := fmt.Sprintf("%v", key.Interface())
			result[keyStr] = fieldToMapValue(val, depth+1)
		}
		return result
	default:
		return safeInterface(derefV)
	}
}

// structToMapValue 核心递归：将 struct 转为 map[string]any。
func structToMapValue(sv reflect.Value, depth int) map[string]any {
	result := make(map[string]any)
	st := sv.Type()

	for i := 0; i < st.NumField(); i++ {
		fieldType := st.Field(i)
		fieldValue := sv.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		result[fieldType.Name] = fieldToMapValue(fieldValue, depth)
	}

	return result
}

// validateStructInput 校验输入必须是指针或 struct。
func validateStructInput(v any) (reflect.Type, reflect.Value, error) {
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)

	if t == nil {
		return nil, reflect.Value{}, fmt.Errorf("input is nil")
	}

	if t.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil, reflect.Value{}, fmt.Errorf("input is nil pointer")
		}
		t = t.Elem()
		val = val.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil, reflect.Value{}, fmt.Errorf("input must be a struct or pointer to struct, got %q", t.Kind().String())
	}

	return t, val, nil
}

// StructToMap 将 struct 转换为 map[string]any，key 为字段名。
func StructToMap(v any) (map[string]any, error) {
	_, val, err := validateStructInput(v)
	if err != nil {
		return nil, err
	}
	return structToMapValue(val, 0), nil
}
