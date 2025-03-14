package util

func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	var chunks [][]T
	for chunkSize < len(slice) {
		slice, chunks = slice[chunkSize:], append(chunks, slice[0:chunkSize:chunkSize])
	}
	return append(chunks, slice)
}
