package cache

import (
	"bytes"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

// serialize encodes the given value into MessagePack bytes.
func serialize[T any](value T) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	if err := enc.Encode(value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// deserialize decodes the given MessagePack bytes into a value.
func deserialize[T any](data []byte) (T, error) {
	var value T
	dec := msgpack.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&value); err != nil {
		return value, err
	}
	return value, nil
}

// serializeStream encodes the given value and writes it to the provided writer in MessagePack format.
func serializeStream[T any](writer io.Writer, value T) error {
	enc := msgpack.NewEncoder(writer)
	return enc.Encode(value)
}

// deserializeStream reads and decodes a value from the provided reader in MessagePack format.
func deserializeStream[T any](reader io.Reader) (T, error) {
	var value T
	dec := msgpack.NewDecoder(reader)
	if err := dec.Decode(&value); err != nil {
		return value, err
	}
	return value, nil
}
