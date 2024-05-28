package updatechecker

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_UpdateInfoFromVersions(t *testing.T) {
	var releasedAt = time.Now()

	tests := []struct {
		name           string
		currentVersion string
		latestVersion  *VersionInfo
		want           *UpdateInfo
		wantErr        bool
	}{
		{
			name:           "same versions",
			currentVersion: "1.0.0",
			latestVersion: &VersionInfo{
				Version:    "1.0.0",
				ReleasedAt: &releasedAt,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:           "latest is older",
			currentVersion: "v2.0.4",
			latestVersion: &VersionInfo{
				Version:    "v2.0.3",
				ReleasedAt: &releasedAt,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name:           "latest is newer",
			currentVersion: "v2.0.5",
			latestVersion: &VersionInfo{
				Version:    "2.0.6",
				ReleasedAt: &releasedAt,
			},
			want:    &UpdateInfo{LatestVersion: "2.0.6", LatestReleaseAt: &releasedAt, CanUpgradeInPlace: true},
			wantErr: false,
		},
		{
			name:           "latest is a prerelease of the current",
			currentVersion: "v2.0.5",
			latestVersion: &VersionInfo{
				Version:    "2.0.5-rc1",
				ReleasedAt: &releasedAt,
			},
			want:    nil,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpdateInfoFromVersions(tt.currentVersion, tt.latestVersion)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.True(t, reflect.DeepEqual(got, tt.want))
		})
	}
}
