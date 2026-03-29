package mkpgo

// init registers built-in raw directive parsers at package load time.
// init 在包加载时注册内置原始指令解析器。
func init() {
	InitParsers()
}
