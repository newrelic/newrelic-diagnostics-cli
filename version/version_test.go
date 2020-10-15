package version

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/newrelic/NrDiag/config"
	"github.com/newrelic/NrDiag/logger"
)

var promptUserAllow = func(string) bool { return true }
var promptUserDeny = func(string) bool { return false }

var runningLog string

type loggerMethods struct{}

func (l loggerMethods) Info(i ...interface{}) {
	runningLog += fmt.Sprintf("%s\n", i)
}

func (l loggerMethods) Infof(s string, f ...interface{}) {
	runningLog += fmt.Sprintf(s, f...) + "\n"
}

func (l loggerMethods) FixedPrefix(int, string, string) { return }
func (l loggerMethods) Debug(...interface{})            { return }
func (l loggerMethods) Debugf(string, ...interface{})   { return }
func (l loggerMethods) Dump(...interface{})             { return }

var logCapture = loggerMethods{}

var getOnlineVersionWrong = func(log logger.API) string { return "test.version.wrong" }
var getOnlineVersionCorrect = func(log logger.API) string { return config.Version }
var getLatestVersionTest = func(logger.API) error { return nil }

func Test_processAutoVersionCheck(t *testing.T) {
	type args struct {
		logger           logger.API
		getOnlineVersion func(logger.API) string
	}

	tests := []struct {
		name   string
		args   args
		appVer string
		want   *regexp.Regexp
		not    *regexp.Regexp
	}{
		{
			name:   "Complains when version is wrong",
			args:   args{logCapture, getOnlineVersionWrong},
			appVer: "1.2.3",
			want:   regexp.MustCompile("is newer than your version"),
		},
		{
			name:   "Says nothing when version is the same",
			args:   args{logCapture, getOnlineVersionCorrect},
			appVer: "1.2.3",
			not:    regexp.MustCompile("is newer than your version"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runningLog = ""
			config.Version = tt.appVer
			processAutoVersionCheck(tt.args.logger, tt.args.getOnlineVersion)
			if tt.want != nil && !tt.want.MatchString(runningLog) {
				t.Error("Failed on match: '" + tt.want.String() + "' pattern not found.\n" + runningLog)
			}
			if tt.not != nil && tt.not.MatchString(runningLog) {
				t.Error("Matched on unwanted pattern: '" + tt.want.String() + "'\n" + runningLog)
			}
		})
	}
}

func Test_processVersion(t *testing.T) {
	type args struct {
		logger           logger.API
		promptUser       func(string) bool
		getOnlineVersion func(logger.API) string
		getLatestVersion func(logger.API) error
	}

	tests := []struct {
		name   string
		args   args
		appVer string
		want   *regexp.Regexp
		not    *regexp.Regexp
	}{
		{
			name:   "Attempts to check for newer version",
			args:   args{logCapture, promptUserAllow, getOnlineVersionWrong, getLatestVersionTest},
			appVer: "1.2.3",
			want:   regexp.MustCompile("Checking for newer version"),
		},
		{
			name:   "Attempts to download",
			args:   args{logCapture, promptUserAllow, getOnlineVersionWrong, getLatestVersionTest},
			appVer: "1.2.3",
			want:   regexp.MustCompile("Downloading latest version"),
		},
		{
			name:   "Does not attempt to check for newer version",
			args:   args{logCapture, promptUserDeny, getOnlineVersionWrong, getLatestVersionTest},
			appVer: "1.2.3",
			not:    regexp.MustCompile("Checking for newer version"),
		},
		{
			name:   "Does not download",
			args:   args{logCapture, promptUserDeny, getOnlineVersionWrong, getLatestVersionTest},
			appVer: "1.2.3",
			not:    regexp.MustCompile("Downloading latest version"),
		},
		{
			name:   "Does not download if version matches",
			args:   args{logCapture, promptUserAllow, getOnlineVersionCorrect, getLatestVersionTest},
			appVer: "1.2.3",
			not:    regexp.MustCompile("Downloading latest version"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runningLog = ""
			config.Version = tt.appVer
			processVersion(tt.args.logger, tt.args.promptUser, tt.args.getOnlineVersion, tt.args.getLatestVersion)
			if tt.want != nil && !tt.want.MatchString(runningLog) {
				t.Error("Failed on match: '" + tt.want.String() + "' pattern not found.\n" + runningLog)
			}
			if tt.not != nil && tt.not.MatchString(runningLog) {
				t.Error("Matched on unwanted pattern: '" + tt.want.String() + "'\n" + runningLog)
			}
		})
	}
}
