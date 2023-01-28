package usrbin

type UpdateChecker interface {
	GetLatestVersion(currentVersion string) (*UpdateInfo, error)
}

var _ UpdateChecker = (*GitHubUpdateChecker)(nil)
