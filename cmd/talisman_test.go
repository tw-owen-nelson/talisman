package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func Test_setLogLevel(t *testing.T) {
	levels := []string{"error", "warn", "info", "debug", "unknown"}
	expectedLogrusLevels := []logrus.Level{
		logrus.ErrorLevel, logrus.WarnLevel,
		logrus.InfoLevel, logrus.DebugLevel, logrus.ErrorLevel}

	for idx, level := range levels {
		options.LogLevel = level
		setLogLevel()
		assert.True(
			t,
			logrus.IsLevelEnabled(expectedLogrusLevels[idx]),
			fmt.Sprintf("expected level to be %v for options.LogLevel = %s", expectedLogrusLevels[idx], level))

		options.Debug = true
		setLogLevel()
		assert.True(
			t,
			logrus.IsLevelEnabled(logrus.DebugLevel),
			"expected level to be debug when options.Debug is set")
	}
}

func Test_validateGitExecutable(t *testing.T) {
	t.Run("given operating systems is windows", func(t *testing.T) {

		operatingSystem := "windows"
		os.Setenv("PATHEXT", ".COM;.EXE;.BAT;.CMD;.VBS;.VBE;.JS;.JSE;.WSF;.WSH;.MSC")

		t.Run("should return error if git executable exists in current directory", func(t *testing.T) {
			fs := afero.NewMemMapFs()
			gitExecutable := "git.exe"
			afero.WriteFile(fs, gitExecutable, []byte("git executable"), 0700)
			err := validateGitExecutable(fs, operatingSystem)
			assert.EqualError(t, err, "not allowed to have git executable located in repository: git.exe")
		})

		t.Run("should return nil if git executable does not exist in current directory", func(t *testing.T) {
			err := validateGitExecutable(afero.NewMemMapFs(), operatingSystem)
			assert.NoError(t, err)
		})

	})

	t.Run("given operating systems is linux", func(t *testing.T) {

		operatingSystem := "linux"

		t.Run("should return nil if git executable exists in current directory", func(t *testing.T) {
			fs := afero.NewMemMapFs()
			gitExecutable := "git.exe"
			afero.WriteFile(fs, gitExecutable, []byte("git executable"), 0700)
			err := validateGitExecutable(fs, operatingSystem)
			assert.NoError(t, err)
		})

		t.Run("should return nil if git executable does not exist in current directory", func(t *testing.T) {
			err := validateGitExecutable(afero.NewMemMapFs(), operatingSystem)
			assert.NoError(t, err)
		})

	})
}
