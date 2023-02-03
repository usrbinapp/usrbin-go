package usrbin

import (
	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

type UpdateChecker interface {
	GetLatestVersion() (*VersionInfo, error)
	DownloadVersion(version string, requireChecksumMatch bool) (string, error)
}

var _ UpdateChecker = (*GitHubUpdateChecker)(nil)

func updateInfoFromVersions(currentVersion string, latestVersion *VersionInfo) (*UpdateInfo, error) {
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
