package usrbin

type UpdateChecker interface {
	GetLatestVersion(currentVersion string) (*UpdateInfo, error)
	DownloadVersion(version string) (string, error)
}

var _ UpdateChecker = (*GitHubUpdateChecker)(nil)
