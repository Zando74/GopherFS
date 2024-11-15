package valueobject

import (
	"fmt"

	"github.com/Zando74/GopherFS/control-plane/config"
)

type Chunk []byte

type ChunkSizeExceededError struct{}

func (e *ChunkSizeExceededError) Error() string {
	return fmt.Sprintf("chunk size exceeds %d bytes", config.ConfigSingleton.GetInstance().FileStorage.Chunk_size)
}

func NewChunk(data []byte) (*Chunk, error) {
	if uint32(len(data)) > config.ConfigSingleton.GetInstance().FileStorage.Chunk_size {
		return nil, &ChunkSizeExceededError{}
	}
	chunk := Chunk(data)
	return &chunk, nil
}
