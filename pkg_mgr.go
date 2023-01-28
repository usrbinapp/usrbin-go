package usrbin

type ExternalPackageManager interface {
	IsInstalled() (bool, error)
	UpgradeCommand() string
}

var _ ExternalPackageManager = (*HomebrewExternalPackageManager)(nil)
