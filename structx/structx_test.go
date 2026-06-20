package structx

import (
	"reflect"
	"testing"
	"time"
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
			name:    "嵌套空 struct",
			input:   struct{ Empty struct{} }{},
			want:    map[string]any{"Empty": map[string]any{}},
			wantErr: false,
		},
		{
			name:    "slice 字段",
			input:   struct{ S []int }{S: []int{1, 2, 3}},
			want:    map[string]any{"S": []any{1, 2, 3}},
			wantErr: false,
		},
		{
			name:    "非空 array 字段",
			input:   struct{ A [3]int }{A: [3]int{1, 2, 3}},
			want:    map[string]any{"A": []any{1, 2, 3}},
			wantErr: false,
		},
		{
			name:    "空 array 字段",
			input:   struct{ A [0]int }{},
			want:    map[string]any{"A": []any{}},
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

func TestStructToMapPointerCycles(t *testing.T) {
	type node struct {
		Name string
		Next *node
	}

	tests := []struct {
		name  string
		input func() *node
		want  map[string]any
	}{
		{
			name: "self cycle",
			input: func() *node {
				root := &node{Name: "root"}
				root.Next = root
				return root
			},
			want: map[string]any{
				"Name": "root",
				"Next": nil,
			},
		},
		{
			name: "mutual cycle",
			input: func() *node {
				first := &node{Name: "first"}
				second := &node{Name: "second"}
				first.Next = second
				second.Next = first
				return first
			},
			want: map[string]any{
				"Name": "first",
				"Next": map[string]any{
					"Name": "second",
					"Next": nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StructToMap(tt.input())
			if err != nil {
				t.Fatalf("StructToMap() error = %v", err)
			}
			if got == nil {
				t.Fatal("StructToMap() returned nil map")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("StructToMap() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestStructToMapDepthLimitBoundary(t *testing.T) {
	type node struct {
		Value int
		Next  *node
	}

	buildChain := func(depth int) *node {
		root := &node{}
		current := root
		for i := 1; i <= depth; i++ {
			current.Next = &node{Value: i}
			current = current.Next
		}
		return root
	}
	var wantChain func(index, depth int) map[string]any
	wantChain = func(index, depth int) map[string]any {
		if index > maxDepth {
			return map[string]any{
				"Value": nil,
				"Next":  nil,
			}
		}
		var next any
		if index < depth {
			next = wantChain(index+1, depth)
		}
		return map[string]any{
			"Value": index,
			"Next":  next,
		}
	}

	tests := []struct {
		name  string
		depth int
	}{
		{
			name:  "below max depth",
			depth: maxDepth - 1,
		},
		{
			name:  "at max depth",
			depth: maxDepth,
		},
		{
			name:  "above max depth",
			depth: maxDepth + 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := buildChain(tt.depth)
			first, err := StructToMap(root)
			if err != nil {
				t.Fatalf("first StructToMap() error = %v", err)
			}
			second, err := StructToMap(root)
			if err != nil {
				t.Fatalf("second StructToMap() error = %v", err)
			}

			want := wantChain(0, tt.depth)
			if !reflect.DeepEqual(first, want) {
				t.Fatalf("StructToMap() = %#v, want %#v", first, want)
			}
			if !reflect.DeepEqual(first, second) {
				t.Fatalf("StructToMap() results differ:\nfirst:  %#v\nsecond: %#v", first, second)
			}
		})
	}
}

func TestStructToMapSharedAndContainerPointers(t *testing.T) {
	type node struct {
		Name     string
		Next     *node
		Children []*node
		Lookup   map[string]*node
	}
	type siblings struct {
		Left  *node
		Right *node
	}

	shared := &node{Name: "shared"}
	selfSlice := &node{Name: "slice"}
	selfSlice.Children = []*node{selfSlice}
	selfMap := &node{Name: "map"}
	selfMap.Lookup = map[string]*node{"self": selfMap}

	tests := []struct {
		name  string
		input any
		want  map[string]any
	}{
		{
			name:  "sibling fields share acyclic pointer",
			input: siblings{Left: shared, Right: shared},
			want: map[string]any{
				"Left": map[string]any{
					"Name":     "shared",
					"Next":     nil,
					"Children": nil,
					"Lookup":   nil,
				},
				"Right": map[string]any{
					"Name":     "shared",
					"Next":     nil,
					"Children": nil,
					"Lookup":   nil,
				},
			},
		},
		{
			name:  "cycle through pointer slice",
			input: selfSlice,
			want: map[string]any{
				"Name": "slice",
				"Next": nil,
				"Children": []any{
					nil,
				},
				"Lookup": nil,
			},
		},
		{
			name:  "cycle through pointer map",
			input: selfMap,
			want: map[string]any{
				"Name":     "map",
				"Next":     nil,
				"Children": nil,
				"Lookup": map[string]any{
					"self": nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := StructToMap(tt.input)
			if err != nil {
				t.Fatalf("StructToMap() error = %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("StructToMap() = %#v, want %#v", got, tt.want)
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
		Name      string
		Age       int
		Inner     Inner
		Ptr       *string
		CreatedAt time.Time
	}

	str1 := "hello"
	str2 := "world"
	createdAt := time.Date(2026, time.June, 20, 12, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

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
			dst:        Outer{Name: "test", Age: 10, CreatedAt: createdAt},
			src:        Outer{Name: "test", Age: 10, CreatedAt: createdAt},
			wantFields: []string{},
			wantErr:    false,
		},
		{
			name:       "time.Time 字段不同",
			dst:        Outer{CreatedAt: createdAt},
			src:        Outer{CreatedAt: updatedAt},
			wantFields: []string{"CreatedAt"},
			wantValues: map[string]any{"CreatedAt": createdAt},
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

func TestDiffStructCyclicValues(t *testing.T) {
	type node struct {
		Name string
		Next *node
	}

	cycle := func(rootName, childName string) *node {
		root := &node{Name: rootName}
		child := &node{Name: childName}
		root.Next = child
		child.Next = root
		return root
	}

	tests := []struct {
		name       string
		dst        *node
		src        *node
		wantFields []string
		wantValues map[string]any
	}{
		{
			name:       "equal cycles",
			dst:        cycle("root", "child"),
			src:        cycle("root", "child"),
			wantFields: []string{},
			wantValues: map[string]any{},
		},
		{
			name:       "different cycles",
			dst:        cycle("root", "before"),
			src:        cycle("root", "after"),
			wantFields: []string{"Next"},
			wantValues: map[string]any{
				"Next": map[string]any{
					"Name": "before",
					"Next": nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotFields, err := DiffStruct(tt.dst, tt.src)
			if err != nil {
				t.Fatalf("DiffStruct() error = %v", err)
			}
			if !reflect.DeepEqual(gotFields, tt.wantFields) {
				t.Fatalf("DiffStruct() fields = %v, want %v", gotFields, tt.wantFields)
			}
			if !reflect.DeepEqual(got, tt.wantValues) {
				t.Fatalf("DiffStruct() = %#v, want %#v", got, tt.wantValues)
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
