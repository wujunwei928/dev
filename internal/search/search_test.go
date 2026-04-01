package search

import (
	"strings"
	"testing"
)

func TestEngineParamMap_HasAllEngines(t *testing.T) {
	engines := []string{
		EngineBing, EngineBaidu, EngineGoogle,
		EngineZhiHu, EngineWeiXin, EngineGithub,
		EngineKaiFa, EngineDouBan, EngineMovie,
		EngineBook, Engine360, EngineSoGou,
	}
	for _, eng := range engines {
		param, ok := EngineParamMap[eng]
		if !ok {
			t.Errorf("engine %q not found in EngineParamMap", eng)
			continue
		}
		if param.Domain == "" {
			t.Errorf("engine %q has empty Domain", eng)
		}
	}
}

func TestFormatSearchUrl_Bing(t *testing.T) {
	u := FormatSearchUrl(EngineBing, "golang")
	if u == "" {
		t.Fatal("expected non-empty URL")
	}
	if !strings.Contains(u, "bing.com") {
		t.Errorf("expected bing.com in URL, got %s", u)
	}
}

func TestFormatSearchUrl_UnknownEngine(t *testing.T) {
	u := FormatSearchUrl("nonexistent", "test")
	if u == "" {
		t.Fatal("expected fallback to bing URL")
	}
	if !strings.Contains(u, "bing.com") {
		t.Errorf("expected bing.com fallback URL, got %s", u)
	}
}

func TestGetEngineParam(t *testing.T) {
	param := getEngineParam(EngineBing)
	if param.Domain == "" {
		t.Error("expected non-empty Domain")
	}

	// 未知引擎应回退到 bing
	fallback := getEngineParam("nonexistent")
	if fallback.Domain != param.Domain {
		t.Error("unknown engine should fallback to bing")
	}
}

func TestFormatSearchCommandModeUsage(t *testing.T) {
	usage := FormatSearchCommandModeUsage()
	if usage == "" {
		t.Fatal("expected non-empty usage string")
	}
	if !strings.Contains(usage, "bing") {
		t.Error("expected 'bing' in usage")
	}
}
