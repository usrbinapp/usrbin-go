package oci

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
	"github.com/usrbinapp/usrbin-go/pkg/updatechecker"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry"
	"oras.land/oras-go/v2/registry/remote"
)

type OCIUpdateChecker struct {
	artifact string
}

var _ updatechecker.UpdateChecker = (*OCIUpdateChecker)(nil)

func NewOCIUpdateChecker(artifact string) updatechecker.UpdateChecker {
	return &OCIUpdateChecker{
		artifact: strings.TrimRight(artifact, ":"),
	}
}

// DownloadVersion will download and extract the specific version, returning
// a path to the extracted file in the archive
// it's the responsibility of the caller to clean up the extracted file
func (c OCIUpdateChecker) DownloadVersion(version string, requireChecksumMatch bool) (string, error) {
	ref := fmt.Sprintf("%s:%s", c.artifact, version)

	tmpDir, err := ioutil.TempDir("", "usrbin")
	if err != nil {
		return "", errors.Wrap(err, "create temp dir")
	}
	// defer os.RemoveAll(tmpDir)

	// Pull file(s) from registry and save to disk
	fileStore, err := file.New(tmpDir)
	if err != nil {
		return "", errors.Wrap(err, "create file store")
	}

	defer fileStore.Close()

	src, err := remote.NewRepository(c.artifact)
	if err != nil {
		return "", errors.Wrap(err, "create remote repository")
	}

	copyOpts := oras.DefaultCopyOptions
	copyOpts.Concurrency = 1

	localTarget := oras.Target(fileStore)
	_, err = oras.Copy(context.Background(), src, ref, localTarget, ref, copyOpts)
	if err != nil {
		return "", errors.Wrap(err, "copy from remote")
	}

	path, err := bestAsset(tmpDir)
	if err != nil {
		return "", errors.Wrap(err, "get best asset")
	}

	if path == "" {
		return "", errors.New("no assets found")
	}

	// make the file executable
	err = os.Chmod(path, 0755)
	if err != nil {
		return "", errors.Wrap(err, "chmod")
	}

	// copy the best asset to a temp file
	tmpFile, err := ioutil.TempFile("", "usrbin")
	if err != nil {
		return "", errors.Wrap(err, "create temp file")
	}
	defer tmpFile.Close()

	// copy the file
	asset, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "open asset")
	}

	_, err = io.Copy(tmpFile, asset)
	if err != nil {
		return "", errors.Wrap(err, "copy asset")
	}

	return tmpFile.Name(), nil

}

// GetLatestVersion will return the latest version information from the oci repository
func (c OCIUpdateChecker) GetLatestVersion(timeout time.Duration) (*updatechecker.VersionInfo, error) {
	repo, err := remote.NewRepository(c.artifact)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	tags, err := registry.Tags(ctx, repo)
	if err != nil {
		return nil, err
	}

	var latestSemver *semver.Version
	var latestUnparsed string
	for _, tag := range tags {
		parsed, err := semver.NewVersion(tag)
		if err != nil {
			continue
		}

		if latestSemver == nil || parsed.GreaterThan(latestSemver) {
			latestSemver = parsed
			latestUnparsed = tag
		}
	}

	latestVersion := &updatechecker.VersionInfo{
		Version:    latestUnparsed,
		ReleasedAt: nil,
	}

	return latestVersion, nil
}

func bestAsset(inPath string) (string, error) {
	bestAssetPath := ""
	if err := filepath.Walk(inPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if bestAssetPath != "" {
			return nil
		}

		bestAssetPath = path
		return nil
	}); err != nil {
		return "", err
	}

	return bestAssetPath, nil
}
