package usrbin

import (
	"time"

	"github.com/pkg/errors"
)

type VersionInfo struct {
	Version    string    `json:"version"`
	ReleasedAt time.Time `json:"releasedAt"`
}

type UpdateInfo struct {
	LatestVersion   string    `json:"latestVersion"`
	LatestReleaseAt time.Time `json:"latestReleaseAt"`

	CheckedAt *time.Time `json:"checkedAt"`

	CanUpgradeInPlace bool `json:"canUpgradeInPlace"`
}

// GetUpdateInfo will return the latest version
func (s SDK) GetUpdateInfo() (*UpdateInfo, error) {
	checkedAt := time.Now()

	latestVersion, err := s.updateChecker.GetLatestVersion()
	if err != nil {
		return nil, errors.Wrap(err, "get latest version")
	}

	if latestVersion == nil {
		return nil, nil
	}

	updateInfo, err := updateInfoFromVersions(s.version, latestVersion)
	if err != nil {
		return nil, errors.Wrap(err, "update info from versions")
	}

	updateInfo.CheckedAt = &checkedAt

	return updateInfo, nil
}

func (s SDK) DownloadUpdate() (string, error) {
	versionInfo, err := s.updateChecker.GetLatestVersion()
	if err != nil {
		return "", errors.Wrap(err, "get latest version")
	}

	if versionInfo == nil {
		return "", errors.New("no update info")
	}

	tmpFile, err := s.updateChecker.DownloadVersion(versionInfo.Version)
	if err != nil {
		return "", errors.Wrap(err, "download version")
	}

	return tmpFile, nil
}
