package oci

import (
	"github.com/usrbinapp/usrbin-go/pkg/updatechecker"
)

type OCIUpdateChecker struct {
	artifact string
}

var _ updatechecker.UpdateChecker = (*OCIUpdateChecker)(nil)

func NewOCIUpdateChecker(artifact string) updatechecker.UpdateChecker {
	return &OCIUpdateChecker{
		artifact: artifact,
	}
}

// DownloadVersion will download and extract the specific version, returning
// a path to the extracted file in the archive
// it's the responsibility of the caller to clean up the extracted file
func (c OCIUpdateChecker) DownloadVersion(version string, requireChecksumMatch bool) (string, error) {
	return "", nil
}

// GetLatestVersion will return the latest version information from the oci repository
func (c OCIUpdateChecker) GetLatestVersion() (*updatechecker.VersionInfo, error) {
	// get the latest tag from the oci repository

	return nil, nil
}
