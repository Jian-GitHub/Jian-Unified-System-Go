package jobResultStatus

const (
	offset        = -3
	RUNNING_ERROR = iota + offset
	COMPILATION_ERROR
	QUEUED
	PROCESSING
	FINISHED
)
