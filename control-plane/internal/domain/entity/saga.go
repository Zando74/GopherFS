package entity

type SagaState int

const (
	Initialization SagaState = iota
	Pending
	WaitingForConfirmation
	Done
	Failed
)

func (s SagaState) String() string {
	return []string{"Initialization", "Pending", "Done", "Failed"}[s]
}

type Saga struct {
	Steps []SagaStep
	State SagaState
}

func (s *Saga) Execute() error {
	for i, step := range s.Steps {
		err := step.Transaction()
		if err != nil {
			for j := i; j >= 0; j-- {
				s.Steps[j].Rollback()
			}
			return err
		}
	}
	return nil
}
