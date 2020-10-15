package color

import (
	"runtime"
)

type Color string

//Color constants for output. Just copy pasted from internet, we can reduce options to only what's used.
const (
	Reset             Color = "\x1b[0m"
	Italics           Color = "\x1b[3m"
	Underline         Color = "\x1b[4m"
	Blink             Color = "\x1b[5m"
	Inverse           Color = "\x1b[7m"
	ItalicsOff        Color = "\x1b[23m"
	UnderlineOff      Color = "\x1b[24m"
	BlinkOff          Color = "\x1b[25m"
	InverseOff        Color = "\x1b[27m"
	Black             Color = "\x1b[30m"
	DarkGray          Color = "\x1b[30;1m"
	Red               Color = "\x1b[31m"
	LightRed          Color = "\x1b[31;1m"
	Green             Color = "\x1b[32m"
	LightGreen        Color = "\x1b[32;1m"
	Yellow            Color = "\x1b[33m"
	LightYellow       Color = "\x1b[33;1m"
	Blue              Color = "\x1b[34m"
	LightBlue         Color = "\x1b[34;1m"
	Magenta           Color = "\x1b[35m"
	LightMagenta      Color = "\x1b[35;1m"
	Cyan              Color = "\x1b[36m"
	LightCyan         Color = "\x1b[36;1m"
	Gray              Color = "\x1b[37m"
	White             Color = "\x1b[37;1m"
	ResetForeground   Color = "\x1b[39m"
	BlackBackground   Color = "\x1b[40m"
	RedBackground     Color = "\x1b[41m"
	GreenBackground   Color = "\x1b[42m"
	YellowBackground  Color = "\x1b[43m"
	BlueBackground    Color = "\x1b[44m"
	MagentaBackground Color = "\x1b[45m"
	CyanBackground    Color = "\x1b[46m"
	GrayBackground    Color = "\x1b[47m"
	ResetBackground   Color = "\x1b[49m"
	Clear             Color = ""
)

func (c Color) GetColor() Color {
	return c
}

type Colorer interface {
	GetColor() Color
}

//ColorString wraps a string in the colorization ascii
func ColorString(c Colorer, s string) string {
	if runtime.GOOS == "windows" {
		return s //If windows no color support so let's just return clean code
	}
	return string(c.GetColor()) + s + string(Reset.GetColor())
}
