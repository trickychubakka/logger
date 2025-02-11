package internal

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPrintStartMessage(t *testing.T) {
	type args struct {
		buildVersion string
		buildDate    string
		buildCommit  string
	}
	tests := []struct {
		name string
		args args
		want args
	}{
		{
			name: "Positive Test PrintStartMessage",
			args: args{
				buildVersion: "version",
				buildDate:    "date",
				buildCommit:  "commit",
			},
			want: args{
				buildVersion: "Build version: version",
				buildDate:    "Build date: date",
				buildCommit:  "Build commit: commit",
			},
		},
		{
			name: "Positive N/A PrintStartMessage",
			args: args{
				buildVersion: "",
				buildDate:    "",
				buildCommit:  "",
			},
			want: args{
				buildVersion: "Build version: N/A",
				buildDate:    "Build date: N/A",
				buildCommit:  "Build commit: N/A",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PrintStartMessage(tt.args.buildVersion, tt.args.buildDate, tt.args.buildCommit)
			assert.Equal(t, got["version"], tt.want.buildVersion)
			assert.Equal(t, got["date"], tt.want.buildDate)
			assert.Equal(t, got["commit"], tt.want.buildCommit)
		})
	}
}
