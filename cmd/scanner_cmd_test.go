package main

import (
	"os"
	"talisman/git_testing"
	"talisman/talismanrc"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScannerCmdRunsSuccessfully(t *testing.T) {
	git_testing.DoInTempGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("some-dir/should-be-included.txt", "safeContents")
		git.AddAndcommit("*", "Start of Scan")
		os.Chdir(git.Root())

		scannerCmd := NewScannerCmd(true, &talismanrc.TalismanRC{}, git.Root())
		scannerCmd.Run()
		assert.Equal(t, 0, scannerCmd.exitStatus(), "Expected ScannerCmd.exitStatus() to return 0 since no secret is found")
	})
}

func TestScannerCmdDetectsSecretAndFails(t *testing.T) {
	git_testing.DoInTempGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("some-dir/file-with-secret.txt", awsAccessKeyIDExample)
		git.AddAndcommit("*", "Initial Commit")
		git.RemoveFile("some-dir/file-with-secret.txt")
		git.AddAndcommit("*", "Removed secret")
		git.CreateFileWithContents("some-dir/safe-file.txt", "safeContents")
		git.AddAndcommit("*", "Start of Scan")
		os.Chdir(git.Root())

		scannerCmd := NewScannerCmd(false, &talismanrc.TalismanRC{}, git.Root())
		scannerCmd.Run()
		assert.Equal(t, 1, scannerCmd.exitStatus(), "Expected ScannerCmd.exitStatus() to return 1 since secret present in history")
	})
}

func TestScannerCmdAddingSecretKeyShouldExitZeroIfFileIsWithinConfiguredScope(t *testing.T) {
	git_testing.DoInTempGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("go.sum", awsAccessKeyIDExample)
		git.CreateFileWithContents("go.mod", awsAccessKeyIDExample)
		git.AddAndcommit("*", "go sum file")
		os.Chdir(git.Root())

		tRC := &talismanrc.TalismanRC{ScopeConfig: []talismanrc.ScopeConfig{{ScopeName: "go"}}}
		scannerCmd := NewScannerCmd(false, tRC, git.Root())
		scannerCmd.Run()
		assert.Equal(t, 0, scannerCmd.exitStatus(), "Expected ScannerCmd.exitStatus() to return 0 since no secret is found")
	})
}

func TestScannerCmdDetectsSecretAndIgnoresWhileRunningInIgnoreHistoryModeWithValidIgnoreConf(t *testing.T) {
	git_testing.DoInTempGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("go.sum", awsAccessKeyIDExample)
		git.CreateFileWithContents("go.mod", awsAccessKeyIDExample)
		git.AddAndcommit("*", "go sum file")
		os.Chdir(git.Root())

		tRC := &talismanrc.TalismanRC{
			FileIgnoreConfig: []talismanrc.FileIgnoreConfig{
				{FileName: "go.sum", Checksum: "582093519ae682d5170aecc9b935af7e90ed528c577ecd2c9dd1fad8f4924ab9"},
				{FileName: "go.mod", Checksum: "8a03b9b61c505ace06d590d2b9b4f4b6fa70136e14c26875ced149180e00d1af"},
			}}
		scannerCmd := NewScannerCmd(true, tRC, git.Root())
		scannerCmd.Run()
		assert.Equal(t, 0, scannerCmd.exitStatus(), "Expected ScannerCmd.exitStatus() to return 0 since secrets file ignore is enabled")
	})
}

func TestScannerCmdDetectsSecretWhileRunningNormalScanMode(t *testing.T) {
	git_testing.DoInTempGitRepo(func(git *git_testing.GitTesting) {
		git.SetupBaselineFiles("simple-file")
		git.CreateFileWithContents("go.sum", awsAccessKeyIDExample)
		git.CreateFileWithContents("go.mod", awsAccessKeyIDExample)
		git.AddAndcommit("*", "go sum file")
		os.Chdir(git.Root())

		tRC := &talismanrc.TalismanRC{
			FileIgnoreConfig: []talismanrc.FileIgnoreConfig{
				{FileName: "go.sum", Checksum: "582093519ae682d5170aecc9b935af7e90ed528c577ecd2c9dd1fad8f4924ab9"},
				{FileName: "go.mod", Checksum: "8a03b9b61c505ace06d590d2b9b4f4b6fa70136e14c26875ced149180e00d1af"},
			}}
		scannerCmd := NewScannerCmd(false, tRC, git.Root())
		scannerCmd.Run()
		assert.Equal(t, 1, scannerCmd.exitStatus(), "Expected ScannerCmd.exitStatus() to return 1 because file ignore is disabled when scanning history")
	})
}
