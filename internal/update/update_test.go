package update

import "testing"

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
		{"bad", "v0.1.1", false},
		{"v0.1.0", "bad", false},
	}
	for _, tt := range tests {
		got := IsNewer(tt.current, tt.remote)
		if got != tt.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", tt.current, tt.remote, got, tt.want)
		}
	}
}

func TestFormatNotice(t *testing.T) {
	notice := FormatNotice("v0.1.0", "v0.2.0")
	if notice == "" {
		t.Error("expected non-empty notice")
	}

	notice = FormatNotice("v0.2.0", "v0.1.0")
	if notice != "" {
		t.Errorf("expected empty notice, got %q", notice)
	}

	notice = FormatNotice("v0.1.0", "")
	if notice != "" {
		t.Errorf("expected empty notice for empty latest, got %q", notice)
	}
}
