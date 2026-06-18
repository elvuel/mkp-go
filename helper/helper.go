package helper

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	mkpgo "github.com/elvuel/mkp-go"
)

func StopRecord(sfport *mkpgo.SFSerialPort) error {
	return StopRecordContext(context.Background(), sfport)
}

func StopRecordContext(ctx context.Context, sfport *mkpgo.SFSerialPort) error {
	return sfport.StopRecordingContext(ctx)
}

func StartRecord(sfport *mkpgo.SFSerialPort, logName string, opt *mkpgo.LogOption) error {
	return StartRecordContext(context.Background(), sfport, logName, opt)
}

func StartRecordContext(ctx context.Context, sfport *mkpgo.SFSerialPort, logName string, opt *mkpgo.LogOption) error {
	args := make([]string, 0)
	args = append(args, logName)

	if opt != nil {
		args = append(args, opt.CliArgs()...)
	}
	return sfport.StartRecordingContext(ctx, strings.Join(args, " "))
}

func Alog(sfport *mkpgo.SFSerialPort, logName string, opt *mkpgo.LogOption, opts ...mkpgo.DirectiveOption) (string, error) {
	return AlogContext(context.Background(), sfport, logName, opt, opts...)
}

func AlogContext(ctx context.Context, sfport *mkpgo.SFSerialPort, logName string, opt *mkpgo.LogOption, opts ...mkpgo.DirectiveOption) (string, error) {
	if !sfport.SyncOuputEnabled {
		return "", errors.New("please enable sync mode first")
	}

	args := make([]string, 0)
	args = append(args, logName)

	if opt != nil {
		args = append(args, opt.CliArgs()...)
	}

	directive := "alog " + strings.Join(args, " ")
	fmt.Println(directive)

	result, err := sfport.SendDirectiveContext(ctx, directive, opts...)

	// log.Println("got ################ alog response:", result)

	if err != nil {
		return "", err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return "", err
		}

		return parsedResult, nil

	}

	return "", mkpgo.ErrDirectiveParserMissing
}

func Astop(sfport *mkpgo.SFSerialPort, opts ...mkpgo.DirectiveOption) error {
	return AstopContext(context.Background(), sfport, opts...)
}

func AstopContext(ctx context.Context, sfport *mkpgo.SFSerialPort, opts ...mkpgo.DirectiveOption) error {
	if !sfport.SyncOuputEnabled {
		return errors.New("please enable sync mode first")
	}
	directive := "astop"

	result, err := sfport.SendDirectiveContext(ctx, directive, opts...)

	if err != nil {
		return err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		_, err := parser.Parse(directive, result)
		return err
	}

	return mkpgo.ErrDirectiveParserMissing
}

func Cancel(sfport *mkpgo.SFSerialPort) error {
	return CancelContext(context.Background(), sfport)
}

func CancelContext(ctx context.Context, sfport *mkpgo.SFSerialPort) error {
	return sfport.CancelReplayContext(ctx)
}

// Join connects the device to Wi-Fi using join directive.
// Join 使用 join 指令连接 Wi-Fi；opt 为 nil 或空时使用最近保存的 Wi-Fi 配置；opts 可覆盖本次同步等待设置。
func Join(sfport *mkpgo.SFSerialPort, opt *mkpgo.JoinOption, opts ...mkpgo.DirectiveOption) (string, error) {
	return JoinContext(context.Background(), sfport, opt, opts...)
}

// JoinContext connects the device to Wi-Fi using join directive with context.
// JoinContext 使用 join 指令和 context 连接 Wi-Fi；opt 为 nil 或空时使用最近保存的 Wi-Fi 配置；opts 可覆盖本次同步等待设置。
func JoinContext(ctx context.Context, sfport *mkpgo.SFSerialPort, opt *mkpgo.JoinOption, opts ...mkpgo.DirectiveOption) (string, error) {
	if !sfport.SyncOuputEnabled {
		return "", errors.New("please enable sync mode first")
	}

	directive := "join"
	if args := opt.CliArgs(); len(args) > 0 {
		directive += " " + strings.Join(args, " ")
	}

	result, err := sfport.SendDirectiveContext(ctx, directive, opts...)
	if err != nil {
		return "", err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		return parser.Parse(directive, result)
	}

	return "", mkpgo.ErrDirectiveParserMissing
}

// DeviceSN 指令 返回设备序列号
func DeviceSN(sfport *mkpgo.SFSerialPort, opts ...mkpgo.DirectiveOption) (*mkpgo.SN, error) {
	return DeviceSNContext(context.Background(), sfport, opts...)
}

// DeviceSNContext 指令 返回设备序列号
func DeviceSNContext(ctx context.Context, sfport *mkpgo.SFSerialPort, opts ...mkpgo.DirectiveOption) (*mkpgo.SN, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "sn"

	result, err := sfport.SendDirectiveContext(ctx, directive, opts...)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			sn := &mkpgo.SN{}
			err = parser.UnmarshalTo(parsedResult, sn)
			return sn, err
		}
	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

// ListDir 指令 返回路径下的所有子目录及文件
func ListDir(sfport *mkpgo.SFSerialPort, path string, opts ...mkpgo.DirectiveOption) (*mkpgo.FileSystem, error) {
	return ListDirContext(context.Background(), sfport, path, opts...)
}

// ListDirContext 指令 返回路径下的所有子目录及文件
func ListDirContext(ctx context.Context, sfport *mkpgo.SFSerialPort, path string, opts ...mkpgo.DirectiveOption) (*mkpgo.FileSystem, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "list_dir " + path

	result, err := sfport.SendDirectiveContext(ctx, directive, opts...)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			fssys := &mkpgo.FileSystem{}
			err = parser.UnmarshalTo(parsedResult, fssys)
			if fssys.Error != "" {
				return nil, errors.New(fssys.Error)
			}
			return fssys, err
		}

	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

func ComposeLogDirctory(logDir string) string {
	if !strings.HasPrefix(logDir, "/eMMC/applog") {
		return "/eMMC/applog/" + logDir
	}

	return logDir
}

func CleanDir(sfport *mkpgo.SFSerialPort, path string, opts ...mkpgo.DirectiveOption) error {
	return CleanDirContext(context.Background(), sfport, path, opts...)
}

func CleanDirContext(ctx context.Context, sfport *mkpgo.SFSerialPort, path string, opts ...mkpgo.DirectiveOption) error {
	if !strings.HasPrefix(path, "/eMMC/applog") {
		return errors.New("only can clean directory in working directory") // only can delete file within /eMMC/applog
	}

	if !sfport.SyncOuputEnabled {
		return errors.New("please enable sync mode first")
	}

	directive := "clean_dir " + path

	result, err := sfport.SendDirectiveContext(ctx, directive, opts...)

	if err != nil {
		return err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		_, err := parser.Parse(directive, result)

		if err != nil {
			return err
		}

		return nil

		// if parser.IsJSONOutput() {
		// 	fssys := &mkpgo.FileSystem{}
		// 	err = parser.UnmarshalTo(parsedResult, fssys)
		// 	if fssys.Error != "" {
		// 		return nil, errors.New(fssys.Error)
		// 	}
		// 	return fssys, err
		// }
	}

	return mkpgo.ErrDirectiveParserMissing
}

func ComposeLogFullpath(logPath string) string {
	if !strings.HasSuffix(logPath, ".log") {
		logPath += ".log"
	}

	if !strings.HasPrefix(logPath, "/eMMC/applog/") {
		return "/eMMC/applog/" + logPath
	}

	return logPath
}

// DeleteFile 指令 只能删除在/eMMC/applog下的文件(path 路径)
func DeleteFile(sfport *mkpgo.SFSerialPort, path string, opts ...mkpgo.DirectiveOption) error {
	return DeleteFileContext(context.Background(), sfport, path, opts...)
}

// DeleteFileContext 指令 只能删除在/eMMC/applog下的文件(path 路径)
func DeleteFileContext(ctx context.Context, sfport *mkpgo.SFSerialPort, path string, opts ...mkpgo.DirectiveOption) error {
	path = ComposeLogFullpath(path)

	if !strings.HasPrefix(path, "/eMMC/applog") {
		return errors.New("only can delete file in working directory") // only can delete file within /eMMC/applog
	}

	if !sfport.SyncOuputEnabled {
		return errors.New("please enable sync mode first")
	}

	directive := "delete_file " + path

	result, err := sfport.SendDirectiveContext(ctx, directive, opts...)

	if err != nil {
		return err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		_, err := parser.Parse(directive, result)

		if err != nil {
			return err
		}

		return nil

		// if parser.IsJSONOutput() {
		// 	fssys := &mkpgo.FileSystem{}
		// 	err = parser.UnmarshalTo(parsedResult, fssys)
		// 	if fssys.Error != "" {
		// 		return nil, errors.New(fssys.Error)
		// 	}
		// 	return fssys, err
		// }
	}

	return mkpgo.ErrDirectiveParserMissing
}

// Alive 指令 心跳时间戳
func Alive(sfport *mkpgo.SFSerialPort, opts ...mkpgo.DirectiveOption) (*mkpgo.Heartbeat, error) {
	return AliveContext(context.Background(), sfport, opts...)
}

// AliveContext 指令 心跳时间戳
func AliveContext(ctx context.Context, sfport *mkpgo.SFSerialPort, opts ...mkpgo.DirectiveOption) (*mkpgo.Heartbeat, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "alive"

	result, err := sfport.SendDirectiveContext(ctx, directive, opts...)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			hb := &mkpgo.Heartbeat{}
			err = parser.UnmarshalTo(parsedResult, hb)
			if err != nil {
				return nil, err
			}
			return hb, nil
		}

	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

// Atime 指令 返回 日志时长。 path可以是相对路径(.log扩展 - mkpdemo/1129f40), 也可以是绝对路径(/eMMC/applog/mkpdemo/1129f40.log)
func Atime(sfport *mkpgo.SFSerialPort, path string, opts ...mkpgo.DirectiveOption) (*mkpgo.LogLength, error) {
	return AtimeContext(context.Background(), sfport, path, opts...)
}

// AtimeContext 指令 返回日志时长
func AtimeContext(ctx context.Context, sfport *mkpgo.SFSerialPort, path string, opts ...mkpgo.DirectiveOption) (*mkpgo.LogLength, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "atime " + path

	result, err := sfport.SendDirectiveContext(ctx, directive, opts...)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			o := &mkpgo.LogLength{}
			err = parser.UnmarshalTo(parsedResult, o)
			if err != nil {
				return nil, err
			}
			return o, nil
		}

	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

// Aversion 指令 返回 版本信息。
func Aversion(sfport *mkpgo.SFSerialPort, opts ...mkpgo.DirectiveOption) (*mkpgo.MKPVersion, error) {
	return AversionContext(context.Background(), sfport, opts...)
}

// AversionContext 指令 返回版本信息。
func AversionContext(ctx context.Context, sfport *mkpgo.SFSerialPort, opts ...mkpgo.DirectiveOption) (*mkpgo.MKPVersion, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "aversion"

	result, err := sfport.SendDirectiveContext(ctx, directive, opts...)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			o := &mkpgo.MKPVersion{}
			err = parser.UnmarshalTo(parsedResult, o)
			if err != nil {
				return nil, err
			}
			return o, nil
		}

	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

// AInspect 指令 返回 日志基础信息。 path可以是相对路径(.log扩展 - mkpdemo/1129f40), 也可以是绝对路径(/eMMC/applog/mkpdemo/1129f40.log)
func AInspect(sfport *mkpgo.SFSerialPort, path string, opts ...mkpgo.DirectiveOption) (*mkpgo.LogInfo, error) {
	return AInspectContext(context.Background(), sfport, path, opts...)
}

// AInspectContext 指令 返回日志基础信息。
func AInspectContext(ctx context.Context, sfport *mkpgo.SFSerialPort, path string, opts ...mkpgo.DirectiveOption) (*mkpgo.LogInfo, error) {
	if !sfport.SyncOuputEnabled {
		return nil, errors.New("please enable sync mode first")
	}

	directive := "ainsp " + path

	result, err := sfport.SendDirectiveContext(ctx, directive, opts...)

	if err != nil {
		return nil, err
	}

	if parser := sfport.GetRawDirectiveOutputParser(directive); parser != nil {
		parsedResult, err := parser.Parse(directive, result)

		if err != nil {
			return nil, err
		}

		if parser.IsJSONOutput() {
			o := &mkpgo.LogInfo{}
			err = parser.UnmarshalTo(parsedResult, o)
			if err != nil {
				return nil, err
			}
			return o, nil
		}

	}

	return nil, mkpgo.ErrDirectiveParserMissing
}

// KeyDown presses one key using optional kpad settings.
// KeyDown 使用可选的 kpad 配置按下单个按键。
func KeyDown(sfport *mkpgo.SFSerialPort, key string, opts ...*mkpgo.KpadOption) error {
	return KeyDownContext(context.Background(), sfport, key, opts...)
}

// KeyDownContext presses one key using optional kpad settings and context.
// KeyDownContext 使用可选的 kpad 配置和 context 按下单个按键。
func KeyDownContext(ctx context.Context, sfport *mkpgo.SFSerialPort, key string, opts ...*mkpgo.KpadOption) error {
	return sendKeyDownContext(ctx, sfport, resolveKpadOption(firstKpadOption(opts...)), key)
}

// KeyUp releases one key using optional kpad settings.
// KeyUp 使用可选的 kpad 配置释放单个按键。
func KeyUp(sfport *mkpgo.SFSerialPort, key string, opts ...*mkpgo.KpadOption) error {
	return KeyUpContext(context.Background(), sfport, key, opts...)
}

// KeyUpContext releases one key using optional kpad settings and context.
// KeyUpContext 使用可选的 kpad 配置和 context 释放单个按键。
func KeyUpContext(ctx context.Context, sfport *mkpgo.SFSerialPort, key string, opts ...*mkpgo.KpadOption) error {
	return sendKeyUpContext(ctx, sfport, resolveKpadOption(firstKpadOption(opts...)), key)
}

// KeyTap performs a press-then-release for one key using optional kpad settings.
// KeyTap 使用可选的 kpad 配置对单个按键执行按下再释放。
func KeyTap(sfport *mkpgo.SFSerialPort, key string, opts ...*mkpgo.KpadOption) error {
	return KeyTapContext(context.Background(), sfport, key, opts...)
}

// KeyTapContext performs a press-then-release for one key using optional kpad settings and context.
// KeyTapContext 使用可选的 kpad 配置和 context 对单个按键执行按下再释放。
func KeyTapContext(ctx context.Context, sfport *mkpgo.SFSerialPort, key string, opts ...*mkpgo.KpadOption) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	opt := resolveKpadOption(firstKpadOption(opts...))
	sleep := rand.Intn(100) + 20
	if err := sendKeyDownContext(ctx, sfport, opt, key); err != nil {
		return err
	}

	timer := time.NewTimer(time.Duration(sleep) * time.Millisecond)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
	}

	if err := sendKeyUpContext(ctx, sfport, opt, key); err != nil {
		return err
	}

	return nil
}

// KeyPresses taps keys sequentially using optional kpad settings.
// KeyPresses 使用可选的 kpad 配置依次点击多个按键。
func KeyPresses(sfport *mkpgo.SFSerialPort, keys []string, sleep int, opts ...*mkpgo.KpadOption) error {
	return KeyPressesContext(context.Background(), sfport, keys, sleep, opts...)
}

// KeyPressesContext taps keys sequentially using optional kpad settings and context.
// KeyPressesContext 使用可选的 kpad 配置和 context 依次点击多个按键。
func KeyPressesContext(ctx context.Context, sfport *mkpgo.SFSerialPort, keys []string, sleep int, opts ...*mkpgo.KpadOption) error {
	opt := resolveKpadOption(firstKpadOption(opts...))
	for _, key := range keys {
		if err := KeyTapContext(ctx, sfport, key, opt); err != nil {
			return err
		}
	}
	return nil
}

// firstKpadOption returns the first provided kpad option, if any.
// firstKpadOption 返回传入的第一个 kpad 配置（若存在）。
func firstKpadOption(opts ...*mkpgo.KpadOption) *mkpgo.KpadOption {
	if len(opts) == 0 {
		return nil
	}
	return opts[0]
}

// resolveKpadOption returns a usable kpad option, filling in the default when nil.
// resolveKpadOption 返回可用的 kpad 配置；若为 nil 则补默认值。
func resolveKpadOption(opt *mkpgo.KpadOption) *mkpgo.KpadOption {
	if opt != nil {
		return opt
	}
	return mkpgo.NewKpadOption().WithDelay(0)
}

// cloneKpadOption makes a shallow copy plus slice copy for mutable kpad fields.
// cloneKpadOption 复制 kpad 配置，并拷贝可变切片字段。
func cloneKpadOption(opt *mkpgo.KpadOption) *mkpgo.KpadOption {
	if opt == nil {
		return nil
	}

	cloned := *opt
	cloned.ModKeys = append(mkpgo.KpadModKeys(nil), opt.ModKeys...)
	return &cloned
}

// resolveKpadReleaseOption clones a preset release option and applies runtime overrides.
// resolveKpadReleaseOption 复制预置释放配置，并合并运行时覆盖项。
func resolveKpadReleaseOption(base *mkpgo.KpadOption, override *mkpgo.KpadOption) *mkpgo.KpadOption {
	opt := cloneKpadOption(base)
	if opt == nil {
		opt = mkpgo.NewKpadOption().WithDelay(0)
	}
	if override != nil {
		opt.Async = override.Async
		opt.SyncIgnoreOutput = override.SyncIgnoreOutput
		opt.Verbose = override.Verbose
	}
	return opt
}

// firstM10Option returns the first provided m10 option, if any.
// firstM10Option 返回传入的第一个 m10 配置（若存在）。
func firstM10Option(opts ...*mkpgo.M10Option) *mkpgo.M10Option {
	if len(opts) == 0 {
		return nil
	}
	return opts[0]
}

// cloneM10Option makes a shallow copy of one m10 option.
// cloneM10Option 浅拷贝一个 m10 配置。
func cloneM10Option(opt *mkpgo.M10Option) *mkpgo.M10Option {
	if opt == nil {
		return nil
	}

	cloned := *opt
	return &cloned
}

// resolveMouseReleaseOption builds a release-all m10 option while preserving overrides such as async.
// resolveMouseReleaseOption 构建鼠标全释放的 m10 配置，同时保留 async 等覆盖项。
func resolveMouseReleaseOption(override *mkpgo.M10Option) *mkpgo.M10Option {
	opt := mkpgo.NewM10Option()
	if override != nil {
		opt = cloneM10Option(override)
		opt.Reset()
	}
	opt.SetButton(0)
	return opt
}

// sendKeyDown sends a prepared key-down sequence.
// sendKeyDown 发送准备好的按下按键序列。
func sendKeyDown(sfport *mkpgo.SFSerialPort, opt *mkpgo.KpadOption, key string) error {
	return sendKeyDownContext(context.Background(), sfport, opt, key)
}

// sendKeyDownContext sends a prepared key-down sequence with context.
// sendKeyDownContext 使用 context 发送准备好的按下按键序列。
func sendKeyDownContext(ctx context.Context, sfport *mkpgo.SFSerialPort, opt *mkpgo.KpadOption, key string) error {
	if strings.TrimSpace(key) == "" {
		return nil
	}

	downOpt := opt.KeyDown(key)
	return sfport.KeypadContext(ctx, downOpt)
}

// sendKeyUp sends a prepared key-up sequence.
// sendKeyUp 发送准备好的释放按键序列。
func sendKeyUp(sfport *mkpgo.SFSerialPort, opt *mkpgo.KpadOption, key string) error {
	return sendKeyUpContext(context.Background(), sfport, opt, key)
}

// sendKeyUpContext sends a prepared key-up sequence with context.
// sendKeyUpContext 使用 context 发送准备好的释放按键序列。
func sendKeyUpContext(ctx context.Context, sfport *mkpgo.SFSerialPort, opt *mkpgo.KpadOption, key string) error {
	releaseOpt, remainHoldOpt := opt.KeyUp(key)
	if releaseOpt != nil {
		if err := sfport.KeypadContext(ctx, releaseOpt); err != nil {
			return err
		}
	}
	if remainHoldOpt != nil {
		if err := sfport.KeypadContext(ctx, remainHoldOpt); err != nil {
			return err
		}
	}

	return nil
}

// KeypadRelease releases current keyboard slots to NONE using optional kpad settings.
// KeypadRelease 使用可选的 kpad 配置将当前键位槽释放为 NONE。
func KeypadRelease(sfport *mkpgo.SFSerialPort, opts ...*mkpgo.KpadOption) error {
	return KeypadReleaseContext(context.Background(), sfport, opts...)
}

// KeypadReleaseContext releases current keyboard slots to NONE using optional kpad settings and context.
// KeypadReleaseContext 使用可选的 kpad 配置和 context 将当前键位槽释放为 NONE。
func KeypadReleaseContext(ctx context.Context, sfport *mkpgo.SFSerialPort, opts ...*mkpgo.KpadOption) error {
	return sfport.KeypadContext(ctx, resolveKpadReleaseOption(mkpgo.HidKpadRelease, firstKpadOption(opts...)))
}

// KeypadReleaseAll sends a full release packet using optional kpad settings.
// KeypadReleaseAll 使用可选的 kpad 配置发送键盘全释放包。
func KeypadReleaseAll(sfport *mkpgo.SFSerialPort, opts ...*mkpgo.KpadOption) error {
	return KeypadReleaseAllContext(context.Background(), sfport, opts...)
}

// KeypadReleaseAllContext sends a full release packet using optional kpad settings and context.
// KeypadReleaseAllContext 使用可选的 kpad 配置和 context 发送键盘全释放包。
func KeypadReleaseAllContext(ctx context.Context, sfport *mkpgo.SFSerialPort, opts ...*mkpgo.KpadOption) error {
	if err := sfport.KeypadContext(ctx, resolveKpadReleaseOption(mkpgo.HidKpadReleaseAll, firstKpadOption(opts...))); err != nil {
		return err
	}
	mkpgo.ResetKpadPressedCaches()
	return nil
}

// MouseReleaseAll releases all mouse buttons using optional m10 settings.
// MouseReleaseAll 使用可选的 m10 配置释放全部鼠标按键。
func MouseReleaseAll(sfport *mkpgo.SFSerialPort, opts ...*mkpgo.M10Option) error {
	return MouseReleaseAllContext(context.Background(), sfport, opts...)
}

// MouseReleaseAllContext releases all mouse buttons using optional m10 settings and context.
// MouseReleaseAllContext 使用可选的 m10 配置和 context 释放全部鼠标按键。
func MouseReleaseAllContext(ctx context.Context, sfport *mkpgo.SFSerialPort, opts ...*mkpgo.M10Option) error {
	return sfport.Mouse10Context(ctx, resolveMouseReleaseOption(firstM10Option(opts...)))
}

// M10 sends one m10 directive with context.
// M10 使用 context 发送一条 m10 指令。
func M10(ctx context.Context, sfport *mkpgo.SFSerialPort, opt *mkpgo.M10Option) error {
	return sfport.Mouse10Context(ctx, opt)
}
