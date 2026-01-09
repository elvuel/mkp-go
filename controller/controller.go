package controller

import (
	mkpgo "github.com/elvuel/mkp-go"
	"github.com/elvuel/mkp-go/helper"
)

type Controller struct {
	sfport *mkpgo.SFSerialPort
}

func NewController(sfport *mkpgo.SFSerialPort) *Controller {
	return &Controller{
		sfport: sfport,
	}
}

func (c *Controller) Open() error {
	return c.sfport.Open()
}

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

// helper func KeyDown(sfport *mkpgo.SFSerialPort, key string) error
func (c *Controller) KeyDown(key string) error {
	return helper.KeyDown(c.sfport, key)
}

// helper func KeyUp(sfport *mkpgo.SFSerialPort, key string) error
func (c *Controller) KeyUp(key string) error {
	return helper.KeyUp(c.sfport, key)
}

// helper func KeyTap(sfport *mkpgo.SFSerialPort, keys []string) error
func (c *Controller) KeyTap(keys []string) error {
	return helper.KeyTap(c.sfport, keys)
}

// helper func KeyPress(sfport *mkpgo.SFSerialPort, key string, sleep int) error
func (c *Controller) KeyPress(key string, sleep int) error {
	return helper.KeyPress(c.sfport, key, sleep)
}

// helper func KeyPresses(sfport *mkpgo.SFSerialPort, keys []string, sleep int) error
func (c *Controller) KeyPresses(keys []string, sleep int) error {
	return helper.KeyPresses(c.sfport, keys, sleep)
}

// helper func KeypadRelease(sfport *mkpgo.SFSerialPort) error
func (c *Controller) KeypadRelease() error {
	return helper.KeypadRelease(c.sfport)
}

// helper func KeypadReleaseAll(sfport *mkpgo.SFSerialPort) error
func (c *Controller) KeypadReleaseAll() error {
	return helper.KeypadReleaseAll(c.sfport)
}
