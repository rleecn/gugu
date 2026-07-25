package colorprofile

import (
	"os"
	"testing"
)

// 测试需要完全控制的环境变量集合：所有影响 Detect 的环境变量都被纳入。
var detectEnvKeys = []string{"NO_COLOR", "COLORTERM", "TERM"}

func setEnv(t *testing.T, kvs map[string]string) {
	t.Helper()
	// 备份原值
	prev := make(map[string]string)
	for _, k := range detectEnvKeys {
		if v, ok := os.LookupEnv(k); ok {
			prev[k] = v
		}
	}
	// 先全部清空
	for _, k := range detectEnvKeys {
		_ = os.Unsetenv(k)
	}
	// 再设置指定值
	for k, v := range kvs {
		if v == "" {
			_ = os.Unsetenv(k)
		} else {
			_ = os.Setenv(k, v)
		}
	}
	t.Cleanup(func() {
		// 先全部清空
		for _, k := range detectEnvKeys {
			_ = os.Unsetenv(k)
		}
		// 再恢复
		for k, v := range prev {
			_ = os.Setenv(k, v)
		}
	})
}

func TestDetectTrueColor(t *testing.T) {
	setEnv(t, map[string]string{
		"NO_COLOR":  "",
		"COLORTERM": "truecolor",
		"TERM":      "xterm-256color",
	})
	if got := Detect(); got != TrueColor {
		t.Fatalf("expected TrueColor, got %v", got)
	}
}

func TestDetect24bit(t *testing.T) {
	setEnv(t, map[string]string{
		"COLORTERM": "24bit",
		"TERM":      "xterm-256color",
	})
	if got := Detect(); got != TrueColor {
		t.Fatalf("expected TrueColor, got %v", got)
	}
}

func TestDetect256Color(t *testing.T) {
	setEnv(t, map[string]string{
		"COLORTERM": "",
		"TERM":      "xterm-256color",
	})
	if got := Detect(); got != ANSI256 {
		t.Fatalf("expected ANSI256, got %v", got)
	}
}

func TestDetectNoColor(t *testing.T) {
	setEnv(t, map[string]string{
		"NO_COLOR":  "1",
		"COLORTERM": "truecolor",
		"TERM":      "xterm-256color",
	})
	if got := Detect(); got != Ascii {
		t.Fatalf("NO_COLOR should force Ascii, got %v", got)
	}
}

func TestDetectDumbTerm(t *testing.T) {
	setEnv(t, map[string]string{
		"TERM": "dumb",
	})
	if got := Detect(); got != Ascii {
		t.Fatalf("TERM=dumb should be Ascii, got %v", got)
	}
}

func TestDetectEmptyTerm(t *testing.T) {
	setEnv(t, map[string]string{
		"TERM": "",
	})
	if got := Detect(); got != Ascii {
		t.Fatalf("empty TERM should be Ascii, got %v", got)
	}
}

func TestDetectAnsiTerm(t *testing.T) {
	setEnv(t, map[string]string{
		"TERM": "ansi",
	})
	if got := Detect(); got != ANSI {
		t.Fatalf("TERM=ansi should be ANSI, got %v", got)
	}
}

func TestDowngrade(t *testing.T) {
	if TrueColor.Downgrade(ANSI) != ANSI {
		t.Fatal("TrueColor.Downgrade(ANSI) should be ANSI")
	}
	if ANSI.Downgrade(TrueColor) != ANSI {
		t.Fatal("ANSI.Downgrade(TrueColor) should keep ANSI")
	}
}

func TestSupports(t *testing.T) {
	if !TrueColor.Supports(ANSI256) {
		t.Fatal("TrueColor should support ANSI256")
	}
	if ANSI.Supports(TrueColor) {
		t.Fatal("ANSI should NOT support TrueColor")
	}
}

func TestProfileString(t *testing.T) {
	cases := []struct {
		p    Profile
		want string
	}{
		{Ascii, "Ascii"},
		{ANSI, "ANSI"},
		{ANSI256, "ANSI256"},
		{TrueColor, "TrueColor"},
	}
	for _, c := range cases {
		if got := c.p.String(); got != c.want {
			t.Fatalf("Profile.String() = %q, want %q", got, c.want)
		}
	}
}
