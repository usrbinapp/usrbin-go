package usrbin

import (
	"time"

	"github.com/pkg/errors"
	"github.com/usrbinapp/usrbin-go/pkg/updatechecker"
)

// GetUpdateInfo will return the latest version
func (s SDK) GetUpdateInfo() (*updatechecker.UpdateInfo, error) {
	checkedAt := time.Now()

	latestVersion, err := s.updateChecker.GetLatestVersion(s.httpTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "get latest version")
	}

	if latestVersion == nil {
		return nil, nil
	}

	updateInfo, err := updatechecker.UpdateInfoFromVersions(s.version, latestVersion)
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
