package shared

type CmdResp string

type ApkBuildingDone struct{}

type ApkZipped struct{}

type FileUploaded struct{}

type CmdError struct {
	Err error
}
