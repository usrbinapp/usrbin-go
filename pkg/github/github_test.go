package github

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_findProbableFileInWhatMightBeAnArchive(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		wantErr error
	}{
		{
			name:    "not an archive file",
			content: "",
			want:    "",
			wantErr: ErrUnknownArchiveType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, err := ioutil.TempFile("", "test")
			require.NoError(t, err)
			defer tmp.Close()

			_, err = tmp.Write([]byte(tt.content))
			require.NoError(t, err)

			req := require.New(t)

			got, err := findProbableFileInWhatMightBeAnArchive(tmp.Name())
			if tt.wantErr == nil {
				req.NoError(err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
			}
		})
	}
}

func Test_isLikelyFile(t *testing.T) {
	tests := []struct {
		name                  string
		mode                  int64
		filename              string
		currentExecutableName string
		want                  bool
	}{
		{
			name:                  "executable",
			mode:                  0755,
			filename:              "foo",
			currentExecutableName: "foo",
			want:                  true,
		},
		{
			name:                  "not executable",
			mode:                  0444,
			filename:              "foo",
			currentExecutableName: "foo",
			want:                  false,
		},
		{
			name:                  "executable, wrong filename",
			mode:                  0755,
			filename:              "foo2",
			currentExecutableName: "foo",
			want:                  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isLikelyFile(tt.mode, tt.filename, tt.currentExecutableName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_bestAsset(t *testing.T) {
	tests := []struct {
		name    string
		assets  []githubAsset
		goos    string
		goarch  string
		want    *githubAsset
		wantErr error
	}{
		{
			name:    "no assets",
			want:    nil,
			wantErr: ErrNoAssets,
		},
		{
			name: "one asset",
			assets: []githubAsset{
				{
					Name:               "foo_linux_amd64",
					State:              "uploaded",
					BrowserDownloadURL: "https://usrbin.app/foo_linux_amd64",
				},
			},
			goos:   "linux",
			goarch: "amd64",
			want: &githubAsset{
				Name:               "foo_linux_amd64",
				State:              "uploaded",
				BrowserDownloadURL: "https://usrbin.app/foo_linux_amd64",
			},
		},
		{
			name: "multiple assets, one matching",
			assets: []githubAsset{
				{
					Name:               "foo_darwin_amd64",
					State:              "uploaded",
					BrowserDownloadURL: "https://usrbin.app/foo_darwin_amd64",
				},
				{
					Name:               "foo_linux_arm64",
					State:              "uploaded",
					BrowserDownloadURL: "https://usrbin.app/foo_linux_arm64",
				},
				{
					Name:               "foo_linux_amd64",
					State:              "uploaded",
					BrowserDownloadURL: "https://usrbin.app/foo_linux_amd64",
				},
				{
					Name:               "checksums.txt",
					State:              "uploaded",
					BrowserDownloadURL: "https://usrbin.app/checksums.txt",
				},
			},
			goos:   "linux",
			goarch: "amd64",
			want: &githubAsset{
				Name:               "foo_linux_amd64",
				State:              "uploaded",
				BrowserDownloadURL: "https://usrbin.app/foo_linux_amd64",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)
			got, err := bestAsset(tt.assets, tt.goos, tt.goarch)
			if tt.wantErr == nil {
				req.NoError(err)
				assert.Equal(t, tt.want, got)
			} else {
				assert.EqualError(t, err, tt.wantErr.Error())
			}
		})
	}
}

func Test_checksum(t *testing.T) {
	tests := []struct {
		name   string
		assets []githubAsset
		asset  githubAsset
		want   *githubAsset
	}{
		{
			name: "goreleaser style",
			assets: []githubAsset{
				{
					Name:  "cli_0.1.0_checksums.txt",
					State: "uploaded",
				},
				{
					Name:  "foo_linux_amd64.tar.gz",
					State: "uploaded",
				},
				{
					Name:  "foo_darwin_amd64.tar.gz",
					State: "uploaded",
				},
				{
					Name:  "foo_windows_all.tar.gz",
					State: "uploaded",
				},
			},
			asset: githubAsset{
				Name: "foo_linux_amd64.tar.gz",
			},
			want: &githubAsset{
				Name:  "cli_0.1.0_checksums.txt",
				State: "uploaded",
			},
		},
		{
			name: "with .sha256 extension",
			assets: []githubAsset{
				{
					Name:  "foo_linux_amd64.sha256",
					State: "uploaded",
				},
				{
					Name:  "foo_linux_amd64",
					State: "uploaded",
				},
			},
			asset: githubAsset{
				Name:  "foo_linux_amd64",
				State: "uploaded",
			},
			want: &githubAsset{
				Name:  "foo_linux_amd64.sha256",
				State: "uploaded",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := require.New(t)

			got, err := checksum(tt.assets, tt.asset.Name)
			req.NoError(err)

			assert.Equal(t, tt.want, got)
		})
	}
}
