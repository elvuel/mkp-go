package mkpgo

import (
	"fmt"
	"strings"
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
	opt.Button = &v
	return opt
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
