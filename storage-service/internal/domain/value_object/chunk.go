package value_object

type Chunk []byte

func NewChunk(data []byte) (*Chunk, error) {
	chunk := Chunk(data)
	return &chunk, nil
}
