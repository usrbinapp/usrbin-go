package updatechecker

import (
	"time"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

type VersionInfo struct {
	Version    string     `json:"version"`
	ReleasedAt *time.Time `json:"releasedAt"`
}

type UpdateInfo struct {
	LatestVersion   string     `json:"latestVersion"`
	LatestReleaseAt *time.Time `json:"latestReleaseAt"`

	CheckedAt *time.Time `json:"checkedAt"`

	CanUpgradeInPlace      bool   `json:"canUpgradeInPlace"`
	ExternalUpgradeCommand string `json:"externalUpgradeCommand"`
}

type UpdateChecker interface {
	GetLatestVersion() (*VersionInfo, error)
	DownloadVersion(version string, requireChecksumMatch bool) (string, error)
}

func UpdateInfoFromVersions(currentVersion string, latestVersion *VersionInfo) (*UpdateInfo, error) {
	if latestVersion == nil {
		return nil, errors.New("latest version is nil")
	}

	latestSemver, err := semver.NewVersion(latestVersion.Version)
	if err != nil {
		return nil, errors.Wrap(err, "latest semver")
	}

	currentSemver, err := semver.NewVersion(currentVersion)
	if err != nil {
		return nil, errors.Wrap(err, "current semver")
	}

	if latestSemver.LessThan(currentSemver) || latestSemver.Equal(currentSemver) {
		return nil, nil
	}

	updateInfo := UpdateInfo{
		LatestVersion:   latestVersion.Version,
		LatestReleaseAt: latestVersion.ReleasedAt,
	}

	return &updateInfo, nil
}
