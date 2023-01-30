package usrbin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_findProbableFileInWhatMightBeAnArchive(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findProbableFileInWhatMightBeAnArchive(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("findProbableFileInWhatMightBeAnArchive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("findProbableFileInWhatMightBeAnArchive() = %v, want %v", got, tt.want)
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
