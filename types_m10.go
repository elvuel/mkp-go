package mkpgo

import (
	"fmt"
	"strings"
)

// M10: x, y的范围 -2048~2047, wheel -128~127
// button: 低5bit, 最低是 左按键. // 5 bit [0~31], [-16 ~ 15],  32 往后 mod 就是 0, 33(mod 32) -> 1 左键,
// --p: port #
// --b: botton
// --x: x
// --y: y
// --w: wheel
// --v: verbose, display send buffer. 1: verbose 0 , default, not verbose.

type M10Button int

const (
	ReleaseMouseButton       M10Button = 0
	LeftMouseButton          M10Button = 1
	RightMouseButton         M10Button = 2
	BothLeftRightMouseButton M10Button = 3
	MiddleMouseButton        M10Button = 4  // [4-7]
	BackMouseButton          M10Button = 8  // [8-9]
	FowardMouseButton        M10Button = 16 // [16-31]
)

type M10Option struct {
	Button *int
	X      *int
	Y      *int
	Wheel  *int
	// --b: botton
	// --x: x
	// --y: y
	// --w: wheel
	// m10 --port 1 --b xx --x xx --y xx --w xx
}

func NewM10Option() *M10Option {
	return &M10Option{}
}

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
func (opt *M10Option) WithButton(v int) *M10Option {
	opt.SetButton(v)
	return opt
}

func (opt *M10Option) WithoutButton() *M10Option {
	opt.Button = nil
	return opt
}

func (opt *M10Option) NoButton() *M10Option {
	return opt.WithoutButton()
}

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

func (opt *M10Option) WithLeftButton() *M10Option {
	return opt.WithButton(int(LeftMouseButton))
}

func (opt *M10Option) WithRightButton() *M10Option {
	return opt.WithButton(int(RightMouseButton))
}

func (opt *M10Option) WithMiddleButton() *M10Option {
	return opt.WithButton(int(MiddleMouseButton))
}

func (opt *M10Option) WithBackButton() *M10Option {
	return opt.WithButton(int(BackMouseButton))
}

func (opt *M10Option) WithFowardButton() *M10Option {
	return opt.WithButton(int(FowardMouseButton))
}

func (opt *M10Option) SetX(v int) *M10Option {
	opt.X = &v
	return opt
}

func (opt *M10Option) SetY(v int) *M10Option {
	opt.Y = &v
	return opt
}

func (opt *M10Option) SetWheel(v int) *M10Option {
	opt.Wheel = &v
	return opt
}

func (opt *M10Option) ToString() string {
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
