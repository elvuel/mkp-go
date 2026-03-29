package mkpgo

// Modifier masks for byte-1 in HID keyboard report.
// 修饰符掩码（Modifier Masks）用于 HID 报告第 1 字节。
// NOTE: byte-2 in HID report is reserved and fixed to 0x00.
// 说明：HID 报告第 2 字节为保留位，固定为 0x00。
const (
	KEY_MOD_LCTRL  uint8 = 0x01 // 左Ctrl键修饰符
	KEY_MOD_LSHIFT uint8 = 0x02 // 左Shift键修饰符
	KEY_MOD_LALT   uint8 = 0x04 // 左Alt键修饰符
	KEY_MOD_LMETA  uint8 = 0x08 // 左Meta键修饰符（通常对应Windows键/Command键）
	KEY_MOD_RCTRL  uint8 = 0x10 // 右Ctrl键修饰符
	KEY_MOD_RSHIFT uint8 = 0x20 // 右Shift键修饰符
	KEY_MOD_RALT   uint8 = 0x40 // 右Alt键修饰符
	KEY_MOD_RMETA  uint8 = 0x80 // 右Meta键修饰符（通常对应Windows键/Command键）
)

// Scan codes for key slots in HID keyboard report (typically 6 slots).
// 扫描码（Scan Codes）用于 HID 键盘报告的按键槽位（通常 6 个）。
// 0x00 means no key; overflow fills slots with KEY_ERR_OVF.
// 0x00 表示无按键；按键过多溢出时槽位会填充 KEY_ERR_OVF。
const (
	KEY_NONE    uint8 = 0x00 // 无按键按下
	KEY_ERR_OVF uint8 = 0x01 // 键盘溢出错误（按键过多时所有槽位填充此值，"幽灵键"）
	// 0x02 // 键盘POST自检失败
	// 0x03 // 键盘未定义错误

	// 字母键（a-z/A-Z）
	KEY_A uint8 = 0x04 // 键盘a键（小写）/A键（大写）
	KEY_B uint8 = 0x05 // 键盘b键（小写）/B键（大写）
	KEY_C uint8 = 0x06 // 键盘c键（小写）/C键（大写）
	KEY_D uint8 = 0x07 // 键盘d键（小写）/D键（大写）
	KEY_E uint8 = 0x08 // 键盘e键（小写）/E键（大写）
	KEY_F uint8 = 0x09 // 键盘f键（小写）/F键（大写）
	KEY_G uint8 = 0x0a // 键盘g键（小写）/G键（大写）
	KEY_H uint8 = 0x0b // 键盘h键（小写）/H键（大写）
	KEY_I uint8 = 0x0c // 键盘i键（小写）/I键（大写）
	KEY_J uint8 = 0x0d // 键盘j键（小写）/J键（大写）
	KEY_K uint8 = 0x0e // 键盘k键（小写）/K键（大写）
	KEY_L uint8 = 0x0f // 键盘l键（小写）/L键（大写）
	KEY_M uint8 = 0x10 // 键盘m键（小写）/M键（大写）
	KEY_N uint8 = 0x11 // 键盘n键（小写）/N键（大写）
	KEY_O uint8 = 0x12 // 键盘o键（小写）/O键（大写）
	KEY_P uint8 = 0x13 // 键盘p键（小写）/P键（大写）
	KEY_Q uint8 = 0x14 // 键盘q键（小写）/Q键（大写）
	KEY_R uint8 = 0x15 // 键盘r键（小写）/R键（大写）
	KEY_S uint8 = 0x16 // 键盘s键（小写）/S键（大写）
	KEY_T uint8 = 0x17 // 键盘t键（小写）/T键（大写）
	KEY_U uint8 = 0x18 // 键盘u键（小写）/U键（大写）
	KEY_V uint8 = 0x19 // 键盘v键（小写）/V键（大写）
	KEY_W uint8 = 0x1a // 键盘w键（小写）/W键（大写）
	KEY_X uint8 = 0x1b // 键盘x键（小写）/X键（大写）
	KEY_Y uint8 = 0x1c // 键盘y键（小写）/Y键（大写）
	KEY_Z uint8 = 0x1d // 键盘z键（小写）/Z键（大写）

	// 数字键（1-0，含Shift组合字符）
	KEY_1 uint8 = 0x1e // 键盘1键/!键（Shift+1）
	KEY_2 uint8 = 0x1f // 键盘2键/@键（Shift+2）
	KEY_3 uint8 = 0x20 // 键盘3键/#键（Shift+3）
	KEY_4 uint8 = 0x21 // 键盘4键/$键（Shift+4）
	KEY_5 uint8 = 0x22 // 键盘5键/%键（Shift+5）
	KEY_6 uint8 = 0x23 // 键盘6键/^键（Shift+6）
	KEY_7 uint8 = 0x24 // 键盘7键/&键（Shift+7）
	KEY_8 uint8 = 0x25 // 键盘8键/*键（Shift+8）
	KEY_9 uint8 = 0x26 // 键盘9键/(键（Shift+9）
	KEY_0 uint8 = 0x27 // 键盘0键/)键（Shift+0）

	// 控制键（回车、ESC、退格等）
	KEY_ENTER      uint8 = 0x28 // 键盘回车键（ENTER）
	KEY_ESC        uint8 = 0x29 // 键盘ESC键（ESCAPE）
	KEY_BACKSPACE  uint8 = 0x2a // 键盘退格键（DELETE/Backspace）
	KEY_TAB        uint8 = 0x2b // 键盘Tab键
	KEY_SPACE      uint8 = 0x2c // 键盘空格键（Spacebar）
	KEY_MINUS      uint8 = 0x2d // 键盘-键/_键（Shift+-）
	KEY_EQUAL      uint8 = 0x2e // 键盘=键/+键（Shift+=）
	KEY_LEFTBRACE  uint8 = 0x2f // 键盘[键/{键（Shift+[）
	KEY_RIGHTBRACE uint8 = 0x30 // 键盘]键/}键（Shift+]）
	KEY_BACKSLASH  uint8 = 0x31 // 键盘\键/|键（Shift+\）
	KEY_HASHTILDE  uint8 = 0x32 // 非美式键盘#键/~键（Shift+#）
	KEY_SEMICOLON  uint8 = 0x33 // 键盘;键/:键（Shift+;）
	KEY_APOSTROPHE uint8 = 0x34 // 键盘'键/"键（Shift+'）
	KEY_GRAVE      uint8 = 0x35 // 键盘`键/~键（Shift+`）
	KEY_COMMA      uint8 = 0x36 // 键盘,键/<键（Shift+,）
	KEY_DOT        uint8 = 0x37 // 键盘.键/>键（Shift+.）
	KEY_SLASH      uint8 = 0x38 // 键盘/键/?键（Shift+/）
	KEY_CAPSLOCK   uint8 = 0x39 // 键盘Caps Lock键（大小写锁定）

	// 功能键（F1-F12）
	KEY_F1  uint8 = 0x3a // 键盘F1键
	KEY_F2  uint8 = 0x3b // 键盘F2键
	KEY_F3  uint8 = 0x3c // 键盘F3键
	KEY_F4  uint8 = 0x3d // 键盘F4键
	KEY_F5  uint8 = 0x3e // 键盘F5键
	KEY_F6  uint8 = 0x3f // 键盘F6键
	KEY_F7  uint8 = 0x40 // 键盘F7键
	KEY_F8  uint8 = 0x41 // 键盘F8键
	KEY_F9  uint8 = 0x42 // 键盘F9键
	KEY_F10 uint8 = 0x43 // 键盘F10键
	KEY_F11 uint8 = 0x44 // 键盘F11键
	KEY_F12 uint8 = 0x45 // 键盘F12键

	// 系统/导航键（打印屏幕、滚动锁定、方向键等）
	KEY_SYSRQ      uint8 = 0x46 // 键盘Print Screen键（截屏）
	KEY_SCROLLLOCK uint8 = 0x47 // 键盘Scroll Lock键（滚动锁定）
	KEY_PAUSE      uint8 = 0x48 // 键盘Pause键（暂停）
	KEY_INSERT     uint8 = 0x49 // 键盘Insert键（插入）
	KEY_HOME       uint8 = 0x4a // 键盘Home键（首页）
	KEY_PAGEUP     uint8 = 0x4b // 键盘Page Up键（上一页）
	KEY_DELETE     uint8 = 0x4c // 键盘Delete Forward键（删除后一个字符）
	KEY_END        uint8 = 0x4d // 键盘End键（末尾）
	KEY_PAGEDOWN   uint8 = 0x4e // 键盘Page Down键（下一页）
	KEY_RIGHT      uint8 = 0x4f // 键盘右方向键（→）
	KEY_LEFT       uint8 = 0x50 // 键盘左方向键（←）
	KEY_DOWN       uint8 = 0x51 // 键盘下方向键（↓）
	KEY_UP         uint8 = 0x52 // 键盘上方向键（↑）

	// 小键盘键（Num Lock及数字/符号）
	KEY_NUMLOCK    uint8 = 0x53 // 键盘Num Lock键（小键盘锁定）/Clear键
	KEY_KPSLASH    uint8 = 0x54 // 小键盘/键（除号）
	KEY_KPASTERISK uint8 = 0x55 // 小键盘*键（乘号）
	KEY_KPMINUS    uint8 = 0x56 // 小键盘-键（减号）
	KEY_KPPLUS     uint8 = 0x57 // 小键盘+键（加号）
	KEY_KPENTER    uint8 = 0x58 // 小键盘回车键（ENTER）
	KEY_KP1        uint8 = 0x59 // 小键盘1键/End键（Num Lock关闭时）
	KEY_KP2        uint8 = 0x5a // 小键盘2键/下方向键（Num Lock关闭时）
	KEY_KP3        uint8 = 0x5b // 小键盘3键/Page Down键（Num Lock关闭时）
	KEY_KP4        uint8 = 0x5c // 小键盘4键/左方向键（Num Lock关闭时）
	KEY_KP5        uint8 = 0x5d // 小键盘5键
	KEY_KP6        uint8 = 0x5e // 小键盘6键/右方向键（Num Lock关闭时）
	KEY_KP7        uint8 = 0x5f // 小键盘7键/Home键（Num Lock关闭时）
	KEY_KP8        uint8 = 0x60 // 小键盘8键/上方向键（Num Lock关闭时）
	KEY_KP9        uint8 = 0x61 // 小键盘9键/Page Up键（Num Lock关闭时）
	KEY_KP0        uint8 = 0x62 // 小键盘0键/Insert键（Num Lock关闭时）
	KEY_KPDOT      uint8 = 0x63 // 小键盘.键/Delete键（Num Lock关闭时）

	// 特殊功能键（非美式键盘、电源键等）
	KEY_102ND   uint8 = 0x64 // 非美式键盘\键/|键（Shift+\）
	KEY_COMPOSE uint8 = 0x65 // 键盘Application键（组合键）
	KEY_POWER   uint8 = 0x66 // 键盘Power键（电源）
	KEY_KPEQUAL uint8 = 0x67 // 小键盘=键（等号）

	// 扩展功能键（F13-F24）
	KEY_F13 uint8 = 0x68 // 键盘F13键
	KEY_F14 uint8 = 0x69 // 键盘F14键
	KEY_F15 uint8 = 0x6a // 键盘F15键
	KEY_F16 uint8 = 0x6b // 键盘F16键
	KEY_F17 uint8 = 0x6c // 键盘F17键
	KEY_F18 uint8 = 0x6d // 键盘F18键
	KEY_F19 uint8 = 0x6e // 键盘F19键
	KEY_F20 uint8 = 0x6f // 键盘F20键
	KEY_F21 uint8 = 0x70 // 键盘F21键
	KEY_F22 uint8 = 0x71 // 键盘F22键
	KEY_F23 uint8 = 0x72 // 键盘F23键
	KEY_F24 uint8 = 0x73 // 键盘F24键

	// 应用控制键（执行、帮助、复制粘贴等）
	KEY_OPEN       uint8 = 0x74 // 键盘Execute键（执行）
	KEY_HELP       uint8 = 0x75 // 键盘Help键（帮助）
	KEY_PROPS      uint8 = 0x76 // 键盘Menu键（菜单）
	KEY_FRONT      uint8 = 0x77 // 键盘Select键（选择）
	KEY_STOP       uint8 = 0x78 // 键盘Stop键（停止）
	KEY_AGAIN      uint8 = 0x79 // 键盘Again键（重做）
	KEY_UNDO       uint8 = 0x7a // 键盘Undo键（撤销）
	KEY_CUT        uint8 = 0x7b // 键盘Cut键（剪切）
	KEY_COPY       uint8 = 0x7c // 键盘Copy键（复制）
	KEY_PASTE      uint8 = 0x7d // 键盘Paste键（粘贴）
	KEY_FIND       uint8 = 0x7e // 键盘Find键（查找）
	KEY_MUTE       uint8 = 0x7f // 键盘Mute键（静音）
	KEY_VOLUMEUP   uint8 = 0x80 // 键盘Volume Up键（音量+）
	KEY_VOLUMEDOWN uint8 = 0x81 // 键盘Volume Down键（音量-）
	// 0x82 // 键盘Locking Caps Lock键（锁定式大小写锁定）
	// 0x83 // 键盘Locking Num Lock键（锁定式小键盘锁定）
	// 0x84 // 键盘Locking Scroll Lock键（锁定式滚动锁定）
	KEY_KPCOMMA uint8 = 0x85 // 小键盘Comma键（逗号）
	// 0x86 // 小键盘Equal Sign键（等号，扩展）

	// 国际键盘键（多语言支持）
	KEY_RO               uint8 = 0x87 // 键盘International1键（RO键，日语键盘）
	KEY_KATAKANAHIRAGANA uint8 = 0x88 // 键盘International2键（片假名/平假名切换）
	KEY_YEN              uint8 = 0x89 // 键盘International3键（日元符号键）
	KEY_HENKAN           uint8 = 0x8a // 键盘International4键（变换键，日语键盘）
	KEY_MUHENKAN         uint8 = 0x8b // 键盘International5键（无变换键，日语键盘）
	KEY_KPJPCOMMA        uint8 = 0x8c // 键盘International6键（日语小键盘逗号）
	// 0x8d // 键盘International7键（保留）
	// 0x8e // 键盘International8键（保留）
	// 0x8f // 键盘International9键（保留）

	// 语言切换键（LANG1-LANG5）
	KEY_HANGEUL        uint8 = 0x90 // 键盘LANG1键（韩语输入法切换）
	KEY_HANJA          uint8 = 0x91 // 键盘LANG2键（韩语汉字切换）
	KEY_KATAKANA       uint8 = 0x92 // 键盘LANG3键（片假名切换）
	KEY_HIRAGANA       uint8 = 0x93 // 键盘LANG4键（平假名切换）
	KEY_ZENKAKUHANKAKU uint8 = 0x94 // 键盘LANG5键（全角/半角切换）
	// 0x95 // 键盘LANG6键（保留）
	// 0x96 // 键盘LANG7键（保留）
	// 0x97 // 键盘LANG8键（保留）
	// 0x98 // 键盘LANG9键（保留）
	// 0x99 // 键盘Alternate Erase键（备用删除）
	// 0x9a // 键盘SysReq/Attention键（系统请求/注意）
	// 0x9b // 键盘Cancel键（取消）
	// 0x9c // 键盘Clear键（清除）
	// 0x9d // 键盘Prior键（优先）
	// 0x9e // 键盘Return键（回车，扩展）
	// 0x9f // 键盘Separator键（分隔符）
	// 0xa0 // 键盘Out键（输出）
	// 0xa1 // 键盘Oper键（操作）
	// 0xa2 // 键盘Clear/Again键（清除/重做）
	// 0xa3 // 键盘CrSel/Props键（光标选择/属性）
	// 0xa4 // 键盘ExSel键（扩展选择）

	// 扩展小键盘键（保留部分未定义）
	// 0xb0 // 小键盘00键
	// 0xb1 // 小键盘000键
	// 0xb2 // 千位分隔符键
	// 0xb3 // 小数点分隔符键
	// 0xb4 // 货币单位键
	// 0xb5 // 货币子单位键
	KEY_KPLEFTPAREN  uint8 = 0xb6 // 小键盘(键（左括号）
	KEY_KPRIGHTPAREN uint8 = 0xb7 // 小键盘)键（右括号）
	// 0xb8 // 小键盘{键（左大括号）
	// 0xb9 // 小键盘}键（右大括号）
	// 0xba // 小键盘Tab键
	// 0xbb // 小键盘Backspace键
	// 0xbc // 小键盘A键
	// 0xbd // 小键盘B键
	// 0xbe // 小键盘C键
	// 0xbf // 小键盘D键
	// 0xc0 // 小键盘E键
	// 0xc1 // 小键盘F键
	// 0xc2 // 小键盘XOR键（异或）
	// 0xc3 // 小键盘^键（异或）
	// 0xc4 // 小键盘%键（百分号）
	// 0xc5 // 小键盘<键（小于）
	// 0xc6 // 小键盘>键（大于）
	// 0xc7 // 小键盘&键（与）
	// 0xc8 // 小键盘&&键（逻辑与）
	// 0xc9 // 小键盘|键（或）
	// 0xca // 小键盘||键（逻辑或）
	// 0xcb // 小键盘:键（冒号）
	// 0xcc // 小键盘#键（井号）
	// 0xcd // 小键盘Space键（空格）
	// 0xce // 小键盘@键（@符号）
	// 0xcf // 小键盘!键（感叹号）
	// 0xd0 // 小键盘Memory Store键（记忆存储）
	// 0xd1 // 小键盘Memory Recall键（记忆读取）
	// 0xd2 // 小键盘Memory Clear键（记忆清除）
	// 0xd3 // 小键盘Memory Add键（记忆加）
	// 0xd4 // 小键盘Memory Subtract键（记忆减）
	// 0xd5 // 小键盘Memory Multiply键（记忆乘）
	// 0xd6 // 小键盘Memory Divide键（记忆除）
	// 0xd7 // 小键盘+/-键（正负切换）
	// 0xd8 // 小键盘Clear键（清除）
	// 0xd9 // 小键盘Clear Entry键（清除输入）
	// 0xda // 小键盘Binary键（二进制）
	// 0xdb // 小键盘Octal键（八进制）
	// 0xdc // 小键盘Decimal键（十进制）
	// 0xdd // 小键盘Hexadecimal键（十六进制）

	// 左右独立控制键（非修饰符，单独触发）
	KEY_LEFTCTRL   uint8 = 0xe0 // 键盘左Ctrl键（独立触发，非修饰符）
	KEY_LEFTSHIFT  uint8 = 0xe1 // 键盘左Shift键（独立触发，非修饰符）
	KEY_LEFTALT    uint8 = 0xe2 // 键盘左Alt键（独立触发，非修饰符）
	KEY_LEFTMETA   uint8 = 0xe3 // 键盘左Meta键（独立触发，非修饰符）
	KEY_RIGHTCTRL  uint8 = 0xe4 // 键盘右Ctrl键（独立触发，非修饰符）
	KEY_RIGHTSHIFT uint8 = 0xe5 // 键盘右Shift键（独立触发，非修饰符）
	KEY_RIGHTALT   uint8 = 0xe6 // 键盘右Alt键（独立触发，非修饰符）
	KEY_RIGHTMETA  uint8 = 0xe7 // 键盘右Meta键（独立触发，非修饰符）

	// 媒体控制键
	KEY_MEDIA_PLAYPAUSE    uint8 = 0xe8 // 媒体播放/暂停键
	KEY_MEDIA_STOPCD       uint8 = 0xe9 // 媒体停止播放键（CD/音频）
	KEY_MEDIA_PREVIOUSSONG uint8 = 0xea // 媒体上一曲键
	KEY_MEDIA_NEXTSONG     uint8 = 0xeb // 媒体下一曲键
	KEY_MEDIA_EJECTCD      uint8 = 0xec // 媒体弹出CD键
	KEY_MEDIA_VOLUMEUP     uint8 = 0xed // 媒体音量+键
	KEY_MEDIA_VOLUMEDOWN   uint8 = 0xee // 媒体音量-键
	KEY_MEDIA_MUTE         uint8 = 0xef // 媒体静音键
	KEY_MEDIA_WWW          uint8 = 0xf0 // 媒体打开浏览器键
	KEY_MEDIA_BACK         uint8 = 0xf1 // 媒体后退键（浏览器/播放器）
	KEY_MEDIA_FORWARD      uint8 = 0xf2 // 媒体前进键（浏览器/播放器）
	KEY_MEDIA_STOP         uint8 = 0xf3 // 媒体停止键（通用）
	KEY_MEDIA_FIND         uint8 = 0xf4 // 媒体查找键（浏览器/文件）
	KEY_MEDIA_SCROLLUP     uint8 = 0xf5 // 媒体向上滚动键
	KEY_MEDIA_SCROLLDOWN   uint8 = 0xf6 // 媒体向下滚动键
	KEY_MEDIA_EDIT         uint8 = 0xf7 // 媒体编辑键（打开编辑器）
	KEY_MEDIA_SLEEP        uint8 = 0xf8 // 媒体睡眠键（系统休眠）
	KEY_MEDIA_COFFEE       uint8 = 0xf9 // 媒体咖啡键（系统不休眠）
	KEY_MEDIA_REFRESH      uint8 = 0xfa // 媒体刷新键（浏览器）
	KEY_MEDIA_CALC         uint8 = 0xfb // 媒体计算器键（打开计算器）
)
