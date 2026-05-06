package id

import (
	"errors"
	"testing"
)

func TestNewSonySnowFlake(t *testing.T) {
	tests := []struct {
		name      string
		machineId func() (int, error)
		wantErr   bool
	}{
		{
			name:      "正常-有效机器ID",
			machineId: func() (int, error) { return 1, nil },
			wantErr:   false,
		},
		{
			name:      "正常-大机器ID",
			machineId: func() (int, error) { return 65535, nil },
			wantErr:   false,
		},
		{
			name:      "错误-machineId返回错误",
			machineId: func() (int, error) { return 0, errors.New("machine id error") },
			wantErr:   true,
		},
		{
			name:      "错误-machineId返回0被CheckMachineID拒绝",
			machineId: func() (int, error) { return 0, nil },
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSonySnowFlake(tt.machineId)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSonySnowFlake() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got == nil {
				t.Errorf("NewSonySnowFlake() returned nil generator")
			}
		})
	}
}

func TestSonySnowFlake_GenID(t *testing.T) {
	gen, err := NewSonySnowFlake(func() (int, error) { return 1, nil })
	if err != nil {
		t.Fatalf("NewSonySnowFlake() error = %v", err)
	}

	id := gen.GenID()
	if id <= 0 {
		t.Errorf("GenID() = %d, want positive integer", id)
	}
}

func TestSonySnowFlake_GenID_Uniqueness(t *testing.T) {
	gen, err := NewSonySnowFlake(func() (int, error) { return 1, nil })
	if err != nil {
		t.Fatalf("NewSonySnowFlake() error = %v", err)
	}

	const count = 1000
	seen := make(map[int64]bool, count)
	for i := 0; i < count; i++ {
		id := gen.GenID()
		if seen[id] {
			t.Errorf("GenID() generated duplicate ID: %d at iteration %d", id, i)
			return
		}
		seen[id] = true
	}
}

func TestSonySnowFlake_GenID_Monotonic(t *testing.T) {
	gen, err := NewSonySnowFlake(func() (int, error) { return 1, nil })
	if err != nil {
		t.Fatalf("NewSonySnowFlake() error = %v", err)
	}

	var prev int64
	for i := 0; i < 100; i++ {
		id := gen.GenID()
		if id < prev {
			t.Errorf("GenID() = %d, want >= %d (monotonically increasing)", id, prev)
			return
		}
		prev = id
	}
}

func TestSonySnowFlake_DifferentMachineID(t *testing.T) {
	gen1, err := NewSonySnowFlake(func() (int, error) { return 1, nil })
	if err != nil {
		t.Fatalf("NewSonySnowFlake(1) error = %v", err)
	}
	gen2, err := NewSonySnowFlake(func() (int, error) { return 2, nil })
	if err != nil {
		t.Fatalf("NewSonySnowFlake(2) error = %v", err)
	}

	id1 := gen1.GenID()
	id2 := gen2.GenID()

	// 不同机器 ID 生成的 ID 应该不同（在同一时间窗口内）
	if id1 == id2 {
		t.Errorf("Different machine IDs produced same ID: %d", id1)
	}
}

func TestIDGenerator_Interface(t *testing.T) {
	// 验证 SonySnowFlake 实现了 IDGenerator 接口
	var _ IDGenerator = (*SonySnowFlake)(nil)
}
