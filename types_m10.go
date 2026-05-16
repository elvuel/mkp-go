package mkpgo

import (
	"encoding/json"
	"fmt"
	"strings"
)

// M10 coordinate and wheel ranges:
// M10 坐标与滚轮范围：x/y 为 -2048~2047，wheel 为 -128~127。
// button uses lower 5 bits:
// button 使用低 5bit（[0~31]），最低位对应左键。
// --p: port #
// --b: botton
// --x: x
// --y: y
// --w: wheel
// --v: verbose, display send buffer. 1: verbose 0 , default, not verbose.

// M10Button represents mouse-button bitmask in m10 directive.
// M10Button 表示 m10 指令中的鼠标按键位掩码。
type M10Button int

// ToString converts M10Button to human-readable button name.
// ToString 将 M10Button 转换为可读按钮名称。
func (b M10Button) ToString() string {
	switch b {
	case ReleaseMouseButton:
		return "none"
	case LeftMouseButton:
		return "left"
	case RightMouseButton:
		return "right"
	case BothLeftRightMouseButton:
		return "both"
	case MiddleMouseButton:
		return "middle"
	case BackMouseButton:
		return "backward"
	case FowardMouseButton:
		return "forward"
	default:
		return ""
	}
}

const (
	ReleaseMouseButton       M10Button = 0
	LeftMouseButton          M10Button = 1
	RightMouseButton         M10Button = 2
	BothLeftRightMouseButton M10Button = 3
	MiddleMouseButton        M10Button = 4  // [4-7]
	BackMouseButton          M10Button = 8  // [8-9]
	FowardMouseButton        M10Button = 16 // [16-31]
)

// Common preset keypad release options.
// 常用的键盘释放预置选项。
var (
	// HidKpadRelease releases current key slots to NONE.
	// HidKpadRelease 将当前键位槽释放为 NONE。
	HidKpadRelease = NewKpadOption().WithDelay(0).WithKey("NONE")
	// HidKpadReleaseAll sends full zero-like release packet.
	// HidKpadReleaseAll 发送全释放（全零态）键盘包。
	HidKpadReleaseAll = NewKpadOption().WithDelay(0).WithRelease(0)
)

// M10Option is the option model for m10 mouse directive.
// M10Option 是 m10 鼠标指令的参数模型。
type M10Option struct {
	Button *int `json:"button,omitempty"`
	X      *int `json:"x,omitempty"`
	Y      *int `json:"y,omitempty"`
	Wheel  *int `json:"wheel,omitempty"`
	Async  bool `json:"async"`
	// --b: botton
	// --x: x
	// --y: y
	// --w: wheel
	// m10 --port 1 --b xx --x xx --y xx --w xx
}

// NewM10Option creates an empty M10Option.
// NewM10Option 创建空的 M10Option。
func NewM10Option() *M10Option {
	return &M10Option{Async: true}
}

type m10OptionJSON struct {
	Button *int  `json:"button,omitempty"`
	X      *int  `json:"x,omitempty"`
	Y      *int  `json:"y,omitempty"`
	Wheel  *int  `json:"wheel,omitempty"`
	Async  *bool `json:"async"`
}

// UnmarshalJSON keeps async default compatible with historical behavior.
// UnmarshalJSON 在未显式提供 async 时保持历史默认异步行为。
func (opt *M10Option) UnmarshalJSON(data []byte) error {
	raw := m10OptionJSON{}
	*opt = *NewM10Option()
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	opt.Button = raw.Button
	opt.X = raw.X
	opt.Y = raw.Y
	opt.Wheel = raw.Wheel
	if raw.Async != nil {
		opt.Async = *raw.Async
	}

	return nil
}

// CheckMouseButton normalizes button text to M10Button.
// CheckMouseButton 将字符串按钮名映射为 M10Button。
func CheckMouseButton(btn string) M10Button {
	switch strings.ToLower(btn) {
	case "left":
		return LeftMouseButton
	case "right":
		return RightMouseButton
	case "both":
		return BothLeftRightMouseButton
	case "middle":
		return MiddleMouseButton
	case "backword":
		return BackMouseButton
	case "forword":
		return FowardMouseButton
	default:
		return ReleaseMouseButton
	}
}

// Reset clears all optional fields.
// Reset 清空当前所有可选参数字段。
func (opt *M10Option) Reset() {
	opt.Button = nil
	opt.X = nil
	opt.Y = nil
	opt.Wheel = nil
}

// SetButton sets button bitmask value directly.
// SetButton 直接设置按钮位掩码值。
func (opt *M10Option) SetButton(v int) *M10Option {
	// 5 bit [0~31], [-16 ~ 15]

	// 0 release

	// Left 键 1

	// Right 键 2

	// Both Left Right 键 3

	// middle 键 4,5,6,7

	// Back 键 8, 9, 10, 11, 12, 13 , 14，15

	// Foward 键 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31

	// 即便 32 开始继续 就是 1 左键
	opt.Button = &v
	return opt
}

// WithButton alias of SetButton
// WithButton is an alias of SetButton.
// WithButton 是 SetButton 的别名。
func (opt *M10Option) WithButton(v int) *M10Option {
	opt.SetButton(v)
	return opt
}

// WithoutButton removes button field from directive.
// WithoutButton 移除按钮参数字段。
func (opt *M10Option) WithoutButton() *M10Option {
	opt.Button = nil
	return opt
}

// NoButton is an alias of WithoutButton.
// NoButton 是 WithoutButton 的别名。
func (opt *M10Option) NoButton() *M10Option {
	return opt.WithoutButton()
}

// WithBothLeftRightButton sets both left and right button bits.
// WithBothLeftRightButton 同时设置左右键按下位。
func (opt *M10Option) WithBothLeftRightButton() *M10Option {
	// 	m10Opt := mkpgo.NewM10Option().SetButton(3).SetX(10).SetY(10)
	// log.Println("----")
	// time.Sleep(5 * time.Second)
	// sfport.Mouse10(m10Opt)
	// log.Println("----")
	// m10Opt.SetX(10).SetY(0).SetButton(3)
	// sfport.Mouse10(m10Opt)
	// time.Sleep(5 * time.Second)
	// log.Println("----")
	// m10Opt.SetX(0).SetY(0).SetButton(0)
	// sfport.Mouse10(m10Opt)

	return opt.WithButton(int(BothLeftRightMouseButton))
}

// WithLeftButton sets left button bit.
// WithLeftButton 设置左键按下位。
func (opt *M10Option) WithLeftButton() *M10Option {
	return opt.WithButton(int(LeftMouseButton))
}

// WithRightButton sets right button bit.
// WithRightButton 设置右键按下位。
func (opt *M10Option) WithRightButton() *M10Option {
	return opt.WithButton(int(RightMouseButton))
}

// WithMiddleButton sets middle button bit.
// WithMiddleButton 设置中键按下位。
func (opt *M10Option) WithMiddleButton() *M10Option {
	return opt.WithButton(int(MiddleMouseButton))
}

// WithBackButton sets back button bit.
// WithBackButton 设置后退键按下位。
func (opt *M10Option) WithBackButton() *M10Option {
	return opt.WithButton(int(BackMouseButton))
}

// WithFowardButton sets forward button bit.
// WithFowardButton 设置前进键按下位。
func (opt *M10Option) WithFowardButton() *M10Option {
	return opt.WithButton(int(FowardMouseButton))
}

// SetX sets relative x movement.
// SetX 设置相对 X 位移。
func (opt *M10Option) SetX(v int) *M10Option {
	opt.X = &v
	return opt
}

// SetY sets relative y movement.
// SetY 设置相对 Y 位移。
func (opt *M10Option) SetY(v int) *M10Option {
	opt.Y = &v
	return opt
}

// SetWheel sets wheel delta.
// SetWheel 设置滚轮位移值。
func (opt *M10Option) SetWheel(v int) *M10Option {
	opt.Wheel = &v
	return opt
}

// WithAsync controls whether m10 uses async send mode.
// WithAsync 控制 m10 是否使用异步发送模式。
func (opt *M10Option) WithAsync(async bool) *M10Option {
	opt.Async = async
	return opt
}

// ToString builds CLI args fragment for m10 option.
// ToString 构建 m10 选项的命令行参数片段。
func (opt *M10Option) ToString() string {
	if opt == nil {
		return ""
	}

	directives := make([]string, 0)
	if opt.Button != nil {
		directives = append(directives, "--b", fmt.Sprintf("%d", *opt.Button))
	}
	if opt.X != nil {
		directives = append(directives, "--x", fmt.Sprintf("%d", *opt.X))
	}
	if opt.Y != nil {
		directives = append(directives, "--y", fmt.Sprintf("%d", *opt.Y))
	}
	if opt.Wheel != nil {
		directives = append(directives, "--w", fmt.Sprintf("%d", *opt.Wheel))
	}
	return strings.Join(directives, " ")
}
