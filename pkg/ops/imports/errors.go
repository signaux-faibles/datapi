package imports

type ChunkHasNullSizeError struct{}

func (e ChunkHasNullSizeError) Error() string {
	return "chunk is zero length, channel might be closed"
}

type InvalidParameterError struct {
	message string
}

func (e InvalidParameterError) Error() string {
	return "invalid parameter: " + e.message
}
