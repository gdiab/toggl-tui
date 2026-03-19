package update

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsNewer(t *testing.T) {
	tests := []struct {
		current, remote string
		want            bool
	}{
		{"v0.1.0", "v0.1.1", true},
		{"v0.1.1", "v0.1.1", false},
		{"v0.1.1", "v0.1.0", false},
		{"v0.1.1", "v0.2.0", true},
		{"v0.1.1", "v1.0.0", true},
		{"v1.0.0", "v0.9.9", false},
		{"0.1.0", "0.1.1", true},
		{"v0.1.0", "0.1.1", true},
		// Pre-release suffixes are stripped
		{"v1.0.0-rc1", "v1.0.0", false},
		{"v1.0.0-rc1", "v1.0.1", true},
		{"v0.9.0", "v1.0.0-beta1", true},
		// Invalid inputs
		{"bad", "v0.1.1", false},
		{"v0.1.0", "bad", false},
		{"", "", false},
		{"v", "v0.1.0", false},
		{"v1.0", "v1.0.1", false},
	}
	for _, tt := range tests {
		t.Run(tt.current+"_vs_"+tt.remote, func(t *testing.T) {
			got := IsNewer(tt.current, tt.remote)
			if got != tt.want {
				t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.remote, got, tt.want)
			}
		})
	}
}

func TestFormatNotice(t *testing.T) {
	t.Run("newer available", func(t *testing.T) {
		notice := FormatNotice("v0.1.0", "v0.2.0")
		if notice == "" {
			t.Fatal("expected non-empty notice")
		}
		if !strings.Contains(notice, "v0.1.0") {
			t.Errorf("notice should contain current version: %q", notice)
		}
		if !strings.Contains(notice, "v0.2.0") {
			t.Errorf("notice should contain latest version: %q", notice)
		}
		if !strings.Contains(notice, "go install") {
			t.Errorf("notice should contain install command: %q", notice)
		}
	})

	t.Run("already latest", func(t *testing.T) {
		notice := FormatNotice("v0.2.0", "v0.1.0")
		if notice != "" {
			t.Errorf("expected empty notice, got %q", notice)
		}
	})

	t.Run("empty latest", func(t *testing.T) {
		notice := FormatNotice("v0.1.0", "")
		if notice != "" {
			t.Errorf("expected empty notice, got %q", notice)
		}
	})

	t.Run("equal versions", func(t *testing.T) {
		notice := FormatNotice("v0.1.0", "v0.1.0")
		if notice != "" {
			t.Errorf("expected empty notice for equal versions, got %q", notice)
		}
	})
}

func TestCheckLatest(t *testing.T) {
	// Save and restore the real URL
	origURL := ReleaseURL
	defer func() { ReleaseURL = origURL }()

	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(githubRelease{TagName: "v1.2.3"})
		}))
		defer srv.Close()
		ReleaseURL = srv.URL

		got := CheckLatest()
		if got != "v1.2.3" {
			t.Errorf("CheckLatest() = %q, want 'v1.2.3'", got)
		}
	})

	t.Run("server error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(500)
		}))
		defer srv.Close()
		ReleaseURL = srv.URL

		got := CheckLatest()
		if got != "" {
			t.Errorf("CheckLatest() = %q, want empty on server error", got)
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("{garbage"))
		}))
		defer srv.Close()
		ReleaseURL = srv.URL

		got := CheckLatest()
		if got != "" {
			t.Errorf("CheckLatest() = %q, want empty on bad JSON", got)
		}
	})

	t.Run("unreachable", func(t *testing.T) {
		ReleaseURL = "http://127.0.0.1:1/nope"

		got := CheckLatest()
		if got != "" {
			t.Errorf("CheckLatest() = %q, want empty on unreachable", got)
		}
	})
}
