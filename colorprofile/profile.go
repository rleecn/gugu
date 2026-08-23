// Package colorprofile 检测当前终端的颜色能力，并提供颜色降级映射。
//
// 设计目标：让 gugu 应用在低端终端（ANSI 16 色、无色）上也能正确显示，
// 而非输出乱码 escape 序列。Program 通过 WithColorProfile 注入检测结果，
// 用户也可手动调用 Detect() 获取当前终端的 profile。
package colorprofile

import (
	"os"
	"runtime"
	"strings"
)

// Profile 表示终端支持的颜色能力。与 program.ColorProfile 一致，
// 单独定义避免 program 与 colorprofile 之间形成循环依赖。
type Profile int

const (
	// Ascii 无颜色（仅黑白）。
	Ascii Profile = iota
	// ANSI 16 色。
	ANSI
	// ANSI256 256 色。
	ANSI256
	// TrueColor 24-bit RGB。
	TrueColor
)

// Detect 根据环境变量自动检测终端颜色能力。
// 检测顺序（优先级从高到低）：
//  1. NO_COLOR set                              -> Ascii
//  2. COLORTERM=truecolor / 24bit               -> TrueColor
//  3. Windows 平台分支（见 detectWindows）       -> Windows 终端大多不设 TERM
//  4. TERM=*-256color                           -> ANSI256
//  5. TERM=ansi|vt100|dumb 或 TERM 为空          -> Ascii
//  6. TERM 包含 color 字样                       -> ANSI
//  7. 默认                                       -> ANSI256（现代终端保守假设）
//
// 检测结果可作为 program.WithColorProfile 入参。
func Detect() Profile {
	// NO_COLOR 协议：https://no-color.org/
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return Ascii
	}
	if ct := os.Getenv("COLORTERM"); ct != "" {
		lc := strings.ToLower(ct)
		if lc == "truecolor" || lc == "24bit" || strings.Contains(lc, "24-bit") {
			return TrueColor
		}
	}
	// Windows 分支必须在 TERM 判空之前：conhost/PowerShell/Windows Terminal
	// 默认都不设置 TERM，通用规则会把整个 Windows 误判为 Ascii（无色）。
	if runtime.GOOS == "windows" {
		return detectWindows()
	}
	term := strings.ToLower(os.Getenv("TERM"))
	if term == "" || term == "dumb" {
		return Ascii
	}
	if strings.HasSuffix(term, "-256color") || strings.Contains(term, "256color") {
		return ANSI256
	}
	switch term {
	case "ansi", "vt100", "vt220":
		return ANSI
	}
	if strings.Contains(term, "color") || strings.Contains(term, "ansi") {
		return ANSI
	}
	// 现代终端默认假设 256 色（覆盖 macOS Terminal.app、gnome-terminal 等）
	return ANSI256
}

// detectWindows 检测 Windows 终端颜色能力。
// 优先级：WT_SESSION（Windows Terminal，TrueColor）> ConEmu > ANSICON >
// TERM（Git Bash/MSYS 会设置，复用通用规则）> 默认 ANSI
// （Win10 1607+ conhost 支持 VT 序列，保守按 16 色处理以兼容老系统）。
func detectWindows() Profile {
	if os.Getenv("WT_SESSION") != "" {
		return TrueColor
	}
	if os.Getenv("ConEmuANSI") != "" {
		return ANSI256
	}
	if os.Getenv("ANSICON") != "" {
		return ANSI
	}
	if term := strings.ToLower(os.Getenv("TERM")); term != "" && term != "dumb" {
		if strings.Contains(term, "256color") {
			return ANSI256
		}
		if strings.Contains(term, "color") || strings.Contains(term, "ansi") {
			return ANSI
		}
	}
	return ANSI
}

// Downgrade 把给定 Profile 降级到 target 及以下。
// 若当前 profile 已低于 target，返回自身。
func (p Profile) Downgrade(target Profile) Profile {
	if p > target {
		return target
	}
	return p
}

// String 返回可读名称。
func (p Profile) String() string {
	switch p {
	case Ascii:
		return "Ascii"
	case ANSI:
		return "ANSI"
	case ANSI256:
		return "ANSI256"
	case TrueColor:
		return "TrueColor"
	}
	return "Unknown"
}

// Supports 表示该 profile 是否支持某种颜色能力。
func (p Profile) Supports(other Profile) bool { return p >= other }
