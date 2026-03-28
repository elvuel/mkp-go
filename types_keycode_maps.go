package mkpgo

import (
	"fmt"
	"sync"
)

// Package usbhid 提供USB HID键盘扫描码的双向映射（键名↔扫描码）
// 转化自C语言头文件：https://source.android.com/devices/input/keyboard-devices.html
// 原作者：MightyPork (2016)，Public Domain（公有领域）

// 全局映射初始化锁（确保并发安全）
var keyCodeInitOnce sync.Once

// KeyNameToHexCode 键名→扫描码映射（键名为原C文件的宏名，如"MOD_LCTRL"）
// var KeyNameToHexCode map[string]int
var KeyNameToHexCode map[string]string
var KeyNameToNCode map[string]int

// KeyCodeToName 扫描码→键名及描述映射（值包含键名和原注释说明，便于反向查询）
var KeyCodeToName map[int]string

// init 初始化全局映射（确保仅执行一次）
func init() {
	keyCodeInitOnce.Do(func() {
		// 初始化键名→扫描码映射
		KeyNameToHexCode = make(map[string]string)
		KeyNameToNCode = make(map[string]int)
		// 初始化扫描码→键名及描述映射
		KeyCodeToName = make(map[int]string)

		// -------------------------- 1. 修饰符掩码（Modifier Masks）--------------------------
		addKeycodeMapping(
			"MOD_LCTRL", 0x01, "左Ctrl键修饰符（HID报告第1字节）",
			"MOD_LSHIFT", 0x02, "左Shift键修饰符（HID报告第1字节）",
			"MOD_LALT", 0x04, "左Alt键修饰符（HID报告第1字节）",
			"MOD_LMETA", 0x08, "左Meta键修饰符（通常对应Windows/Command键，HID报告第1字节）",
			"MOD_RCTRL", 0x10, "右Ctrl键修饰符（HID报告第1字节）",
			"MOD_RSHIFT", 0x20, "右Shift键修饰符（HID报告第1字节）",
			"MOD_RALT", 0x40, "右Alt键修饰符（HID报告第1字节）",
			"MOD_RMETA", 0x80, "右Meta键修饰符（通常对应Windows/Command键，HID报告第1字节）",
		)

		// -------------------------- 2. 基础扫描码（无按键/溢出/字母键）--------------------------
		addKeycodeMapping(
			"NONE", 0x00, "无按键按下（扫描码槽位填充值）",
			"ERR_OVF", 0x01, "键盘溢出错误（按键过多时所有槽位填充，\"幽灵键\"）",
			"A", 0x04, "Keyboard a and A（字母A键，大小写切换）",
			"B", 0x05, "Keyboard b and B（字母B键，大小写切换）",
			"C", 0x06, "Keyboard c and C（字母C键，大小写切换）",
			"D", 0x07, "Keyboard d and D（字母D键，大小写切换）",
			"E", 0x08, "Keyboard e and E（字母E键，大小写切换）",
			"F", 0x09, "Keyboard f and F（字母F键，大小写切换）",
			"G", 0x0a, "Keyboard g and G（字母G键，大小写切换）",
			"H", 0x0b, "Keyboard h and H（字母H键，大小写切换）",
			"I", 0x0c, "Keyboard i and I（字母I键，大小写切换）",
			"J", 0x0d, "Keyboard j and J（字母J键，大小写切换）",
			"K", 0x0e, "Keyboard k and K（字母K键，大小写切换）",
			"L", 0x0f, "Keyboard l and L（字母L键，大小写切换）",
			"M", 0x10, "Keyboard m and M（字母M键，大小写切换）",
			"N", 0x11, "Keyboard n and N（字母N键，大小写切换）",
			"O", 0x12, "Keyboard o and O（字母O键，大小写切换）",
			"P", 0x13, "Keyboard p and P（字母P键，大小写切换）",
			"Q", 0x14, "Keyboard q and Q（字母Q键，大小写切换）",
			"R", 0x15, "Keyboard r and R（字母R键，大小写切换）",
			"S", 0x16, "Keyboard s and S（字母S键，大小写切换）",
			"T", 0x17, "Keyboard t and T（字母T键，大小写切换）",
			"U", 0x18, "Keyboard u and U（字母U键，大小写切换）",
			"V", 0x19, "Keyboard v and V（字母V键，大小写切换）",
			"W", 0x1a, "Keyboard w and W（字母W键，大小写切换）",
			"X", 0x1b, "Keyboard x and X（字母X键，大小写切换）",
			"Y", 0x1c, "Keyboard y and Y（字母Y键，大小写切换）",
			"Z", 0x1d, "Keyboard z and Z（字母Z键，大小写切换）",
		)

		// -------------------------- 3. 数字键（含Shift组合字符）--------------------------
		addKeycodeMapping(
			"1", 0x1e, "Keyboard 1 and !（数字1键，Shift+1为!）",
			"2", 0x1f, "Keyboard 2 and @（数字2键，Shift+2为@）",
			"3", 0x20, "Keyboard 3 and #（数字3键，Shift+3为#）",
			"4", 0x21, "Keyboard 4 and $（数字4键，Shift+4为$）",
			"5", 0x22, "Keyboard 5 and %（数字5键，Shift+5为%）",
			"6", 0x23, "Keyboard 6 and ^（数字6键，Shift+6为^）",
			"7", 0x24, "Keyboard 7 and &（数字7键，Shift+7为&）",
			"8", 0x25, "Keyboard 8 and *（数字8键，Shift+8为*）",
			"9", 0x26, "Keyboard 9 and (（数字9键，Shift+9为(）",
			"0", 0x27, "Keyboard 0 and )（数字0键，Shift+0为)）",
		)

		// -------------------------- 4. 控制键（回车/ESC/退格等）--------------------------
		addKeycodeMapping(
			"ENTER", 0x28, "Keyboard Return (ENTER)（回车键）",
			"ESC", 0x29, "Keyboard ESCAPE（ESC键）",
			"BACKSPACE", 0x2a, "Keyboard DELETE (Backspace)（退格键）",
			"TAB", 0x2b, "Keyboard Tab（Tab键）",
			"SPACE", 0x2c, "Keyboard Spacebar（空格键）",
			"MINUS", 0x2d, "Keyboard - and _（减号键，Shift+-为_）",
			"EQUAL", 0x2e, "Keyboard = and +（等号键，Shift+=为+）",
			"LEFTBRACE", 0x2f, "Keyboard [ and {（左中括号键，Shift+[为{）",
			"RIGHTBRACE", 0x30, "Keyboard ] and }（右中括号键，Shift+]为}）",
			"BACKSLASH", 0x31, "Keyboard \\ and |（反斜杠键，Shift+\\为|）",
			"HASHTILDE", 0x32, "Keyboard Non-US # and ~（非美式键盘#键，Shift+#为~）",
			"SEMICOLON", 0x33, "Keyboard ; and :（分号键，Shift+;为:）",
			"APOSTROPHE", 0x34, "Keyboard ' and \"（单引号键，Shift+'为\"）",
			"GRAVE", 0x35, "Keyboard ` and ~（反引号键，Shift+`为~）",
			"COMMA", 0x36, "Keyboard , and <（逗号键，Shift+,为<）",
			"DOT", 0x37, "Keyboard . and >（句号键，Shift+.为>）",
			"SLASH", 0x38, "Keyboard / and ?（斜杠键，Shift+/为?）",
			"CAPSLOCK", 0x39, "Keyboard Caps Lock（大小写锁定键）",
		)

		// -------------------------- 5. 功能键（F1-F24）--------------------------
		addKeycodeMapping(
			"F1", 0x3a, "Keyboard F1（F1功能键）",
			"F2", 0x3b, "Keyboard F2（F2功能键）",
			"F3", 0x3c, "Keyboard F3（F3功能键）",
			"F4", 0x3d, "Keyboard F4（F4功能键）",
			"F5", 0x3e, "Keyboard F5（F5功能键）",
			"F6", 0x3f, "Keyboard F6（F6功能键）",
			"F7", 0x40, "Keyboard F7（F7功能键）",
			"F8", 0x41, "Keyboard F8（F8功能键）",
			"F9", 0x42, "Keyboard F9（F9功能键）",
			"F10", 0x43, "Keyboard F10（F10功能键）",
			"F11", 0x44, "Keyboard F11（F11功能键）",
			"F12", 0x45, "Keyboard F12（F12功能键）",
			"F13", 0x68, "Keyboard F13（F13功能键）",
			"F14", 0x69, "Keyboard F14（F14功能键）",
			"F15", 0x6a, "Keyboard F15（F15功能键）",
			"F16", 0x6b, "Keyboard F16（F16功能键）",
			"F17", 0x6c, "Keyboard F17（F17功能键）",
			"F18", 0x6d, "Keyboard F18（F18功能键）",
			"F19", 0x6e, "Keyboard F19（F19功能键）",
			"F20", 0x6f, "Keyboard F20（F20功能键）",
			"F21", 0x70, "Keyboard F21（F21功能键）",
			"F22", 0x71, "Keyboard F22（F22功能键）",
			"F23", 0x72, "Keyboard F23（F23功能键）",
			"F24", 0x73, "Keyboard F24（F24功能键）",
		)

		// -------------------------- 6. 系统/导航键（打印屏幕/方向键等）--------------------------
		addKeycodeMapping(
			"SYSRQ", 0x46, "Keyboard Print Screen（截屏键）",
			"SCROLLLOCK", 0x47, "Keyboard Scroll Lock（滚动锁定键）",
			"PAUSE", 0x48, "Keyboard Pause（暂停键）",
			"INSERT", 0x49, "Keyboard Insert（插入键）",
			"HOME", 0x4a, "Keyboard Home（首页键）",
			"PAGEUP", 0x4b, "Keyboard Page Up（上一页键）",
			"DELETE", 0x4c, "Keyboard Delete Forward（删除后一个字符键）",
			"END", 0x4d, "Keyboard End（末尾键）",
			"PAGEDOWN", 0x4e, "Keyboard Page Down（下一页键）",
			"RIGHT", 0x4f, "Keyboard Right Arrow（右方向键→）",
			"LEFT", 0x50, "Keyboard Left Arrow（左方向键←）",
			"DOWN", 0x51, "Keyboard Down Arrow（下方向键↓）",
			"UP", 0x52, "Keyboard Up Arrow（上方向键↑）",
		)

		// -------------------------- 7. 小键盘键（Num Lock及数字/符号）--------------------------
		addKeycodeMapping(
			"NUMLOCK", 0x53, "Keyboard Num Lock and Clear（小键盘锁定/清除键）",
			"KPSLASH", 0x54, "Keypad /（小键盘除号键）",
			"KPASTERISK", 0x55, "Keypad *（小键盘乘号键）",
			"KPMINUS", 0x56, "Keypad -（小键盘减号键）",
			"KPPLUS", 0x57, "Keypad +（小键盘加号键）",
			"KPENTER", 0x58, "Keypad ENTER（小键盘回车键）",
			"KP1", 0x59, "Keypad 1 and End（小键盘1键，Num Lock关闭时为End）",
			"KP2", 0x5a, "Keypad 2 and Down Arrow（小键盘2键，Num Lock关闭时为下方向键）",
			"KP3", 0x5b, "Keypad 3 and PageDn（小键盘3键，Num Lock关闭时为Page Down）",
			"KP4", 0x5c, "Keypad 4 and Left Arrow（小键盘4键，Num Lock关闭时为左方向键）",
			"KP5", 0x5d, "Keypad 5（小键盘5键）",
			"KP6", 0x5e, "Keypad 6 and Right Arrow（小键盘6键，Num Lock关闭时为右方向键）",
			"KP7", 0x5f, "Keypad 7 and Home（小键盘7键，Num Lock关闭时为Home）",
			"KP8", 0x60, "Keypad 8 and Up Arrow（小键盘8键，Num Lock关闭时为上方向键）",
			"KP9", 0x61, "Keypad 9 and Page Up（小键盘9键，Num Lock关闭时为Page Up）",
			"KP0", 0x62, "Keypad 0 and Insert（小键盘0键，Num Lock关闭时为Insert）",
			"KPDOT", 0x63, "Keypad . and Delete（小键盘句号键，Num Lock关闭时为Delete）",
			"KPCOMMA", 0x85, "Keypad Comma（小键盘逗号键）",
			"KPEQUAL", 0x67, "Keypad =（小键盘等号键）",
			"KPLEFTPAREN", 0xb6, "Keypad (（小键盘左括号键）",
			"KPRIGHTPAREN", 0xb7, "Keypad )（小键盘右括号键）",
		)

		// -------------------------- 8. 特殊功能键（非美式键盘/电源等）--------------------------
		addKeycodeMapping(
			"102ND", 0x64, "Keyboard Non-US \\ and |（非美式键盘反斜杠键，Shift+\\为|）",
			"COMPOSE", 0x65, "Keyboard Application（组合键/应用键）",
			"POWER", 0x66, "Keyboard Power（电源键）",
		)

		// -------------------------- 9. 应用控制键（复制/粘贴/帮助等）--------------------------
		addKeycodeMapping(
			"OPEN", 0x74, "Keyboard Execute（执行键）",
			"HELP", 0x75, "Keyboard Help（帮助键）",
			"PROPS", 0x76, "Keyboard Menu（菜单键）",
			"FRONT", 0x77, "Keyboard Select（选择键）",
			"STOP", 0x78, "Keyboard Stop（停止键）",
			"AGAIN", 0x79, "Keyboard Again（重做键）",
			"UNDO", 0x7a, "Keyboard Undo（撤销键）",
			"CUT", 0x7b, "Keyboard Cut（剪切键）",
			"COPY", 0x7c, "Keyboard Copy（复制键）",
			"PASTE", 0x7d, "Keyboard Paste（粘贴键）",
			"FIND", 0x7e, "Keyboard Find（查找键）",
			"MUTE", 0x7f, "Keyboard Mute（静音键）",
			"VOLUMEUP", 0x80, "Keyboard Volume Up（音量+键）",
			"VOLUMEDOWN", 0x81, "Keyboard Volume Down（音量-键）",
		)

		// -------------------------- 10. 国际键盘键（多语言支持）--------------------------
		addKeycodeMapping(
			"RO", 0x87, "Keyboard International1（RO键，日语键盘）",
			"KATAKANAHIRAGANA", 0x88, "Keyboard International2（片假名/平假名切换键）",
			"YEN", 0x89, "Keyboard International3（日元符号键）",
			"HENKAN", 0x8a, "Keyboard International4（变换键，日语键盘）",
			"MUHENKAN", 0x8b, "Keyboard International5（无变换键，日语键盘）",
			"KPJPCOMMA", 0x8c, "Keyboard International6（日语小键盘逗号键）",
			"HANGEUL", 0x90, "Keyboard LANG1（韩语输入法切换键）",
			"HANJA", 0x91, "Keyboard LANG2（韩语汉字切换键）",
			"KATAKANA", 0x92, "Keyboard LANG3（片假名切换键）",
			"HIRAGANA", 0x93, "Keyboard LANG4（平假名切换键）",
			"ZENKAKUHANKAKU", 0x94, "Keyboard LANG5（全角/半角切换键）",
		)

		// -------------------------- 11. 左右独立控制键（非修饰符）--------------------------
		addKeycodeMapping(
			"LEFTCTRL", 0xe0, "Keyboard Left Control（左Ctrl键，独立触发，非修饰符）",
			"LEFTSHIFT", 0xe1, "Keyboard Left Shift（左Shift键，独立触发，非修饰符）",
			"LEFTALT", 0xe2, "Keyboard Left Alt（左Alt键，独立触发，非修饰符）",
			"LEFTMETA", 0xe3, "Keyboard Left GUI（左Meta键，独立触发，非修饰符）",
			"RIGHTCTRL", 0xe4, "Keyboard Right Control（右Ctrl键，独立触发，非修饰符）",
			"RIGHTSHIFT", 0xe5, "Keyboard Right Shift（右Shift键，独立触发，非修饰符）",
			"RIGHTALT", 0xe6, "Keyboard Right Alt（右Alt键，独立触发，非修饰符）",
			"RIGHTMETA", 0xe7, "Keyboard Right GUI（右Meta键，独立触发，非修饰符）",
		)

		// -------------------------- 12. 媒体控制键--------------------------
		addKeycodeMapping(
			"MEDIA_PLAYPAUSE", 0xe8, "媒体播放/暂停键",
			"MEDIA_STOPCD", 0xe9, "媒体停止播放键（CD/音频）",
			"MEDIA_PREVIOUSSONG", 0xea, "媒体上一曲键",
			"MEDIA_NEXTSONG", 0xeb, "媒体下一曲键",
			"MEDIA_EJECTCD", 0xec, "媒体弹出CD键",
			"MEDIA_VOLUMEUP", 0xed, "媒体音量+键",
			"MEDIA_VOLUMEDOWN", 0xee, "媒体音量-键",
			"MEDIA_MUTE", 0xef, "媒体静音键",
			"MEDIA_WWW", 0xf0, "媒体打开浏览器键",
			"MEDIA_BACK", 0xf1, "媒体后退键（浏览器/播放器）",
			"MEDIA_FORWARD", 0xf2, "媒体前进键（浏览器/播放器）",
			"MEDIA_STOP", 0xf3, "媒体停止键（通用）",
			"MEDIA_FIND", 0xf4, "媒体查找键（浏览器/文件）",
			"MEDIA_SCROLLUP", 0xf5, "媒体向上滚动键",
			"MEDIA_SCROLLDOWN", 0xf6, "媒体向下滚动键",
			"MEDIA_EDIT", 0xf7, "媒体编辑键（打开编辑器）",
			"MEDIA_SLEEP", 0xf8, "媒体睡眠键（系统休眠）",
			"MEDIA_COFFEE", 0xf9, "媒体咖啡键（系统不休眠）",
			"MEDIA_REFRESH", 0xfa, "媒体刷新键（浏览器）",
			"MEDIA_CALC", 0xfb, "媒体计算器键（打开计算器）",
		)
	})
}

// addKeycodeMapping 辅助函数：批量添加键名→扫描码、扫描码→键名及描述的映射
// 参数格式：keyName1, code1, desc1, keyName2, code2, desc2, ...
func addKeycodeMapping(params ...interface{}) {
	for i := 0; i < len(params); i += 3 {
		keyName := params[i].(string)
		code := params[i+1].(int)
		desc := params[i+2].(string)

		// 填充键名→扫描码映射
		KeyNameToHexCode[keyName] = fmt.Sprintf("0x%02x", code)
		KeyNameToNCode[keyName] = code
		// 填充扫描码→键名及描述映射（格式："键名（描述）"）
		KeyCodeToName[code] = keyName + "（" + desc + "）"
	}
}
