package mkpgo

import (
	"encoding/json"
	"errors"
	"strings"
)

// Parser registry and shared parser errors.
// 解析器注册表与通用错误定义。
var (
	// RawDirectiveOutputParsers maps directive verb to its output parser.
	// RawDirectiveOutputParsers 将指令名映射到对应输出解析器。
	RawDirectiveOutputParsers = make(map[string]RawDirectiveOutputParser)
	// Parser related common errors.
	// 解析器相关通用错误。
	ErrDirectiveParserMissing      = errors.New("directive parser missing")
	ErrRawDirectiveParseFailed     = errors.New("raw directive parse failed")
	ErrRawDirecitveExecutionFailed = errors.New("raw directive execution failed")
)

// Built-in parser assertions and singleton instances.
// 内置解析器的接口断言与单例实例。
var (
	// Interface implementation assertions.
	// 接口实现断言。
	_ RawDirectiveOutputParser = (*RawDirective_sn)(nil)
	_ RawDirectiveOutputParser = (*RawDirective_list_dir)(nil)

	// Built-in singleton parser instances.
	// 内置解析器单例实例。
	InstantRawDirective_sn          = NewRawDirective_sn()
	InstantRawDirective_list_dir    = NewRawDirective_list_dir()
	InstantRawDirective_clean_dir   = NewRawDirective_clean_dir()
	InstantRawDirective_alive       = NewRawDirective_alive()
	InstantRawDirective_atime       = NewRawDirective_atime()
	InstantRawDirective_aversion    = NewRawDirective_aversion()
	InstantRawDirective_delete_file = NewRawDirective_delete_file()
	InstantRawDirective_ainsp       = NewRawDirective_ainsp()
	InstantRawDirective_astop       = NewRawDirective_astop()
	InstantRawDirective_acancel     = NewRawDirective_acancel()
	InstantRawDirective_alog        = NewRawDirective_alog()
)

// InitParsers registers built-in directive output parsers.
// InitParsers 注册内置指令输出解析器。
func InitParsers() {
	parsers := []RawDirectiveOutputParser{
		InstantRawDirective_sn,
		InstantRawDirective_list_dir,
		InstantRawDirective_delete_file,
		InstantRawDirective_clean_dir,
		InstantRawDirective_aversion,
		InstantRawDirective_atime,
		InstantRawDirective_alive,
		InstantRawDirective_ainsp,
		InstantRawDirective_astop,
		InstantRawDirective_acancel,
		InstantRawDirective_alog,
	}

	for _, parser := range parsers {
		RegisterRawDirectiveOutputParser(parser)
	}
}

// GetRawDirective extracts directive verb from full command.
// GetRawDirective 从完整命令提取原始指令名。
func GetRawDirective(directive string) string {
	return strings.Split(directive, " ")[0]
}

// GetRawDirectiveOutputParser returns parser bound to directive verb.
// GetRawDirectiveOutputParser 返回对应指令解析器。
func GetRawDirectiveOutputParser(directive string) RawDirectiveOutputParser {
	rawDirective := GetRawDirective(directive)

	return RawDirectiveOutputParsers[rawDirective]
}

// RegisterRawDirectiveOutputParser registers one parser by name.
// RegisterRawDirectiveOutputParser 按名称注册一个解析器。
func RegisterRawDirectiveOutputParser(parser RawDirectiveOutputParser) {
	RawDirectiveOutputParsers[parser.String()] = parser
}

// RawDirectiveOutputParser defines parse contract for raw CLI output.
// RawDirectiveOutputParser 定义原始 CLI 输出解析契约。
type RawDirectiveOutputParser interface {
	String() string
	Parse(string, string) (string, error) // 第一个是cli 完整指令（含参数), 第二个是输出
	UnmarshalTo(string, interface{}) error
	IsJSONOutput() bool
}

// RawDirective provides shared parser behavior.
// RawDirective 提供解析器公共行为。
type RawDirective struct {
	JSONOutput bool
}

// SanitizeR removes carriage returns from raw output.
// SanitizeR 移除输出中的回车符。
func (*RawDirective) SanitizeR(data string) string {
	return strings.ReplaceAll(data, "\r", "")
}

// UnmarshalTo decodes JSON text into target object.
// UnmarshalTo 将 JSON 文本反序列化到目标对象。
func (r *RawDirective) UnmarshalTo(data string, ref interface{}) error {
	return json.Unmarshal([]byte(data), ref)
}

// PreFlight performs common sanitation and execution-failure check.
// PreFlight 执行公共清洗与执行失败检测。
func (r *RawDirective) PreFlight(data string) (string, error) {
	data = r.SanitizeR(data)

	if r.IsExecutionFailed(data) {
		return "", ErrRawDirecitveExecutionFailed
	}

	return data, nil
}

// IsExecutionFailed checks firmware error pattern in output.
// IsExecutionFailed 判断输出中是否包含执行失败特征。
func (r *RawDirective) IsExecutionFailed(data string) bool {
	return strings.Contains(strings.ToLower(data), "command returned non-zero error code")
}

// IsJSONOutput reports whether parser expects JSON payload.
// IsJSONOutput 返回是否期望 JSON 输出。
func (r *RawDirective) IsJSONOutput() bool {
	return r.JSONOutput
}

/* Built-in directive parsers / 内置指令解析器 */

// RawDirective_astop parses astop output.
// RawDirective_astop 解析 astop 指令输出。
type RawDirective_astop struct {
	*RawDirective
	Name string
}

// NewRawDirective_astop creates astop parser.
// NewRawDirective_astop 创建 astop 解析器。
func NewRawDirective_astop() *RawDirective_astop {
	return &RawDirective_astop{Name: "astop", RawDirective: &RawDirective{}}
}

// String returns directive name.
// String 返回对应指令名。
func (r *RawDirective_astop) String() string {
	return r.Name
}

// Parse parses astop raw output.
// Parse 解析 astop 原始输出。
func (r *RawDirective_astop) Parse(_, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	if len(data) > 0 {
		return data, nil
	}

	return "", ErrRawDirectiveParseFailed
}

// RawDirective_acancel parses acancel output.
// RawDirective_acancel 解析 acancel 指令输出。
type RawDirective_acancel struct {
	*RawDirective
	Name string
}

// NewRawDirective_acancel creates acancel parser.
// NewRawDirective_acancel 创建 acancel 解析器。
func NewRawDirective_acancel() *RawDirective_acancel {
	return &RawDirective_acancel{Name: "acancel", RawDirective: &RawDirective{JSONOutput: true}}
}

// String returns directive name.
// String 返回对应指令名。
func (r *RawDirective_acancel) String() string {
	return r.Name
}

// Parse parses acancel raw output.
// Parse 解析 acancel 原始输出。
func (r *RawDirective_acancel) Parse(cli, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, cli)

	if len(data) > 0 {
		return data, nil
	}

	return "", ErrRawDirectiveParseFailed
}

// RawDirective_sn parses sn output.
// RawDirective_sn 解析 sn 指令输出。
type RawDirective_sn struct {
	*RawDirective
	Name string
}

// NewRawDirective_sn creates sn parser.
// NewRawDirective_sn 创建 sn 解析器。
func NewRawDirective_sn() *RawDirective_sn {
	return &RawDirective_sn{Name: "sn", RawDirective: &RawDirective{JSONOutput: true}}
}

// String returns directive name.
// String 返回对应指令名。
func (r *RawDirective_sn) String() string {
	return r.Name
}

// Parse parses sn raw output.
// Parse 解析 sn 原始输出。
func (r *RawDirective_sn) Parse(cli, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, cli)

	if len(data) > 0 {
		return data, nil
	}

	return "", ErrRawDirectiveParseFailed
}

// RawDirective_list_dir parses list_dir output.
// RawDirective_list_dir 解析 list_dir 指令输出。
type RawDirective_list_dir struct {
	*RawDirective
	Name string
}

// NewRawDirective_list_dir creates list_dir parser.
// NewRawDirective_list_dir 创建 list_dir 解析器。
func NewRawDirective_list_dir() *RawDirective_list_dir {
	return &RawDirective_list_dir{Name: "list_dir", RawDirective: &RawDirective{JSONOutput: true}}
}

// String returns directive name.
// String 返回对应指令名。
func (r *RawDirective_list_dir) String() string {
	return r.Name
}

// Parse parses list_dir raw output.
// Parse 解析 list_dir 原始输出。
func (r *RawDirective_list_dir) Parse(cli, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, cli)

	if len(data) > 0 {
		return data, nil
	}

	return "", nil
}

// RawDirective_clean_dir parses clean_dir output.
// RawDirective_clean_dir 解析 clean_dir 指令输出。
type RawDirective_clean_dir struct {
	*RawDirective
	Name string
}

// NewRawDirective_clean_dir creates clean_dir parser.
// NewRawDirective_clean_dir 创建 clean_dir 解析器。
func NewRawDirective_clean_dir() *RawDirective_clean_dir {
	return &RawDirective_clean_dir{Name: "clean_dir", RawDirective: &RawDirective{JSONOutput: false}}
}

// String returns directive name.
// String 返回对应指令名。
func (r *RawDirective_clean_dir) String() string {
	return r.Name
}

// Parse parses clean_dir raw output.
// Parse 解析 clean_dir 原始输出。
func (r *RawDirective_clean_dir) Parse(cli, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, cli)

	// E (2055915) sdcard: Failed to open directory: No such file or directory
	if strings.Contains(data, "Failed to") {
		return "", errors.New("failed to clean directory")
	}

	return "", nil
}

// RawDirective_delete_file parses delete_file output.
// RawDirective_delete_file 解析 delete_file 指令输出。
type RawDirective_delete_file struct {
	*RawDirective
	Name string
}

// NewRawDirective_delete_file creates delete_file parser.
// NewRawDirective_delete_file 创建 delete_file 解析器。
func NewRawDirective_delete_file() *RawDirective_delete_file {
	return &RawDirective_delete_file{Name: "delete_file", RawDirective: &RawDirective{JSONOutput: false}}
}

// String returns directive name.
// String 返回对应指令名。
func (r *RawDirective_delete_file) String() string {
	return r.Name
}

// Parse parses delete_file raw output.
// Parse 解析 delete_file 原始输出。
func (r *RawDirective_delete_file) Parse(cli, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, cli)

	// E (2055915) sdcard: Failed to remove /eMMC/applog/mkpdemo/1111.log: No such file or directory
	if strings.Contains(data, "Failed to remove") {
		return "", errors.New("failed to remove file")
	}

	return "", nil
}

// RawDirective_alive parses alive output.
// RawDirective_alive 解析 alive 指令输出。
type RawDirective_alive struct {
	*RawDirective
	Name string
}

// NewRawDirective_alive creates alive parser.
// NewRawDirective_alive 创建 alive 解析器。
func NewRawDirective_alive() *RawDirective_alive {
	return &RawDirective_alive{Name: "alive", RawDirective: &RawDirective{JSONOutput: true}}
}

// String returns directive name.
// String 返回对应指令名。
func (r *RawDirective_alive) String() string {
	return r.Name
}

// Parse parses alive raw output.
// Parse 解析 alive 原始输出。
func (r *RawDirective_alive) Parse(cli, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, cli)

	if len(data) > 0 {
		return data, nil
	}

	return "", nil
}

// RawDirective_atime parses atime output.
// RawDirective_atime 解析 atime 指令输出。
type RawDirective_atime struct {
	*RawDirective
	Name string
}

// NewRawDirective_atime creates atime parser.
// NewRawDirective_atime 创建 atime 解析器。
func NewRawDirective_atime() *RawDirective_atime {
	return &RawDirective_atime{Name: "atime", RawDirective: &RawDirective{JSONOutput: true}}
}

// String returns directive name.
// String 返回对应指令名。
func (r *RawDirective_atime) String() string {
	return r.Name
}

// Parse parses atime raw output.
// Parse 解析 atime 原始输出。
func (r *RawDirective_atime) Parse(cli, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, cli)

	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") && strings.Contains(line, `"seconds":`) {
			return line, nil
		}
	}

	return "", nil
}

// RawDirective_aversion parses aversion output.
// RawDirective_aversion 解析 aversion 指令输出。
type RawDirective_aversion struct {
	*RawDirective
	Name string
}

// NewRawDirective_aversion creates aversion parser.
// NewRawDirective_aversion 创建 aversion 解析器。
func NewRawDirective_aversion() *RawDirective_aversion {
	return &RawDirective_aversion{Name: "aversion", RawDirective: &RawDirective{JSONOutput: true}}
}

// String returns directive name.
// String 返回对应指令名。
func (r *RawDirective_aversion) String() string {
	return r.Name
}

// Parse parses aversion raw output.
// Parse 解析 aversion 原始输出。
func (r *RawDirective_aversion) Parse(cli, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, cli)

	if len(data) > 0 {
		return data, nil
	}

	return "", nil
}

// RawDirective_ainsp parses ainsp output (log metadata).
// RawDirective_ainsp 解析 ainsp 指令输出（日志元信息）。
type RawDirective_ainsp struct {
	*RawDirective
	Name string
}

// NewRawDirective_ainsp creates ainsp parser.
// NewRawDirective_ainsp 创建 ainsp 解析器。
func NewRawDirective_ainsp() *RawDirective_ainsp {
	return &RawDirective_ainsp{Name: "ainsp", RawDirective: &RawDirective{JSONOutput: true}}
}

// String returns directive name.
// String 返回对应指令名。
func (r *RawDirective_ainsp) String() string {
	return r.Name
}

// Parse parses ainsp raw output.
// Parse 解析 ainsp 原始输出。
func (r *RawDirective_ainsp) Parse(cli, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, cli)

	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") && strings.Contains(line, `"seconds":`) && strings.Contains(line, `"width":`) {
			return line, nil
		}
	}

	return "", nil
}

// RawDirective_alog parses alog output.
// RawDirective_alog 解析 alog 指令输出。
type RawDirective_alog struct {
	*RawDirective
	Name string
}

// NewRawDirective_alog creates alog parser.
// NewRawDirective_alog 创建 alog 解析器。
func NewRawDirective_alog() *RawDirective_alog {
	return &RawDirective_alog{Name: "alog", RawDirective: &RawDirective{JSONOutput: false}}
}

// String returns directive name.
// String 返回对应指令名。
func (r *RawDirective_alog) String() string {
	return r.Name
}

// Parse parses alog raw output.
// Parse 解析 alog 原始输出。
func (r *RawDirective_alog) Parse(cli, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	data = strings.TrimSpace(data)
	data = strings.TrimPrefix(data, cli)

	if len(data) > 0 {
		return data, nil
	}

	return "", nil
}
