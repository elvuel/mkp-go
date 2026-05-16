package controller

import (
	"context"
	"math/rand"
	"time"

	mkpgo "github.com/elvuel/mkp-go"
	"github.com/elvuel/mkp-go/helper"
)

type Controller struct {
	sfport        *mkpgo.SFSerialPort
	MouseMovement *mkpgo.MouseMovementSimulator
}

// sleepMs sleeps for the given milliseconds when positive.
// sleepMs 在毫秒值大于 0 时休眠指定时长。
func sleepMs(ms int) {
	if ms <= 0 {
		return
	}
	time.Sleep(time.Duration(ms) * time.Millisecond)
}

// NewController creates a controller bound to one serial port.
// NewController 创建并绑定一个串口控制器。
func NewController(sfport *mkpgo.SFSerialPort) *Controller {
	ctrl := &Controller{
		sfport:        sfport,
		MouseMovement: mkpgo.NewMouseMovementSimulator(mkpgo.DefaultMouseMovementSimulatorConfig(), true, true, true),
	}

	ctrl.MouseMovement.SetSFPort(sfport)

	return ctrl
}

// firstKpadOption returns the first provided kpad option, if any.
// firstKpadOption 返回传入的第一个 kpad 配置（若存在）。
func firstKpadOption(opts ...*mkpgo.KpadOption) *mkpgo.KpadOption {
	if len(opts) == 0 {
		return nil
	}
	return opts[0]
}

// firstM10Option returns the first provided m10 option, if any.
// firstM10Option 返回传入的第一个 m10 配置（若存在）。
func firstM10Option(opts ...*mkpgo.M10Option) *mkpgo.M10Option {
	if len(opts) == 0 {
		return nil
	}
	return opts[0]
}

// controllerM10Async resolves async mode for controller mouse helpers.
// controllerM10Async 解析控制器鼠标辅助方法的 async 模式。
func controllerM10Async(opts ...*mkpgo.M10Option) bool {
	if opt := firstM10Option(opts...); opt != nil {
		return opt.Async
	}
	return true
}

// BindSFPort binds a new serial port and updates movement simulator binding.
// BindSFPort 绑定新的串口，并同步更新移动模拟器的端口绑定。
func (c *Controller) BindSFPort(port *mkpgo.SFSerialPort) {
	c.sfport = port
	if c.MouseMovement != nil {
		c.MouseMovement.SetSFPort(port)
	}
}

// Open opens the bound serial port.
// Open 打开已绑定的串口。
func (c *Controller) Open() error {
	return c.sfport.Open()
}

// Close closes the bound serial port.
// Close 关闭已绑定的串口。
func (c *Controller) Close() {
	c.sfport.Close()
}

// helper func StartRecord(sfport *mkpgo.SFSerialPort, logName string, opt *mkpgo.LogOption) error
func (c *Controller) StartRecord(logName string, opt *mkpgo.LogOption) error {
	return helper.StartRecord(c.sfport, logName, opt)
}

// helper func StopRecord(sfport *mkpgo.SFSerialPort) error
func (c *Controller) StopRecord() error {
	return helper.StopRecord(c.sfport)
}

// helper func Alog(sfport *mkpgo.SFSerialPort, logName string, opt *mkpgo.LogOption) (string, error)
func (c *Controller) Alog(logName string, opt *mkpgo.LogOption) (string, error) {
	return helper.Alog(c.sfport, logName, opt)
}

// helper func Astop(sfport *mkpgo.SFSerialPort) error
func (c *Controller) Astop() error {
	return helper.Astop(c.sfport)
}

// helper func Cancel(sfport *mkpgo.SFSerialPort) error
func (c *Controller) Cancel() error {
	return helper.Cancel(c.sfport)
}

// helper func DeviceSN(sfport *mkpgo.SFSerialPort) (*mkpgo.SN, error)
func (c *Controller) DeviceSN() (*mkpgo.SN, error) {
	return helper.DeviceSN(c.sfport)
}

// helper func ListDir(sfport *mkpgo.SFSerialPort, path string) (*mkpgo.FileSystem, error)
func (c *Controller) ListDir(path string) (*mkpgo.FileSystem, error) {
	return helper.ListDir(c.sfport, path)
}

// helper func ComposeLogDirctory(logDir string) string
func (c *Controller) ComposeLogDirctory(logDir string) string {
	return helper.ComposeLogDirctory(logDir)
}

// helper func CleanDir(sfport *mkpgo.SFSerialPort, path string) error
func (c *Controller) CleanDir(path string) error {
	return helper.CleanDir(c.sfport, path)
}

// helper func ComposeLogFullpath(logPath string) string
func (c *Controller) ComposeLogFullpath(logPath string) string {
	return helper.ComposeLogFullpath(logPath)
}

// helper func DeleteFile(sfport *mkpgo.SFSerialPort, path string) error
func (c *Controller) DeleteFile(path string) error {
	return helper.DeleteFile(c.sfport, path)
}

// helper func Alive(sfport *mkpgo.SFSerialPort) (*mkpgo.Heartbeat, error)
func (c *Controller) Alive() (*mkpgo.Heartbeat, error) {
	return helper.Alive(c.sfport)
}

// helper func Atime(sfport *mkpgo.SFSerialPort, path string) (*mkpgo.LogLength, error)
func (c *Controller) Atime(path string) (*mkpgo.LogLength, error) {
	return helper.Atime(c.sfport, path)
}

// helper func Aversion(sfport *mkpgo.SFSerialPort) (*mkpgo.MKPVersion, error)
func (c *Controller) Aversion() (*mkpgo.MKPVersion, error) {
	return helper.Aversion(c.sfport)
}

// helper func AInspect(sfport *mkpgo.SFSerialPort, path string) (*mkpgo.LogInfo, error)
func (c *Controller) AInspect(path string) (*mkpgo.LogInfo, error) {
	return helper.AInspect(c.sfport, path)
}

// KeyDown proxies helper.KeyDown and accepts optional kpad settings.
// KeyDown 代理 helper.KeyDown，并接受可选的 kpad 配置。
func (c *Controller) KeyDown(key string, opts ...*mkpgo.KpadOption) error {
	return helper.KeyDown(c.sfport, key, opts...)
}

// KeyUp proxies helper.KeyUp and accepts optional kpad settings.
// KeyUp 代理 helper.KeyUp，并接受可选的 kpad 配置。
func (c *Controller) KeyUp(key string, opts ...*mkpgo.KpadOption) error {
	return helper.KeyUp(c.sfport, key, opts...)
}

// KeyTap proxies helper.KeyTap and accepts optional kpad settings.
// KeyTap 代理 helper.KeyTap，并接受可选的 kpad 配置。
func (c *Controller) KeyTap(key string, opts ...*mkpgo.KpadOption) error {
	return helper.KeyTap(c.sfport, key, opts...)
}

// KeyPresses proxies helper.KeyPresses and accepts optional kpad settings.
// KeyPresses 代理 helper.KeyPresses，并接受可选的 kpad 配置。
func (c *Controller) KeyPresses(keys []string, sleep int, opts ...*mkpgo.KpadOption) error {
	return helper.KeyPresses(c.sfport, keys, sleep, opts...)
}

// KeypadRelease proxies helper.KeypadRelease and accepts optional kpad settings.
// KeypadRelease 代理 helper.KeypadRelease，并接受可选的 kpad 配置。
func (c *Controller) KeypadRelease(opts ...*mkpgo.KpadOption) error {
	return helper.KeypadRelease(c.sfport, opts...)
}

// KeypadReleaseAll proxies helper.KeypadReleaseAll and accepts optional kpad settings.
// KeypadReleaseAll 代理 helper.KeypadReleaseAll，并接受可选的 kpad 配置。
func (c *Controller) KeypadReleaseAll(opts ...*mkpgo.KpadOption) error {
	return helper.KeypadReleaseAll(c.sfport, opts...)
}

// MouseClick
// MouseClick("left|right|both|middle|backword|forword", true)
func (c *Controller) MouseClick(args ...interface{}) {
	var button int
	button = int(mkpgo.LeftMouseButton)
	var double bool
	var sleepInterval int
	var override *mkpgo.M10Option

	if len(args) > 0 {
		button = int(mkpgo.CheckMouseButton(args[0].(string)))
	}

	if len(args) > 1 {
		double = args[1].(bool)
	}

	if len(args) > 2 {
		sleepInterval = args[2].(int)
	}

	if len(args) > 3 {
		if v, ok := args[3].(*mkpgo.M10Option); ok {
			override = v
		}
	}

	c.MouseClickWithOption(button, double, sleepInterval, override)
}

// MouseClickWithOption clicks one mouse button and optionally performs a double click with m10 override.
// MouseClickWithOption 点击鼠标按键，并可通过 m10 覆盖配置执行双击。
func (c *Controller) MouseClickWithOption(button int, double bool, sleepInterval int, override *mkpgo.M10Option) {
	opt := mkpgo.NewM10Option().WithAsync(controllerM10Async(override))
	opt.WithButton(button).SetX(0).SetY(0)
	c.sfport.Mouse10(opt)

	opt.Reset()
	c.sfport.Mouse10(opt.SetX(0).SetY(0).WithoutButton())

	if double {
		if sleepInterval > 0 {
			sleepMs(sleepInterval)
		} else {
			// rand(50 - 150) + 1
			time.Sleep(time.Duration(rand.Intn(50)+100+1) * time.Millisecond)
		}
		opt.WithButton(button).SetX(0).SetY(0)
		c.sfport.Mouse10(opt)

		opt.Reset()
		c.sfport.Mouse10(opt.SetX(0).SetY(0).WithoutButton())
	}
}

// 直接滚轮滚动
// sleepInterval 为次滚轮间间隔, -1 使用随机间隔
func (c *Controller) MouseScroll(dir string, steps int, sleepInterval int) error {
	return c.MouseScrollWithOption(dir, steps, sleepInterval, nil)
}

// MouseScrollWithOption scrolls the wheel using an optional m10 override.
// MouseScrollWithOption 使用可选的 m10 覆盖配置执行滚轮滚动。
func (c *Controller) MouseScrollWithOption(dir string, steps int, sleepInterval int, override *mkpgo.M10Option) error {
	opt := mkpgo.NewM10Option().WithAsync(controllerM10Async(override))

	mult := 1
	if dir == "up" {
		mult = 1
	} else {
		mult = -1
	}

	for i := 1; i <= steps; i++ {
		opt = opt.SetX(0).SetY(0).SetWheel(mult)
		c.sfport.Mouse10(opt)

		time.Sleep(8 * time.Millisecond) // 配合硬件规格8ms

		opt.Reset()
		c.sfport.Mouse10(opt.SetX(0).SetY(0).WithoutButton())

		if steps > 1 {
			if sleepInterval > 0 {
				sleepMs(sleepInterval)
			} else {
				// rand(50 - 150) + 1
				time.Sleep(time.Duration(rand.Intn(50)+100+1) * time.Millisecond)
			}
		}
	}

	return nil
}

// 鼠标键按下 滚轮滚动
// sleepInterval 为次滚轮间间隔, -1 使用随机间隔
func (c *Controller) MouseScrollWithButton(dir string, steps int, button string, sleepInterval int) error {
	return c.MouseScrollWithButtonOption(dir, steps, button, sleepInterval, nil)
}

// MouseScrollWithButtonOption scrolls the wheel while optionally holding a mouse button with m10 override.
// MouseScrollWithButtonOption 使用 m10 覆盖配置在按住鼠标键时执行滚轮滚动。
func (c *Controller) MouseScrollWithButtonOption(dir string, steps int, button string, sleepInterval int, override *mkpgo.M10Option) error {
	opt := mkpgo.NewM10Option().WithAsync(controllerM10Async(override))

	mult := 1
	if dir == "up" {
		mult = 1
	} else {
		mult = -1
	}

	if button != "" {
		c.MouseDown(button, override)
	}

	for i := 1; i <= steps; i++ {
		opt = opt.SetX(0).SetY(0).SetWheel(mult)

		c.sfport.Mouse10(opt)

		time.Sleep(8 * time.Millisecond) // 配合硬件规格8ms

		if button == "" {
			opt.Reset()
			c.sfport.Mouse10(opt.SetX(0).SetY(0).WithoutButton())
		}

		if steps > 1 {
			if sleepInterval > 0 {
				sleepMs(sleepInterval)
			} else {
				// rand(50 - 150) + 1
				time.Sleep(time.Duration(rand.Intn(50)+100+1) * time.Millisecond)
			}
		}
	}

	if button != "" {
		c.MouseReleaseAll(override)
	}

	return nil
}

// MouseDown presses one mouse button using optional m10 settings.
// MouseDown 使用可选的 m10 配置按下一个鼠标按键。
func (c *Controller) MouseDown(button string, opts ...*mkpgo.M10Option) error {
	opt := mkpgo.NewM10Option().WithAsync(controllerM10Async(opts...))
	opt.WithButton(int(mkpgo.CheckMouseButton(button))).SetX(0).SetY(0)
	c.sfport.Mouse10(opt)
	return nil
}

// MouseReleaseAll releases all mouse buttons using optional m10 settings.
// MouseReleaseAll 使用可选的 m10 配置释放全部鼠标按键。
func (c *Controller) MouseReleaseAll(opts ...*mkpgo.M10Option) error {
	opt := mkpgo.NewM10Option().WithAsync(controllerM10Async(opts...))
	opt.WithoutButton().SetX(0).SetY(0)
	c.sfport.Mouse10(opt)
	return nil
}

// MouseUp is an alias of MouseReleaseAll.
// MouseUp 是 MouseReleaseAll 的别名。
func (c *Controller) MouseUp(opts ...*mkpgo.M10Option) error {
	return c.MouseReleaseAll(opts...)
}

// M10Move sends one prepared m10 directive.
// M10Move 发送一条准备好的 m10 指令。
func (c *Controller) M10Move(opt *mkpgo.M10Option) {
	helper.M10(context.Background(), c.sfport, opt)
}

// MouseMove  Move the mouse to the specified position relative to the current position.
// button is the name of the mouse button to press while moving.
// relX and relY are the relative X and Y coordinates to move to.
// interval is the time to take to move to the new position.
func (c *Controller) MouseMove(button string, relX, relY int, interval time.Duration, opts ...mkpgo.MouseMovementSimulatorOption) error {
	base := c.MouseMovement
	if base == nil {
		base = mkpgo.NewMouseMovementSimulator(mkpgo.DefaultMouseMovementSimulatorConfig(), true, true, true)
	}

	callMovement := *base
	if base.Cfg != nil {
		cfg := *base.Cfg
		callMovement.Cfg = &cfg
	} else {
		callMovement.Cfg = mkpgo.DefaultMouseMovementSimulatorConfig()
	}
	callMovement.SetSFPort(c.sfport)

	if len(opts) > 0 {
		callMovement.ApplyOptions(opts...)
	}
	callMovement.MoveTo(int(mkpgo.CheckMouseButton(button)), [2]float64{0, 0}, [2]float64{float64(relX), float64(relY)}, interval)
	return nil
}
