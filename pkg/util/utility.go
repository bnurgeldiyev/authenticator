package util

import (
	"math/rand"
	"strconv"
	"time"
)

func Int64FromString(s string) (n int64, err error) {
	n, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return
	}
	return
}

func VersionInc(cv int) int {
	nv := cv + 1
	if nv > 10000 {
		// reset version each 10K updates
		return 0
	}
	return nv
}

func PageToLimitOffset(page, perPage uint64) (offset, limit uint64) {
	if page < 1 {
		page = 1
	}
	if perPage > 100 {
		perPage = 100
	}
	limit = perPage
	offset = perPage * (page - 1)
	return
}

func RandomString(alphabet string, length int) string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	result := ""
	for i := 0; i < length; i++ {
		result += string(alphabet[rnd.Intn(len(alphabet)-1)])
	}
	return result
}
