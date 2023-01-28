package usrbin

import (
	"time"

	"github.com/pkg/errors"
)

type UpdateInfo struct {
	LatestVersion   string     `json:"latestVersion"`
	LatestReleaseAt *time.Time `json:"latestReleaseAt"`

	VersionsBehind         *int `json:"versionsBehind,omitempty"`
	AbsoluteVersionAgeDays *int `json:"absoluteVersionAgeDays,omitempty"`

	CheckedAt *time.Time `json:"checkedAt"`

	CanUpgradeInPlace bool `json:"canUpgradeInPlace"`
}

// GetUpdateInfo will return the latest version
func (s SDK) GetUpdateInfo() (*UpdateInfo, error) {
	updateInfo, err := s.updateChecker.GetLatestVersion(s.version)
	if err != nil {
		return nil, errors.Wrap(err, "get latest version")
	}

	updateInfo.CanUpgradeInPlace = true

	return updateInfo, nil
}
