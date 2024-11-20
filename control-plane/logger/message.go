package logger

const (
	// Controller
	RunningApplicationMessage = "Running application"
	ConfigLoadedMessage       = "config: %s"

	// use cases
	StartingFileUploadSagaStepMessage       = "STARTING_SAGA_STEP : %s, FileID: %s"
	FinishedFileUploadSagaStepMessage       = "FINISHED_SAGA_STEP : %s, FileID: %s"
	FailedFileUploadSagaStepMessage         = "FAILED_SAGA_STEP : %s, FileID: %s"
	RecoveringFileUploadSagaStepSagaMessage = "RECOVERING_SAGA_STEP : %s, FileID: %s"
)
