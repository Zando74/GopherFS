package entity

type Location struct {
	StorageID string
}

type ChunkLocations struct {
	ChunkID            string
	SequenceNumber     int
	PrimaryLocation    Location
	SecondaryLocations []Location
}

type FileMetadata struct {
	FileID   string
	FileName string
	FileSize int64
	Chunks   []ChunkLocations
}
