package hash

import (
	"crypto/md5"
	"fmt"
	"hash/fnv"
)

func Hash(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(Md5(data))
	return h.Sum32()
}

func Md5(data []byte) []byte {
	digest := md5.New()
	digest.Write(data)
	return digest.Sum(nil)
}

func Md5Hex(data []byte) string {
	return fmt.Sprintf("%x", Md5(data))
}
