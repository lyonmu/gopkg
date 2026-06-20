package id

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type stubIDNode struct {
	id  int64
	err error
}

func (s stubIDNode) NextID() (int64, error) {
	return s.id, s.err
}

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

	id, err := gen.GenID()
	if err != nil {
		t.Fatalf("GenID() error = %v", err)
	}
	if id <= 0 {
		t.Errorf("GenID() = %d, want positive integer", id)
	}
}

func TestSonySnowFlake_GenID_PropagatesNodeResult(t *testing.T) {
	nodeErr := errors.New("next id error")
	tests := []struct {
		name    string
		node    idNode
		wantID  int64
		wantErr error
	}{
		{
			name:   "成功-返回节点ID",
			node:   stubIDNode{id: 42},
			wantID: 42,
		},
		{
			name:    "错误-传播节点错误",
			node:    stubIDNode{err: nodeErr},
			wantErr: nodeErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := &SonySnowFlake{node: tt.node}

			gotID, err := gen.GenID()

			if gotID != tt.wantID {
				t.Errorf("GenID() ID = %d, want %d", gotID, tt.wantID)
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GenID() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSonySnowFlake_GenID_ConcurrentUniqueness(t *testing.T) {
	gen, err := NewSonySnowFlake(func() (int, error) { return 1, nil })
	if err != nil {
		t.Fatalf("NewSonySnowFlake() error = %v", err)
	}

	const count = 1000
	ids := make(chan int64, count)
	errs := make(chan error, count)

	var wg sync.WaitGroup
	wg.Add(count)
	for range count {
		go func() {
			defer wg.Done()

			id, err := gen.GenID()
			if err != nil {
				errs <- err
				return
			}
			ids <- id
		}()
	}
	wg.Wait()
	close(ids)
	close(errs)

	for err := range errs {
		t.Errorf("GenID() error = %v", err)
	}

	seen := make(map[int64]struct{}, count)
	for id := range ids {
		if _, exists := seen[id]; exists {
			t.Errorf("GenID() generated duplicate ID: %d", id)
		}
		seen[id] = struct{}{}
	}
	if len(seen) != count {
		t.Errorf("GenID() generated %d unique IDs, want %d", len(seen), count)
	}
}

func TestSonySnowFlake_GenID_Monotonic(t *testing.T) {
	gen, err := NewSonySnowFlake(func() (int, error) { return 1, nil })
	if err != nil {
		t.Fatalf("NewSonySnowFlake() error = %v", err)
	}

	var prev int64
	for range 100 {
		id, err := gen.GenID()
		if err != nil {
			t.Fatalf("GenID() error = %v", err)
		}
		if id < prev {
			t.Errorf("GenID() = %d, want >= %d (monotonically increasing)", id, prev)
			return
		}
		prev = id
	}
}

func TestSonySnowFlake_SequentialInstancesUseFixedTimeBase(t *testing.T) {
	machineID := func() (int, error) { return 1, nil }

	first, err := NewSonySnowFlake(machineID)
	if err != nil {
		t.Fatalf("NewSonySnowFlake() first error = %v", err)
	}
	firstID, err := first.GenID()
	if err != nil {
		t.Fatalf("first GenID() error = %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	second, err := NewSonySnowFlake(machineID)
	if err != nil {
		t.Fatalf("NewSonySnowFlake() second error = %v", err)
	}
	secondID, err := second.GenID()
	if err != nil {
		t.Fatalf("second GenID() error = %v", err)
	}

	if secondID <= firstID {
		t.Errorf("second GenID() = %d, want > first ID %d after more than one 10ms time unit", secondID, firstID)
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

	id1, err := gen1.GenID()
	if err != nil {
		t.Fatalf("GenID() machine 1 error = %v", err)
	}
	id2, err := gen2.GenID()
	if err != nil {
		t.Fatalf("GenID() machine 2 error = %v", err)
	}

	// 不同机器 ID 生成的 ID 应该不同（在同一时间窗口内）
	if id1 == id2 {
		t.Errorf("Different machine IDs produced same ID: %d", id1)
	}
}

func TestIDGenerator_Interface(t *testing.T) {
	// 验证 SonySnowFlake 实现了 IDGenerator 接口
	var _ IDGenerator = (*SonySnowFlake)(nil)
}
