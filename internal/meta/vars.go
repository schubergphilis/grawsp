package meta

import "time"

var (
	// AppName is the name of current application.
	AppName string = "grawsp"

	// Commit is the hash of the commit used to build the current binary.
	Commit string = "unknown"

	// BuildTime is a representation of the build process timestamp in RFC3339 format.
	BuildTime string = time.Now().Format(time.RFC3339)

	// Version is the current version of the binary.
	Version string = "dev"
)
