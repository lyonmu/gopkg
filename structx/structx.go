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
	dstState := newConversionState(dst)
	srcState := newConversionState(src)

	for i := 0; i < dstType.NumField(); i++ {
		dstField := dstVal.Field(i)
		srcField := srcVal.Field(i)

		if !dstField.CanInterface() || !srcField.CanInterface() {
			continue
		}

		dstConverted := fieldToMapValue(dstField, 0, dstState)
		srcConverted := fieldToMapValue(srcField, 0, srcState)

		if !reflect.DeepEqual(dstConverted, srcConverted) {
			result[dstType.Field(i).Name] = dstConverted
			fields = append(fields, dstType.Field(i).Name)
		}
	}

	return result, fields, nil
}

// assignFields 将 src 中字段赋值给 dst。skipZero 为 true 时跳过零值字段。
// dst 必须是指针。嵌套 struct 整体替换，不做字段级合并。
func assignFields(dst, src any, skipZero bool) error {
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

		if skipZero && srcField.IsZero() {
			continue
		}

		dstField.Set(srcField)
	}

	return nil
}

// AssignNonZero 将 src 中非零值字段赋值给 dst。dst 必须是指针。
// 嵌套 struct 整体替换，不做字段级合并。
// 零值字段（如 0、""、false、nil）不会被赋值，dst 中对应字段保持原值。
func AssignNonZero(dst, src any) error {
	return assignFields(dst, src, true)
}

// AssignOverwrite 将 src 中所有字段赋值给 dst，包括零值字段。dst 必须是指针。
// 嵌套 struct 整体替换，不做字段级合并。
func AssignOverwrite(dst, src any) error {
	return assignFields(dst, src, false)
}

// Deprecated: 使用 AssignNonZero 获取明确语义。
func Assign(dst, src any) error {
	return AssignNonZero(dst, src)
}

const maxDepth = 10

type visit struct {
	typ reflect.Type
	ptr uintptr
}

type conversionState struct {
	visiting map[visit]struct{}
}

func newConversionState(v any) *conversionState {
	state := &conversionState{visiting: make(map[visit]struct{})}
	value := reflect.ValueOf(v)
	if value.IsValid() && value.Kind() == reflect.Ptr && !value.IsNil() {
		state.visiting[visit{typ: value.Type(), ptr: value.Pointer()}] = struct{}{}
	}
	return state
}

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

// hasExportedField 判断 struct 类型是否包含导出字段。
func hasExportedField(t reflect.Type) bool {
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).PkgPath == "" {
			return true
		}
	}
	return false
}

// fieldToMapValue 将单个字段值转为 map 可存储格式，递归处理复杂类型。
func fieldToMapValue(v reflect.Value, depth int, state *conversionState) interface{} {
	if depth > maxDepth {
		return nil
	}
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.Ptr && !v.IsNil() {
		current := visit{typ: v.Type(), ptr: v.Pointer()}
		if _, ok := state.visiting[current]; ok {
			return nil
		}
		state.visiting[current] = struct{}{}
		defer delete(state.visiting, current)
	}

	derefV, isNil := dereference(v)
	if isNil {
		return nil
	}

	switch derefV.Kind() {
	case reflect.Struct:
		if derefV.NumField() == 0 {
			return structToMapValue(derefV, depth+1, state)
		}
		if !hasExportedField(derefV.Type()) {
			return safeInterface(derefV)
		}
		return structToMapValue(derefV, depth+1, state)
	case reflect.Slice:
		if derefV.IsNil() {
			return nil
		}
		result := make([]any, derefV.Len())
		for i := 0; i < derefV.Len(); i++ {
			elem := derefV.Index(i)
			result[i] = fieldToMapValue(elem, depth+1, state)
		}
		return result
	case reflect.Array:
		result := make([]any, derefV.Len())
		for i := 0; i < derefV.Len(); i++ {
			elem := derefV.Index(i)
			result[i] = fieldToMapValue(elem, depth+1, state)
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
			result[keyStr] = fieldToMapValue(val, depth+1, state)
		}
		return result
	default:
		return safeInterface(derefV)
	}
}

// structToMapValue 核心递归：将 struct 转为 map[string]any。
func structToMapValue(sv reflect.Value, depth int, state *conversionState) map[string]any {
	result := make(map[string]any)
	st := sv.Type()

	for i := 0; i < st.NumField(); i++ {
		fieldType := st.Field(i)
		fieldValue := sv.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		result[fieldType.Name] = fieldToMapValue(fieldValue, depth, state)
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
	return structToMapValue(val, 0, newConversionState(v)), nil
}
