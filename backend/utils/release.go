package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"IndicoDataFusion/backend/consts"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
)

// ReleaseAsset represents a downloadable asset attached to a GitHub release
type ReleaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int    `json:"size"`
}

// ReleaseCheckResult is returned to the frontend when checking for a new release
type ReleaseCheckResult struct {
	LatestTag string         `json:"latestTag"`
	IsNew     bool           `json:"isNew"`
	Assets    []ReleaseAsset `json:"assets"`
	HTMLURL   string         `json:"htmlURL"`
	Body      string         `json:"body"`
	Repo      string         `json:"repo"`
}

// CheckLatestRelease queries the given GitHub repository URL for the latest release
// and returns structured information about it. currentVersion should be the
// current application version (e.g., from consts.AppVersion) and is used to set
// the IsNew flag. Comparison prefers semantic version ordering when both
// versions parse as semver; otherwise it falls back to simple string inequality.
func CheckLatestRelease(repoURL string, currentVersion string) (*ReleaseCheckResult, error) {
	if repoURL == "" {
		return nil, errors.New("repository URL not provided")
	}

	u, err := url.Parse(repoURL)
	if err != nil {
		return nil, errors.Wrap(err, "invalid repo url")
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 {
		return nil, errors.Errorf("invalid repo url (expected github.com/owner/repo): %s", repoURL)
	}
	owner := parts[0]
	repoName := parts[1]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repoName)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create request")
	}
	// Provide a user-agent to avoid generic API rejections
	req.Header.Set("User-Agent", consts.AppName)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query GitHub API")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, errors.Errorf("GitHub API returned %d: %s", resp.StatusCode, string(b))
	}

	var payload struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
		Body    string `json:"body"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int    `json:"size"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, errors.Wrap(err, "failed to decode GitHub response")
	}

	latestRaw := strings.TrimPrefix(payload.TagName, "v")
	currentRaw := strings.TrimPrefix(currentVersion, "v")
	isNew := false

	// Prefer semver-aware comparison when both parse as semantic versions
	if latestRaw != "" {
		if lv, lerr := semver.NewVersion(latestRaw); lerr == nil {
			if cv, cerr := semver.NewVersion(currentRaw); cerr == nil {
				if lv.GreaterThan(cv) {
					isNew = true
				}
			} else {
				// If current doesn't parse but latest does, treat as new
				isNew = true
			}
		} else {
			// Fallback: simple string inequality
			if latestRaw != currentRaw {
				isNew = true
			}
		}
	}

	assets := make([]ReleaseAsset, 0, len(payload.Assets))
	for _, aasset := range payload.Assets {
		assets = append(assets, ReleaseAsset{
			Name: aasset.Name,
			URL:  aasset.BrowserDownloadURL,
			Size: aasset.Size,
		})
	}

	res := &ReleaseCheckResult{
		LatestTag: payload.TagName,
		IsNew:     isNew,
		Assets:    assets,
		HTMLURL:   payload.HTMLURL,
		Body:      payload.Body,
		Repo:      fmt.Sprintf("%s/%s", owner, repoName),
	}

	return res, nil
}
