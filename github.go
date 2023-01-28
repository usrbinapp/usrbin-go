package usrbin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type GitHubUpdateChecker struct {
	repo string

	host       string
	parsedRepo struct {
		owner string
		repo  string
	}
}

type gitHubReleaseInfo struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
}

var ErrReleaseNotFound = errors.New("release not found")

func NewGitHubUpdateChecker(repo string) UpdateChecker {
	return GitHubUpdateChecker{
		repo: repo,
		host: "https://api.github.com",
		parsedRepo: struct {
			owner string
			repo  string
		}{
			owner: "",
			repo:  "",
		},
	}
}

func (c GitHubUpdateChecker) GetLatestVersion(currentVersion string) (*UpdateInfo, error) {
	checkedAt := time.Now()

	latestReleaseInfo, err := getReleaseDetails(c.host, c.parsedRepo.owner, c.parsedRepo.repo, "latest")
	if err != nil {
		return nil, errors.Wrap(err, "get release details")
	}

	currentReleaseInfo, err := getReleaseDetails(c.host, c.parsedRepo.owner, c.parsedRepo.repo, currentVersion)
	if err != nil && err != ErrReleaseNotFound {
		return nil, errors.Wrap(err, "get release details")
	}

	updateInfo := UpdateInfo{
		LatestVersion:   latestReleaseInfo.TagName,
		LatestReleaseAt: &latestReleaseInfo.PublishedAt,
		CheckedAt:       &checkedAt,
	}

	if currentReleaseInfo != nil {
		updateInfo.VersionsBehind = nil
		updateInfo.AbsoluteVersionAgeDays = nil
	}

	return &updateInfo, nil
}

func getReleaseDetails(host string, owner string, repo string, releaseName string) (*gitHubReleaseInfo, error) {
	uri := ""

	if releaseName == "latest" {
		uri = fmt.Sprintf("%s/repos/%s/%s/releases/latest", host, owner, repo)
	} else {
		uri = fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", host, owner, repo, releaseName)
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, "new request")
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrReleaseNotFound
		}

		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	releaseInfo := gitHubReleaseInfo{}

	if err := json.NewDecoder(resp.Body).Decode(&releaseInfo); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	return &releaseInfo, nil
}
