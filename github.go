package usrbin

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type GitHubUpdateChecker struct {
	repo string

	host     string
	apiToken string

	parsedRepo struct {
		owner string
		repo  string
	}
}

type gitHubReleaseInfo struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
	Assets      []struct {
		Name               string `json:"name"`
		ContentType        string `json:"content_type"`
		State              string `json:"state"`
		Size               int    `json:"size"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var ErrReleaseNotFound = errors.New("release not found")

func NewGitHubUpdateChecker(fqRepo string) UpdateChecker {
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
		repo:     repo,
		host:     host,
		apiToken: os.Getenv("GITHUB_TOKEN"),
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
func (c GitHubUpdateChecker) DownloadVersion(version string) (string, error) {
	releaseInfo, err := getReleaseDetails(c.host, c.apiToken, c.parsedRepo.owner, c.parsedRepo.repo, version)
	if err != nil {
		return "", errors.Wrap(err, "get release details")
	}

	if len(releaseInfo.Assets) == 0 {
		return "", errors.New("no assets found")
	}

	// find the most appropriate asset
	for _, asset := range releaseInfo.Assets {
		if asset.State != "uploaded" {
			continue
		}

		lowercaseName := strings.ToLower(asset.Name)
		if strings.Contains(lowercaseName, runtime.GOOS) {
			if strings.Contains(lowercaseName, runtime.GOARCH) {
				return downloadFile(asset.BrowserDownloadURL)
			}
		}
	}

	// we didn't find a specific match, look for the os with "all" for the arch
	for _, asset := range releaseInfo.Assets {
		if asset.State != "uploaded" {
			continue
		}

		lowercaseName := strings.ToLower(asset.Name)
		if strings.Contains(lowercaseName, runtime.GOOS) {
			if strings.Contains(lowercaseName, "all") {
				return downloadFile(asset.BrowserDownloadURL)
			}
		}
	}

	return "", errors.New("unable to find file matching architecture")
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
	defer f.Close()

	// check if it's a gzip file
	_, err = gzip.NewReader(f)
	if err == nil {
		return findProbableFileInGzip(path)
	}

	return "", errors.New("unable to determine file type of archive")
}

func findProbableFileInGzip(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", errors.Wrap(err, "open file")
	}
	defer f.Close()

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

				defer tmpFile.Close()
				if _, err := io.Copy(tmpFile, tr); err != nil {
					log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
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

// GetLatestVersion will return the latest version from the git repository
func (c GitHubUpdateChecker) GetLatestVersion() (*VersionInfo, error) {
	latestReleaseInfo, err := getReleaseDetails(c.host, c.apiToken, c.parsedRepo.owner, c.parsedRepo.repo, "latest")
	if err != nil {
		return nil, errors.Wrap(err, "get release details")
	}

	latestVersion := &VersionInfo{
		Version:    latestReleaseInfo.TagName,
		ReleasedAt: latestReleaseInfo.PublishedAt,
	}

	return latestVersion, nil
}

func getReleaseDetails(host string, token string, owner string, repo string, releaseName string) (*gitHubReleaseInfo, error) {
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

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
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
