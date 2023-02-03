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

	CanUpgradeInPlace      bool   `json:"canUpgradeInPlace"`
	ExternalUpgradeCommand string `json:"externalUpgradeCommand"`
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
	if updateInfo == nil {
		return nil, nil
	}

	updateInfo.ExternalUpgradeCommand = s.ExternalUpgradeCommand()
	updateInfo.CanUpgradeInPlace = updateInfo.ExternalUpgradeCommand == ""
	updateInfo.CheckedAt = &checkedAt

	return updateInfo, nil
}
