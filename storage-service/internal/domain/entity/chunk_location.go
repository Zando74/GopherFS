package entity

type Location struct {
	StorageID string
}

type ChunkLocations struct {
	ChunkID            string
	SequenceNumber     uint32
	PrimaryLocation    Location
	SecondaryLocations []Location
}
