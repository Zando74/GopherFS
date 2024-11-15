package valueobject

import (
	"crypto/sha256"
	"encoding/hex"
)

type ChunkHash string

func NewChunkHash(chunk Chunk, stableContext string) ChunkHash {
	hash := sha256.Sum256(append(chunk, stableContext...))
	return ChunkHash(hex.EncodeToString(hash[:]))
}
