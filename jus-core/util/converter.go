package util

import "encoding/binary"

func Int64ToBytes(i int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(i))
	return b
}

func BytesToInt64(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}
