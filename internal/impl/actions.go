package impl

import (
	"time"

	"github.com/briandowns/spinner"

	"github.com/m-horky/insights-client-next/internal"
)

type spin struct {
	spin *spinner.Spinner
}

var Spinner = spin{spin: spinner.New(spinner.CharSets[14], 100*time.Millisecond)}

func (s *spin) Maybe(input *Input, message string) {
	if input.Format != internal.Human || input.Debug {
		return
	}
	s.spin.Suffix = " " + message
	s.spin.Start()
}

func (s *spin) Stop() {
	if s.spin.Active() {
		s.spin.Stop()
	}
}

type InputAction uint

// TODO Explore the possibility to run --group on its own with lightweight collection,
//	instead of triggering the default collector.

const (
	ANone InputAction = iota
	AHelp
	ARegister
	AUnregister
	AStatus
	ACheckIn
	ASetDisplayName
	ASetAnsibleHostname
	ARunModule
	AUploadLocalArchive
	ATestConnection
	ASupport
	ASetGroupLocally
)

type Input struct {
	Action InputAction
	Debug  bool
	Format internal.Format
	Args   any
}

type ARegisterArgs struct {
	Group           string
	DisplayName     string
	AnsibleHostname string
}

type ASetDisplayNameArgs struct {
	Name string
}

type ASetAnsibleHostnameArgs struct {
	Name string
}

type ARunModuleArgs struct {
	Command []string
	Options []string
	// ArchiveParent is a filesystem path to a directory.
	ArchiveParent string
	// ArchiveName is a filename with no parents and no extension.
	ArchiveName string
	// StopAtDir performs collection, but not compression, upload or deletion of the directory.
	StopAtDir bool
	// StopAtFile performs collection and compression, but not upload or deletion of the archive.
	StopAtFile bool
	// KeepArchive performs collection, compression and upload, but not deletion of an archive.
	StopAtCleanup bool
}

type AUploadLocalArchiveArgs struct {
	Path        string
	ContentType string
}
