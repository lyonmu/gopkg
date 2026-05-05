package pinyin

import (
	"regexp"
	"strings"

	go_pinyin "github.com/mozillazg/go-pinyin"
)

// cleanInput 清理输入字符串：移除空格和 emoji 字符
// 返回仅保留中文、字母、数字和标点的有效字符串
func cleanInput(s string) string {
	// 移除所有空格
	s = strings.ReplaceAll(s, " ", "")
	// 移除 emoji（Unicode 范围：杂项符号、表情符号、补充符号等）
	reg := regexp.MustCompile(`[\x{1F600}-\x{1F64F}\x{1F300}-\x{1F5FF}\x{1F680}-\x{1F6FF}\x{1F1E0}-\x{1F1FF}\x{2600}-\x{26FF}\x{2700}-\x{27BF}\x{FE00}-\x{FE0F}\x{1F900}-\x{1F9FF}\x{1FA00}-\x{1FA6F}\x{1FA70}-\x{1FAFF}\x{200D}\x{20E3}\x{00A9}\x{00AE}]`)
	return reg.ReplaceAllString(s, "")
}

// GetPinyinInitial 返回中文字符串的拼音首字母（大写）
// 非中文字符中，字母和数字原样保留，标点符号和 emoji 被过滤
// 示例: "张三" -> "ZS", "张a3" -> "ZA3"
func GetPinyinInitial(chinese string) string {
	chinese = cleanInput(chinese)
	args := go_pinyin.NewArgs()
	args.Style = go_pinyin.FirstLetter // 仅获取首字母
	args.Fallback = func(r rune, a go_pinyin.Args) []string {
		// 字母和数字原样保留
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return []string{string(r)}
		}
		return []string{} // 其他非中文字符跳过
	}

	// 转换为拼音
	pinyinSlice := go_pinyin.Pinyin(chinese, args)
	var result strings.Builder
	for _, p := range pinyinSlice {
		if len(p) > 0 && len(p[0]) > 0 {
			result.WriteString(strings.ToUpper(p[0]))
		}
	}
	return result.String()
}

// GetPinyinSpelling 返回中文字符串的完整拼音（全大写，无分隔符）
// 字母原样保留并转为大写，非中文字符中只有字母被保留
// 示例: "张三" -> "ZHANGSAN"
func GetPinyinSpelling(chinese string) string {
	chinese = cleanInput(chinese)
	args := go_pinyin.NewArgs()
	args.Style = go_pinyin.Normal // 获取完整拼音
	args.Fallback = func(r rune, a go_pinyin.Args) []string {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			return []string{string(r)}
		}
		return []string{}
	}

	pinyinSlice := go_pinyin.Pinyin(chinese, args)
	var result strings.Builder
	for _, p := range pinyinSlice {
		if len(p) > 0 && len(p[0]) > 0 {
			result.WriteString(strings.ToUpper(p[0]))
		}
	}
	return result.String()
}

// GetPinyinBigHump 返回中文字符串的拼音（大驼峰格式：每个拼音首字母大写）
// 字母原样保留（首字母大写处理），非中文字符中只有字母被保留
// 示例: "张三" -> "ZhangSan"
func GetPinyinBigHump(chinese string) string {
	chinese = cleanInput(chinese)
	args := go_pinyin.NewArgs()
	args.Style = go_pinyin.Normal
	args.Fallback = func(r rune, a go_pinyin.Args) []string {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			return []string{string(r)}
		}
		return []string{}
	}

	pinyinSlice := go_pinyin.Pinyin(chinese, args)
	var result strings.Builder
	for _, p := range pinyinSlice {
		if len(p) > 0 && len(p[0]) > 0 {
			result.WriteString(strings.ToUpper(string(p[0][0])) + p[0][1:])
		}
	}
	return result.String()
}

// GetPinyinSmallHump 返回中文字符串的拼音（小驼峰格式：首字母小写，后续每个拼音首字母大写）
// 字母原样保留，非中文字符中只有字母被保留
// 示例: "张三" -> "zhangSan", "你好世界" -> "niHaoShiJie"
func GetPinyinSmallHump(chinese string) string {
	chinese = cleanInput(chinese)
	args := go_pinyin.NewArgs()
	args.Style = go_pinyin.Normal
	args.Fallback = func(r rune, a go_pinyin.Args) []string {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			return []string{string(r)}
		}
		return []string{}
	}

	pinyinSlice := go_pinyin.Pinyin(chinese, args)
	var result strings.Builder
	isFirst := true
	for _, p := range pinyinSlice {
		if len(p) > 0 && len(p[0]) > 0 {
			if isFirst {
				// 第一个拼音全小写
				result.WriteString(strings.ToLower(p[0]))
				isFirst = false
			} else {
				// 后续拼音首字母大写
				result.WriteString(strings.ToUpper(string(p[0][0])) + p[0][1:])
			}
		}
	}
	return result.String()
}

// IsAllChinese 判断字符串是否全部由中文字符组成
// 处理前会移除空格和 emoji，仅判断剩余的字符是否全部为中文
// 空字符串返回 false
// 示例: "张三" -> true, "abc" -> false, "张 三" -> true
func IsAllChinese(s string) bool {
	s = cleanInput(s)
	if s == "" {
		return false
	}
	reg := regexp.MustCompile(`^\p{Han}+$`)
	return reg.MatchString(s)
}
