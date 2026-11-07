package search

import (
	"testing"
)

func TestWindowsOpenUsesRundll32(t *testing.T) {
	args := buildOpenArgs("windows", "https://example.com")
	if args == nil {
		t.Fatal("expected non-nil args for windows")
	}
	if args[0] != "rundll32" || args[1] != "url.dll,FileProtocolHandler" {
		t.Errorf("windows should open via rundll32 url.dll,FileProtocolHandler, got %v", args)
	}
}

// exec.Command 不经过 shell, 含 cmd 元字符的网址必须原样传递, 否则 windows 下存在命令注入风险
func TestOpenArgsPassUrlVerbatim(t *testing.T) {
	url := `https://example.com/?q=1&x=|whoami|^&pause`
	for _, goos := range []string{"windows", "darwin", "linux"} {
		args := buildOpenArgs(goos, url)
		if args == nil {
			t.Fatalf("expected non-nil args for %s", goos)
		}
		if got := args[len(args)-1]; got != url {
			t.Errorf("%s: url was modified: %q", goos, got)
		}
	}
}

func TestBuildOpenArgsUnknownGoos(t *testing.T) {
	if buildOpenArgs("plan9", "x") != nil {
		t.Error("unknown goos should return nil")
	}
}
