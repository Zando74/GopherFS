package saga_entity

type SagaRecoveryStrategy string

const (
	RollBack SagaRecoveryStrategy = "ROLLBACK"
	Retry    SagaRecoveryStrategy = "RETRY"
)

type StepName string

const (
	FileChunkSave    StepName = "FILE_CHUNK_SAVE"
	FileMetadataSave StepName = "FILE_METADATA_SAVE"
	ReplicateFile    StepName = "REPLICATE_FILE"
)

type SagaName string

const (
	UploadSaga SagaName = "UPLOAD_SAGA"
)

type SagaStatus string

const (
	Pending SagaStatus = "PENDING"
	Done    SagaStatus = "DONE"
	Failed  SagaStatus = "FAILED"
)

type SagaInformation struct {
	FileID           string
	SagaName         SagaName
	StepName         StepName
	CreatedAt        int64
	Status           SagaStatus
	RecoveryStrategy SagaRecoveryStrategy
}
