package upload_saga

import (
	"fmt"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/repository"
)

type FileSaveMetadata struct {
	entity.ConcreteSagaStep
	FileMetadataSaver repository.FileMetadataRepository
	FileMetadata      *entity.FileMetadata
}

func NewFileSaveMetadata(fileMetadataSaver repository.FileMetadataRepository, fileMetadata *entity.FileMetadata) *FileSaveMetadata {
	return &FileSaveMetadata{
		FileMetadataSaver: fileMetadataSaver,
		FileMetadata:      fileMetadata,
	}
}

func (f *FileSaveMetadata) Execute() error {
	f.ExpectConfirmation()
	f.StartWaitingForConfirmations()
	err := f.FileMetadataSaver.SaveFileMetadata(f.FileMetadata, f.Ack)
	if err != nil {
		return err
	}
	return nil
}

func (f *FileSaveMetadata) Transaction() error {
	f.SetTTL(config.ConfigSingleton.GetInstance().FileStorage.Saga_ttl)
	err := f.Execute()

	if err != nil {
		f.Fail()
	}

	for {
		f.DecreaseTTL()
		if f.IsWaitingForConfirmations() {
			if f.NoPendingConfirmation() {
				return nil
			}
		}

		if f.IsFailed() {
			return fmt.Errorf("Canceled")
		}

		if f.IsOverTTL() {
			return fmt.Errorf("TTL expired")
		}
	}

}

func (f *FileSaveMetadata) Rollback() error {
	f.FileMetadataSaver.DeleteFileMetadata(f.FileMetadata.FileID)
	return nil
}
