package mkpgo

import (
	"encoding/json"
	"errors"
	"strings"
)

var (
	RawDirectiveOutputParsers      = make(map[string]RawDirectiveOutputParser)
	ErrDirectiveParserMissing      = errors.New("directive parser missing")
	ErrRawDirectiveParseFailed     = errors.New("raw directive parse failed")
	ErrRawDirecitveExecutionFailed = errors.New("raw directive execution failed")
)

var (
	_ RawDirectiveOutputParser = (*RawDirective_sn)(nil)
	_ RawDirectiveOutputParser = (*RawDirective_list_dir)(nil)

	InstantRawDirective_sn          = NewRawDirective_sn()
	InstantRawDirective_list_dir    = NewRawDirective_list_dir()
	InstantRawDirective_clean_dir   = NewRawDirective_clean_dir()
	InstantRawDirective_alive       = NewRawDirective_alive()
	InstantRawDirective_atime       = NewRawDirective_atime()
	InstantRawDirective_aversion    = NewRawDirective_aversion()
	InstantRawDirective_delete_file = NewRawDirective_delete_file()
	InstantRawDirective_ainsp       = NewRawDirective_ainsp()
	InstantRawDirective_astop       = NewRawDirective_astop()
	InstantRawDirective_alog        = NewRawDirective_alog()
)

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
		InstantRawDirective_alog,
	}

	for _, parser := range parsers {
		RegisterRawDirectiveOutputParser(parser)
	}
}

func GetRawDirective(directive string) string {
	return strings.Split(directive, " ")[0]
}

func GetRawDirectiveOutputParser(directive string) RawDirectiveOutputParser {
	rawDirective := GetRawDirective(directive)

	return RawDirectiveOutputParsers[rawDirective]
}

func RegisterRawDirectiveOutputParser(parser RawDirectiveOutputParser) {
	RawDirectiveOutputParsers[parser.String()] = parser
}

type RawDirectiveOutputParser interface {
	String() string
	Parse(string, string) (string, error) // 第一个是cli 完整指令（含参数), 第二个是输出
	UnmarshalTo(string, interface{}) error
	IsJSONOutput() bool
}

type RawDirective struct {
	JSONOutput bool
}

func (*RawDirective) SanitizeR(data string) string {
	return strings.ReplaceAll(data, "\r", "")
}

func (r *RawDirective) UnmarshalTo(data string, ref interface{}) error {
	return json.Unmarshal([]byte(data), ref)
}

func (r *RawDirective) PreFlight(data string) (string, error) {
	data = r.SanitizeR(data)

	if r.IsExecutionFailed(data) {
		return "", ErrRawDirecitveExecutionFailed
	}

	return data, nil
}

func (r *RawDirective) IsExecutionFailed(data string) bool {
	return strings.Contains(strings.ToLower(data), "command returned non-zero error code")
}

func (r *RawDirective) IsJSONOutput() bool {
	return r.JSONOutput
}

/* 以下是指令解析列表 */

// astop 指令
type RawDirective_astop struct {
	*RawDirective
	Name string
}

func NewRawDirective_astop() *RawDirective_astop {
	return &RawDirective_astop{Name: "astop", RawDirective: &RawDirective{}}
}

func (r *RawDirective_astop) String() string {
	return r.Name
}

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

// sn 指令
type RawDirective_sn struct {
	*RawDirective
	Name string
}

func NewRawDirective_sn() *RawDirective_sn {
	return &RawDirective_sn{Name: "sn", RawDirective: &RawDirective{}}
}

func (r *RawDirective_sn) String() string {
	return r.Name
}

func (r *RawDirective_sn) Parse(_, data string) (string, error) {
	data, err := r.PreFlight(data)
	if err != nil {
		return "", err
	}

	lines := strings.Split(data, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "sn") && len(line) > 3 {
			return line[3:], nil
		}
	}

	return "", ErrRawDirectiveParseFailed
}

// list_dir 指令
type RawDirective_list_dir struct {
	*RawDirective
	Name string
}

func NewRawDirective_list_dir() *RawDirective_list_dir {
	return &RawDirective_list_dir{Name: "list_dir", RawDirective: &RawDirective{JSONOutput: true}}
}

func (r *RawDirective_list_dir) String() string {
	return r.Name
}

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

// clean_dir 指令
type RawDirective_clean_dir struct {
	*RawDirective
	Name string
}

func NewRawDirective_clean_dir() *RawDirective_clean_dir {
	return &RawDirective_clean_dir{Name: "clean_dir", RawDirective: &RawDirective{JSONOutput: false}}
}

func (r *RawDirective_clean_dir) String() string {
	return r.Name
}

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

// delete_file 指令
type RawDirective_delete_file struct {
	*RawDirective
	Name string
}

func NewRawDirective_delete_file() *RawDirective_delete_file {
	return &RawDirective_delete_file{Name: "delete_file", RawDirective: &RawDirective{JSONOutput: false}}
}

func (r *RawDirective_delete_file) String() string {
	return r.Name
}

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

// alive 指令
type RawDirective_alive struct {
	*RawDirective
	Name string
}

func NewRawDirective_alive() *RawDirective_alive {
	return &RawDirective_alive{Name: "alive", RawDirective: &RawDirective{JSONOutput: true}}
}

func (r *RawDirective_alive) String() string {
	return r.Name
}

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

// atime 指令
type RawDirective_atime struct {
	*RawDirective
	Name string
}

func NewRawDirective_atime() *RawDirective_atime {
	return &RawDirective_atime{Name: "atime", RawDirective: &RawDirective{JSONOutput: true}}
}

func (r *RawDirective_atime) String() string {
	return r.Name
}

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

// aversion 指令
type RawDirective_aversion struct {
	*RawDirective
	Name string
}

func NewRawDirective_aversion() *RawDirective_aversion {
	return &RawDirective_aversion{Name: "aversion", RawDirective: &RawDirective{JSONOutput: true}}
}

func (r *RawDirective_aversion) String() string {
	return r.Name
}

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

// ainsp 指令 返回log基础信息
type RawDirective_ainsp struct {
	*RawDirective
	Name string
}

func NewRawDirective_ainsp() *RawDirective_ainsp {
	return &RawDirective_ainsp{Name: "ainsp", RawDirective: &RawDirective{JSONOutput: true}}
}

func (r *RawDirective_ainsp) String() string {
	return r.Name
}

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

// alog 指令 返回log基础信息
type RawDirective_alog struct {
	*RawDirective
	Name string
}

func NewRawDirective_alog() *RawDirective_alog {
	return &RawDirective_alog{Name: "alog", RawDirective: &RawDirective{JSONOutput: false}}
}

func (r *RawDirective_alog) String() string {
	return r.Name
}

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
