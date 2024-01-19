package imports

import "fmt"

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

type CopyFromSourceNotReady struct {
	sourceType string
}

func (e CopyFromSourceNotReady) Error() string {
	return fmt.Sprintf("la source n'a pas été initialisée (%s)", e.sourceType)
}

type CopyFromSourceDepleted struct {
	sourceType string
}

func (e CopyFromSourceDepleted) Error() string {
	return fmt.Sprintf("la source est épuisée (%s)", e.sourceType)
}
