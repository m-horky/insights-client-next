package internal

var Version = "development"

// ConfigPath points to a file where a configuration file is stored.
var ConfigPath = "/etc/insights-client/insights-client.conf"

// LogPath points to a file into which logs should be written.
var LogPath = "/var/log/insights-client/insights-client.log"

// ArchiveDirectoryParentPath is a parent directory for archive directories.
var ArchiveDirectoryParentPath = "/var/cache/insights-client/"

// DefaultModuleName is run when CLI did not specify anything else.
var DefaultModuleName = "advisor"

// MachineIDFilePath points to a file where the client UUID is stored.
var MachineIDFilePath = "/etc/insights-client/machine-id"
var DotRegisteredPath = "/etc/insights-client/.registered"
var DotUnregisteredPath = "/etc/insights-client/.unregistered"
