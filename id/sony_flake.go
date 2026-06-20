package id

import sonyflakeV2 "github.com/sony/sonyflake/v2"

type IDGenerator interface {
	GenID() (int64, error)
}

type idNode interface {
	NextID() (int64, error)
}

type SonySnowFlake struct {
	node idNode
}

func NewSonySnowFlake(machineId func() (int, error)) (IDGenerator, error) {
	settings := sonyflakeV2.Settings{
		// 序列号所占的位数，越大能在同一时间窗口内生成的 ID 越多；0 表示使用默认 8 位，≥31 会报错
		BitsSequence: 0,
		// 机器 ID 所占位数，用于区分不同节点；0 表示默认 16 位，≥31 会报错
		BitsMachineID: 0,
		// 时间粒度，控制时间位的步长；0 表示默认 10ms，必须 ≥1ms
		TimeUnit: 0,
		// 起始时间留空，使用 Sonyflake 固定默认值 2025-01-01 00:00:00 +0000 UTC
		// 生成当前实例机器 ID 的函数；若返回错误，实例创建失败；留空则默认使用私有 IP 的低 16 位
		MachineID: machineId,
		// 校验机器 ID 唯一性的函数；返回 false 则实例不会创建；留空则不做校验
		CheckMachineID: func(id int) bool {
			return id != 0
		},
	}
	node, err := sonyflakeV2.New(settings)
	if err != nil {
		return nil, err
	}
	return &SonySnowFlake{node: node}, nil
}

func (s *SonySnowFlake) GenID() (int64, error) {
	return s.node.NextID()
}
