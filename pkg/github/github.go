package github

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/usrbinapp/usrbin-go/pkg/logger"
	"github.com/usrbinapp/usrbin-go/pkg/updatechecker"
)

var (
	ErrUnknownArchiveType        = errors.New("unknown archive type")
	ErrNoMatchingArchitectures   = errors.New("no matching architectures")
	ErrNoAssets                  = errors.New("no assets")
	ErrChecksumMismatch          = errors.New("checksum mismatch")
	ErrUnsupportedChecksumFormat = errors.New("unsupported checksum format")
)

type GitHubUpdateChecker struct {
	repo string

	host string

	parsedRepo struct {
		owner string
		repo  string
	}
}

var _ updatechecker.UpdateChecker = (*GitHubUpdateChecker)(nil)

type githubAsset struct {
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	State              string `json:"state"`
	Size               int    `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type gitHubReleaseInfo struct {
	TagName     string        `json:"tag_name"`
	PublishedAt time.Time     `json:"published_at"`
	Assets      []githubAsset `json:"assets"`
}

var ErrReleaseNotFound = errors.New("release not found")

func NewGitHubUpdateChecker(fqRepo string) updatechecker.UpdateChecker {
	host := "https://api.github.com"
	owner, repo := "", ""
	repoParts := strings.Split(fqRepo, "/")
	if len(repoParts) == 2 {
		owner = repoParts[0]
		repo = repoParts[1]
	} else if len(repoParts) == 3 {
		owner = repoParts[1]
		repo = repoParts[2]
	} else {
		panic(fmt.Sprintf("invalid repo: %s", fqRepo))
	}

	return GitHubUpdateChecker{
		repo: repo,
		host: host,
		parsedRepo: struct {
			owner string
			repo  string
		}{
			owner: owner,
			repo:  repo,
		},
	}
}

// DownloadVersion will download and extract the specific version, returning
// a path to the extracted file in the archive
// it's the responsibility of the caller to clean up the extracted file
func (c GitHubUpdateChecker) DownloadVersion(version string, requireChecksumMatch bool) (string, error) {
	releaseInfo, err := getReleaseDetails(c.host, c.parsedRepo.owner, c.parsedRepo.repo, version)
	if err != nil {
		return "", errors.Wrap(err, "get release details")
	}

	asset, err := bestAsset(releaseInfo.Assets, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", errors.Wrap(err, "best asset")
	}

	archivePath, err := downloadFile(asset.BrowserDownloadURL)
	if err != nil {
		return "", errors.Wrap(err, "download file")
	}

	checksumAsset, err := checksum(releaseInfo.Assets, asset.Name)
	if err != nil {
		return "", errors.Wrap(err, "checksum")
	}

	if checksumAsset != nil {
		desiredChecksum, err := downloadAndParseChecksum(checksumAsset.BrowserDownloadURL, asset.Name)
		if err != nil {
			return "", errors.Wrap(err, "download and parse checksum")
		}

		actualChecksum, err := checksumFile(archivePath)
		if err != nil {
			return "", errors.Wrap(err, "checksum file")
		}

		if actualChecksum != desiredChecksum {
			return "", ErrChecksumMismatch
		}
	}

	return archivePath, nil
}

// GetLatestVersion will return the latest version information from the git repository
func (c GitHubUpdateChecker) GetLatestVersion() (*updatechecker.VersionInfo, error) {
	latestReleaseInfo, err := getReleaseDetails(c.host, c.parsedRepo.owner, c.parsedRepo.repo, "latest")
	if err != nil {
		return nil, errors.Wrap(err, "get release details")
	}

	latestVersion := &updatechecker.VersionInfo{
		Version:    latestReleaseInfo.TagName,
		ReleasedAt: latestReleaseInfo.PublishedAt,
	}

	return latestVersion, nil
}

func downloadAndParseChecksum(url string, assetName string) (string, error) {
	// download the file
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// parse the file
	// supported formats are sha256[whitespace]filepath per line

	// first try to find the exact match
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " ")
		if len(parts) == 2 {
			if strings.HasSuffix(strings.TrimSpace(parts[1]), assetName) {
				return strings.TrimSpace(parts[0]), nil
			}
		}
	}

	return "", ErrUnsupportedChecksumFormat
}

func checksumFile(path string) (string, error) {
	return "", nil
}

// checksumAsset will search through the assets and attempt to find the
// download url for the sha256 checksum for the asset provided
// this works by looking for the asset name with the checksum appended to it
// it will return empty string and no error if there is not checksum
func checksum(assets []githubAsset, assetName string) (*githubAsset, error) {
	for _, asset := range assets {
		if asset.State != "uploaded" {
			continue
		}

		if strings.HasPrefix(asset.Name, assetName) {
			if strings.HasSuffix(asset.Name, ".sha256") {
				return &asset, nil
			}
		}
	}

	// no exact match, look for a common checksums file
	for _, asset := range assets {
		if asset.State != "uploaded" {
			continue
		}

		if strings.Contains(asset.Name, "checksums") {
			if strings.HasSuffix(asset.Name, ".txt") {
				return &asset, nil
			}
		}
	}

	return nil, nil
}

// bestAsset will search through the assets, find the best (most appropriate)
// asset for the os nad arch provided. this will that asset
// for the asset
func bestAsset(assets []githubAsset, goos string, goarch string) (*githubAsset, error) {
	if len(assets) == 0 {
		return nil, ErrNoAssets
	}

	// find the most appropriate asset
	for _, asset := range assets {
		if asset.State != "uploaded" {
			continue
		}

		lowercaseName := strings.ToLower(asset.Name)
		if strings.Contains(lowercaseName, goos) {
			if strings.Contains(lowercaseName, goarch) {
				return &asset, nil
			}
		}
	}

	// we didn't find a specific match, look for the os with "all" for the arch
	for _, asset := range assets {
		if asset.State != "uploaded" {
			continue
		}

		lowercaseName := strings.ToLower(asset.Name)
		if strings.Contains(lowercaseName, runtime.GOOS) {
			if strings.Contains(lowercaseName, "all") {
				return &asset, nil
			}
		}
	}

	return nil, ErrNoMatchingArchitectures
}

func downloadFile(url string) (string, error) {
	tmpFile, err := ioutil.TempFile("", "usrbin")
	if err != nil {
		return "", errors.Wrap(err, "create temp file")
	}
	defer os.RemoveAll(tmpFile.Name())

	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Wrap(err, "get file")
	}

	defer resp.Body.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "copy file")
	}

	return findProbableFileInWhatMightBeAnArchive(tmpFile.Name())
}

func findProbableFileInWhatMightBeAnArchive(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "open file")
	}

	defer func() {
		if err := f.Close(); err != nil {
			logger.Error(err)
		}
	}()

	// check if it's a gzip file
	_, err = gzip.NewReader(f)
	if err == nil {
		return findProbableFileInGzip(path)
	}

	return "", ErrUnknownArchiveType
}

func findProbableFileInGzip(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "open file")
	}

	defer func() {
		if err := f.Close(); err != nil {
			logger.Error(err)
		}
	}()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", errors.Wrap(err, "open gzip file")
	}

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return "", errors.Wrap(err, "read next file")
		}

		if header.Typeflag == tar.TypeReg {
			// if the file is executable and matches the name of the current process
			// then it's almost certainly the file we want

			if isLikelyFile(header.Mode, header.Name, filepath.Base(os.Args[0])) {
				tmpFile, err := ioutil.TempFile("", "usrbin")
				if err != nil {
					return "", errors.Wrap(err, "create temp file")
				}

				defer func() {
					if err := tmpFile.Close(); err != nil {
						logger.Error(err)
					}
				}()

				if _, err := io.Copy(tmpFile, tr); err != nil {
					return "", errors.Wrap(err, "copy file")
				}

				// set the mode on the file to match
				if err := os.Chmod(tmpFile.Name(), os.FileMode(header.Mode)); err != nil {
					return "", errors.Wrap(err, "set file mode")
				}

				return tmpFile.Name(), nil
			}
		}
	}

	return "", errors.New("unable to find matching file in archive")
}

func isLikelyFile(mode int64, name string, currentExecutableName string) bool {
	if mode&0111 != 0 {
		if currentExecutableName == filepath.Base(name) {
			return true
		}
	}

	return false
}

func getReleaseDetails(host string, owner string, repo string, releaseName string) (*gitHubReleaseInfo, error) {
	uri := ""

	if releaseName == "latest" {
		uri = fmt.Sprintf("%s/repos/%s/%s/releases/latest", host, owner, repo)
	} else {
		uri = fmt.Sprintf("%s/repos/%s/%s/releases/tags/%s", host, owner, repo, releaseName)
	}

	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, errors.Wrap(err, "new request")
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "do request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrReleaseNotFound
		}

		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	releaseInfo := gitHubReleaseInfo{}

	if err := json.NewDecoder(resp.Body).Decode(&releaseInfo); err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	return &releaseInfo, nil
}
