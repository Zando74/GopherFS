package entity

type SagaInformation struct {
	FileID    string
	StepName  string
	CreatedAt int64
	Status    SagaState
}
