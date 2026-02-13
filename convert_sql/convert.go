package convert_sql

import (
	"fmt"
	"regexp"
	"strings"
)

// ConvertView1 转换视图中的convert函数，例如convert(nvarchar(20),name)为name::varchar(20)
func ConvertView1(input string) string {
	// 改进的正则表达式：支持带逗号和小数点的类型参数
	re := regexp.MustCompile(`convert\s*\(\s*(\w+)(\(\s*[\d,\s.]+\s*\))?\s*,\s*([\w.'']+|'[^']*'|\([^)]+\))\s*\)`)

	// 定义类型映射（SQL Server -> PostgreSQL）
	typeMappings := map[string]string{
		"nvarchar":  "varchar",          // 移除n前缀
		"nchar":     "char",             // 移除n前缀
		"double":    "double precision", // 转换double为double precision
		"datetime":  "timestamp",        // 转换datetime为timestamp
		"datetime2": "timestamp",        // 转换datetime为timestamp
		"money":     "numeric",          // 转换money为numeric
		// 添加更多类型映射...
	}

	// 允许转换的数据类型集合
	allowedTypes := map[string]bool{
		"nvarchar":  true,
		"varchar":   true,
		"nchar":     true,
		"char":      true,
		"decimal":   true,
		"numeric":   true,
		"float":     true,
		"double":    true,
		"real":      true,
		"int":       true,
		"bigint":    true,
		"smallint":  true,
		"tinyint":   true,
		"datetime":  true,
		"datetime2": true,
		"date":      true,
		"time":      true,
		"money":     true,
		// 添加更多支持的类型...
	}

	result := re.ReplaceAllStringFunc(input, func(match string) string {
		submatches := re.FindStringSubmatch(match)
		if len(submatches) < 4 {
			return match
		}

		origType := strings.TrimSpace(submatches[1])
		typeLength := strings.TrimSpace(submatches[2])
		expr := strings.TrimSpace(submatches[3])

		// 检查是否支持此数据类型
		lowerType := strings.ToLower(origType)
		if !allowedTypes[lowerType] {
			return match // 不支持的类型保持原样
		}

		// 应用类型映射
		targetType := origType
		if mappedType, ok := typeMappings[lowerType]; ok {
			targetType = mappedType
		}

		// 添加类型参数（如(20,2)）
		if typeLength != "" {
			targetType += typeLength
		}

		// 特殊处理空字符串
		if expr == "''" {
			return "''::" + targetType
		}

		return fmt.Sprintf("%s::%s", expr, targetType)
	})

	return result
}

// ConvertView2 转换视图中的isnumeric函数为pgsql的正则表达式
func ConvertView2(input string) string {
	// 编译正则表达式，匹配 isnumeric(...)=1 并提取字段名
	// 模式解释:
	//   - `isnumeric\(` 匹配 "isnumeric("
	//   - `([^)]+)` 捕获组，匹配除 ")" 外的所有字符（即字段名）
	//   - `\)\s*=\s*1` 匹配 ")=1"，允许等号前后有空格
	re := regexp.MustCompile(`isnumeric\(([^)]+)\)\s*=\s*1`)

	// 替换字符串: $1 引用捕获的字段名
	replacement := "$1 ~ '^[-+]?[0-9]*\\.?[0-9]+([eE][-+]?[0-9]+)?$'"

	// 执行替换
	return re.ReplaceAllString(input, replacement)
}

// ConvertView3 批量去除视图中的top   100   percent
func ConvertView3(input string) string {
	// 正则解释：匹配"top"、"100"、"percent"三部分，中间允许任意空白符（包括0或多个空格）
	re := regexp.MustCompile(`(?i)\btop\s*100\s*percent\b`)
	result := re.ReplaceAllString(input, "")

	// 执行替换
	return result
}

// ConvertView4 批量将len函数转为length函数
func ConvertView4(input string) string {
	// 正则表达式解释：
	// (?i)        - 忽略大小写
	// \blen\b     - 匹配独立的"len"单词（避免匹配"length"等）
	// \s*         - 匹配0个或多个空白字符
	// \(          - 匹配左括号
	// (           - 开始捕获组
	//   [^)]*     - 匹配除右括号外的任意字符（0次或多次）
	//   |         - 或
	//   \(        - 匹配左括号（用于嵌套括号）
	//   [^)]*     - 匹配除右括号外的任意字符
	//   \)        - 匹配右括号
	//   [^)]*     - 匹配除右括号外的任意字符
	// )*          - 以上组合0次或多次
	// \)          - 匹配右括号
	re := regexp.MustCompile(`(?i)\blen\b\s*\(((?:[^)]*|\([^)]*\)[^)]*)*)\)`)

	// 替换为"length(捕获组内容)"
	return re.ReplaceAllString(input, "length($1)")
}

// ConvertView5 isnull(任意字符串,”)!=” 替换为 任意字符串 IS NOT NULL
func ConvertView5(input string) string {
	// 正则表达式解释：
	// (?i)       - 忽略大小写
	// \bisnull\b - 匹配独立单词"isnull"
	// \s*\(\s*  - 匹配左括号（前后可能有空格）
	// (          - 开始捕获组1（任意字符串）
	//   [^,]+    - 匹配除逗号外的任意字符（至少一个）
	// )          - 结束捕获组1
	// \s*,\s*''\s*\)\s*!=\s*'' - 匹配固定模式
	pattern := `(?i)\bisnull\b\s*\(\s*([^,]+)\s*,\s*''\s*\)\s*!=\s*''`
	re := regexp.MustCompile(pattern)

	// 替换为"$1 IS NOT NULL"
	return re.ReplaceAllString(input, "$1 IS NOT NULL")
}

// ConvertView6 正则表达式将 SQL 中的 convert(varchar(任意长度), 任意字符串, 120) 格式转换为 TO_CHAR(任意字符串, 'YYYY-MM-DD')
func ConvertView6(input string) string {
	// 正则表达式解释：
	// (?i) - 忽略大小写
	// \bconvert\b - 匹配独立的convert函数
	// \s*\(\s* - 匹配左括号（允许空格）
	// varchar\s*\(\s*\d+\s*\) - 匹配varchar(数字)
	// \s*,\s* - 逗号分隔符
	// ([^,]+) - 捕获组1：任意字符串（不含逗号）
	// \s*,\s*120\s*\) - 匹配",120)"
	pattern := `(?i)convert\s*\(\s*varchar\s*(?:\(\d+\))?\s*,\s*([^,]+?)\s*,\s*120\s*\)`

	re := regexp.MustCompile(pattern)

	// 替换为 TO_CHAR(捕获组1, 'YYYY-MM-DD')
	return re.ReplaceAllString(input, "TO_CHAR($1, 'YYYY-MM-DD')")
}
