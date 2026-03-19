package update

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const releaseURL = "https://api.github.com/repos/gdiab/toggl-tui/releases/latest"

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckLatest returns the latest release tag, or empty string on any error.
func CheckLatest() string {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(releaseURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return ""
	}
	var rel githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return ""
	}
	return rel.TagName
}

// IsNewer returns true if remote is a higher semver than current.
// Both should be like "v0.1.1" or "0.1.1".
func IsNewer(current, remote string) bool {
	cur := parseSemver(current)
	rem := parseSemver(remote)
	if cur == nil || rem == nil {
		return false
	}
	for i := 0; i < 3; i++ {
		if rem[i] > cur[i] {
			return true
		}
		if rem[i] < cur[i] {
			return false
		}
	}
	return false
}

func parseSemver(s string) []int {
	s = strings.TrimPrefix(s, "v")
	parts := strings.SplitN(s, ".", 3)
	if len(parts) != 3 {
		return nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		// Strip any pre-release suffix (e.g. "1-rc1")
		p = strings.SplitN(p, "-", 2)[0]
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}

// FormatNotice returns the update notice string, or empty if no update.
func FormatNotice(current, latest string) string {
	if latest == "" || !IsNewer(current, latest) {
		return ""
	}
	return fmt.Sprintf("Update available: %s -> %s | go install github.com/gdiab/toggl-tui@latest", current, latest)
}
