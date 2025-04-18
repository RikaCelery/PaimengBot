package utils

import (
	"hash/crc64"
	"math/rand"
	"time"
	"unsafe"
)

// RandSenderPerDayN 每个用户每天随机数
func RandSenderPerDayN(uid int64, n int) int {
	sum := crc64.New(crc64.MakeTable(crc64.ISO))
	sum.Write([]byte(time.Now().Format("20060102")))
	sum.Write((*[8]byte)(unsafe.Pointer(&uid))[:])
	r := rand.New(rand.NewSource(int64(sum.Sum64())))
	return r.Intn(n)
}
