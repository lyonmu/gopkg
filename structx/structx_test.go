package structx

import (
	"reflect"
	"testing"
)

func TestStructToMap(t *testing.T) {
	type Inner struct {
		X int
		Y string
	}
	type Outer struct {
		Name   string
		Age    int
		Active bool
		Inner  Inner
		Ptr    *string
		Slice  []int
		MapF   map[string]int
	}

	tests := []struct {
		name    string
		input   any
		want    map[string]any
		wantErr bool
	}{
		{
			name:    "空 struct",
			input:   struct{}{},
			want:    map[string]any{},
			wantErr: false,
		},
		{
			name: "简单类型",
			input: struct {
				A int
				B string
			}{A: 42, B: "hello"},
			want:    map[string]any{"A": 42, "B": "hello"},
			wantErr: false,
		},
		{
			name:    "指针字段 nil",
			input:   struct{ P *int }{P: nil},
			want:    map[string]any{"P": nil},
			wantErr: false,
		},
		{
			name: "指针字段非 nil",
			input: func() any {
				v := 99
				return struct{ P *int }{P: &v}
			}(),
			want:    map[string]any{"P": 99},
			wantErr: false,
		},
		{
			name:    "嵌套 struct",
			input:   Outer{Name: "test", Age: 10, Inner: Inner{X: 1, Y: "inner"}},
			want:    map[string]any{"Name": "test", "Age": 10, "Active": false, "Inner": map[string]any{"X": 1, "Y": "inner"}, "Ptr": nil, "Slice": nil, "MapF": nil},
			wantErr: false,
		},
		{
			name:    "slice 字段",
			input:   struct{ S []int }{S: []int{1, 2, 3}},
			want:    map[string]any{"S": []any{1, 2, 3}},
			wantErr: false,
		},
		{
			name:    "map 字段",
			input:   struct{ M map[string]int }{M: map[string]int{"a": 1}},
			want:    map[string]any{"M": map[string]any{"a": 1}},
			wantErr: false,
		},
		{
			name:    "指针 struct 输入",
			input:   &struct{ A string }{A: "ptr"},
			want:    map[string]any{"A": "ptr"},
			wantErr: false,
		},
		{
			name:    "非 struct 输入",
			input:   "string",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "slice 输入",
			input:   []int{1, 2},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "slice of structs",
			input:   struct{ S []Inner }{S: []Inner{{X: 1, Y: "a"}, {X: 2, Y: "b"}}},
			want:    map[string]any{"S": []any{map[string]any{"X": 1, "Y": "a"}, map[string]any{"X": 2, "Y": "b"}}},
			wantErr: false,
		},
		{
			name:    "map with struct values",
			input:   struct{ M map[string]Inner }{M: map[string]Inner{"k1": {X: 1, Y: "a"}}},
			want:    map[string]any{"M": map[string]any{"k1": map[string]any{"X": 1, "Y": "a"}}},
			wantErr: false,
		},
		{
			name:    "nil pointer input",
			input:   (*struct{ A int })(nil),
			want:    nil,
			wantErr: true,
		},
		{
			name:    "nil interface input",
			input:   nil,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StructToMap(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("StructToMap() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StructToMap() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDiffStruct(t *testing.T) {
	type Inner struct {
		X int
		Y string
	}
	type Outer struct {
		Name  string
		Age   int
		Inner Inner
		Ptr   *string
	}

	str1 := "hello"
	str2 := "world"

	tests := []struct {
		name       string
		dst        any
		src        any
		wantFields []string
		wantValues map[string]any
		wantErr    bool
	}{
		{
			name:       "完全相同",
			dst:        Outer{Name: "test", Age: 10},
			src:        Outer{Name: "test", Age: 10},
			wantFields: []string{},
			wantErr:    false,
		},
		{
			name:       "单字段不同",
			dst:        Outer{Name: "test", Age: 10},
			src:        Outer{Name: "test", Age: 20},
			wantFields: []string{"Age"},
			wantValues: map[string]any{"Age": 10},
			wantErr:    false,
		},
		{
			name:       "嵌套 struct 不同",
			dst:        Outer{Inner: Inner{X: 1}},
			src:        Outer{Inner: Inner{X: 2}},
			wantFields: []string{"Inner"},
			wantErr:    false,
		},
		{
			name:       "指针不同",
			dst:        Outer{Ptr: &str1},
			src:        Outer{Ptr: &str2},
			wantFields: []string{"Ptr"},
			wantErr:    false,
		},
		{
			name:       "指针 nil vs 非 nil",
			dst:        Outer{Ptr: nil},
			src:        Outer{Ptr: &str1},
			wantFields: []string{"Ptr"},
			wantValues: map[string]any{"Ptr": nil},
			wantErr:    false,
		},
		{
			name:       "指针 struct 输入",
			dst:        &Outer{Name: "a"},
			src:        &Outer{Name: "b"},
			wantFields: []string{"Name"},
			wantErr:    false,
		},
		{
			name:    "非 struct dst",
			dst:     "string",
			src:     Outer{},
			wantErr: true,
		},
		{
			name:    "不同类型",
			dst:     Outer{},
			src:     Inner{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotFields, err := DiffStruct(tt.dst, tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("DiffStruct() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if len(gotFields) != len(tt.wantFields) {
				t.Errorf("DiffStruct() fields = %v, want %v", gotFields, tt.wantFields)
				return
			}
			for i, f := range tt.wantFields {
				if gotFields[i] != f {
					t.Errorf("DiffStruct() fields[%d] = %q, want %q", i, gotFields[i], f)
				}
				if _, ok := got[f]; !ok {
					t.Errorf("DiffStruct() result missing key %q", f)
				}
			}
			if tt.wantValues != nil {
				for k, v := range tt.wantValues {
					gotVal, ok := got[k]
					if !ok {
						t.Errorf("DiffStruct() missing key %q in result map", k)
						continue
					}
					if !reflect.DeepEqual(gotVal, v) {
						t.Errorf("DiffStruct() got[%q] = %#v, want %#v", k, gotVal, v)
					}
				}
			}
		})
	}
}

func TestAssign(t *testing.T) {
	type Inner struct {
		X int
		Y string
	}
	type Outer struct {
		Name  string
		Age   int
		Inner Inner
		Ptr   *string
	}

	tests := []struct {
		name    string
		dst     any
		src     any
		want    Outer
		wantErr bool
	}{
		{
			name:    "空 src dst 不变",
			dst:     &Outer{Name: "keep"},
			src:     Outer{},
			want:    Outer{Name: "keep"},
			wantErr: false,
		},
		{
			name:    "部分字段赋值",
			dst:     &Outer{Name: "old"},
			src:     Outer{Name: "new", Age: 10},
			want:    Outer{Name: "new", Age: 10},
			wantErr: false,
		},
		{
			name: "指针字段赋值",
			dst:  &Outer{},
			src: func() Outer {
				s := "hello"
				return Outer{Ptr: &s}
			}(),
			want: func() Outer {
				s := "hello"
				return Outer{Ptr: &s}
			}(),
			wantErr: false,
		},
		{
			name:    "嵌套 struct 赋值",
			dst:     &Outer{},
			src:     Outer{Inner: Inner{X: 42, Y: "nested"}},
			want:    Outer{Inner: Inner{X: 42, Y: "nested"}},
			wantErr: false,
		},
		{
			name:    "非指针 dst",
			dst:     Outer{},
			src:     Outer{Name: "new"},
			wantErr: true,
		},
		{
			name:    "不同类型",
			dst:     &Outer{},
			src:     Inner{X: 1},
			wantErr: true,
		},
		{
			name:    "nil dst",
			dst:     (*Outer)(nil),
			src:     Outer{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Assign(tt.dst, tt.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("Assign() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			got := tt.dst.(*Outer)
			if got.Name != tt.want.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.want.Name)
			}
			if got.Age != tt.want.Age {
				t.Errorf("Age = %d, want %d", got.Age, tt.want.Age)
			}
			if got.Inner.X != tt.want.Inner.X {
				t.Errorf("Inner.X = %d, want %d", got.Inner.X, tt.want.Inner.X)
			}
			if got.Inner.Y != tt.want.Inner.Y {
				t.Errorf("Inner.Y = %q, want %q", got.Inner.Y, tt.want.Inner.Y)
			}
			if (got.Ptr == nil) != (tt.want.Ptr == nil) {
				t.Errorf("Ptr nil mismatch: got %v, want %v", got.Ptr == nil, tt.want.Ptr == nil)
			}
			if got.Ptr != nil && tt.want.Ptr != nil && *got.Ptr != *tt.want.Ptr {
				t.Errorf("Ptr = %q, want %q", *got.Ptr, *tt.want.Ptr)
			}
		})
	}
}
