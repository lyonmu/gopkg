package pinyin

import "testing"

func TestGetPinyinInitial(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正常输入
		{"纯中文-张三", "张三", "ZS"},
		{"纯中文-李四", "李四", "LS"},
		{"中文+字母", "张a三", "ZAS"},
		{"中文+数字", "张3三", "Z3S"},
		// 边界情况
		{"空字符串", "", ""},
		{"纯空格", "   ", ""},
		// Emoji
		{"纯emoji", "🎉🔥", ""},
		{"中文+emoji", "🎉张🔥三", "ZS"},
		// 混合
		{"空格+字母+emoji", " 张 a 三 🎉", "ZAS"},
		{"标点符号", "张,三!", "ZS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPinyinInitial(tt.input)
			if got != tt.expected {
				t.Errorf("GetPinyinInitial(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetPinyinSpelling(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正常输入
		{"纯中文-张三", "张三", "ZHANGSAN"},
		{"纯中文-李四", "李四", "LISI"},
		{"中文+字母", "张a", "ZHANGA"},
		// 边界情况
		{"空字符串", "", ""},
		{"纯空格", "  ", ""},
		// Emoji
		{"emoji+中文", "🔥张三", "ZHANGSAN"},
		{"emoji+字母", "🎉a", "A"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPinyinSpelling(tt.input)
			if got != tt.expected {
				t.Errorf("GetPinyinSpelling(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetPinyinBigHump(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正常输入
		{"纯中文-张三", "张三", "ZhangSan"},
		{"纯中文-李四", "李四", "LiSi"},
		{"中文+字母", "张a", "ZhangA"},
		// 边界情况
		{"空字符串", "", ""},
		{"纯空格", "  ", ""},
		// Emoji
		{"emoji+中文", "🎉张", "Zhang"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPinyinBigHump(tt.input)
			if got != tt.expected {
				t.Errorf("GetPinyinBigHump(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetPinyinSmallHump(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// 正常输入
		{"纯中文-张三", "张三", "zhangSan"},
		{"纯中文-你好世界", "你好世界", "niHaoShiJie"},
		{"中文+字母", "张a三", "zhangASan"},
		// 边界情况
		{"空字符串", "", ""},
		{"纯空格", "  ", ""},
		// Emoji
		{"emoji+中文", "🎉张三", "zhangSan"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPinyinSmallHump(tt.input)
			if got != tt.expected {
				t.Errorf("GetPinyinSmallHump(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsAllChinese(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// 正常输入
		{"纯中文", "张三", true},
		{"单字", "张", true},
		// 含数字
		{"中文+数字", "李四123", false},
		{"纯数字", "123", false},
		// 含字母
		{"纯字母", "abc", false},
		{"中文+字母", "张a", false},
		// 边界情况
		{"空字符串", "", false},
		{"纯空格", "   ", false},
		// 空格（被移除后判断）
		{"中文+空格", "张 三", true},
		// Emoji（被移除后判断）
		{"中文+emoji", "🎉张三", true},
		{"纯emoji", "🎉🔥", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsAllChinese(tt.input)
			if got != tt.expected {
				t.Errorf("IsAllChinese(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
