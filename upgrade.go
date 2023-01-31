package usrbin

import (
	"os"

	"github.com/minio/selfupdate"
	"github.com/pkg/errors"
	"github.com/usrbinapp/usrbin-go/pkg/logger"
)

// CanSupportUpgrade
func (s SDK) CanSupportUpgrade() (bool, error) {
	for _, epm := range s.externalPackageManagers {
		isInstalled, err := epm.IsInstalled()
		if err != nil {
			return false, err
		}
		if isInstalled {
			return false, nil
		}
	}

	return true, nil
}

func (s SDK) ExternalUpgradeCommand() string {
	for _, epm := range s.externalPackageManagers {
		isInstalled, err := epm.IsInstalled()
		if err != nil {
			return ""
		}
		if isInstalled {
			return epm.UpgradeCommand()
		}
	}

	return ""
}

// Upgrade is the entrypoint that the app will use to perform an in-place upgrade
// we need to assume that the app is running and we are running in the main thread
func (s SDK) Upgrade() error {
	// assume the latest
	updateInfo, err := s.GetUpdateInfo()
	if err != nil {
		return errors.Wrap(err, "get update info")
	}

	if updateInfo == nil {
		return errors.New("no update info")
	}

	newVersionPath, err := s.updateChecker.DownloadVersion(updateInfo.LatestVersion, true)
	if err != nil {
		return errors.Wrap(err, "download version")
	}

	f, err := os.Open(newVersionPath)
	if err != nil {
		return errors.Wrap(err, "open new version")
	}
	defer func() {
		if err := f.Close(); err != nil {
			logger.Error(err)
		}
	}()

	err = selfupdate.Apply(f, selfupdate.Options{})
	if err != nil {
		return errors.Wrap(err, "apply update")
	}

	return nil

}
