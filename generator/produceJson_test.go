package generator

import (
	"math/rand"
	"testing"
	"time"

	"MyLoadGen/request"
	"encoding/json"
)

var operators = []string{"+", "-", "*", "/"}

func TestJson(t *testing.T) {
	id := time.Now().UnixNano()
	sreq := request.ServerReq{
		ID: id,
		Operands: []int{
			int(rand.Int31n(1000) + 1),
			int(rand.Int31n(1000) + 1)},
		Operator: func() string {
			return operators[rand.Int31n(100)%4]
		}(),
	}
	bytes, err := json.Marshal(sreq)
	if err != nil {
		t.Log(err)
	}
	t.Log(string(bytes))
}
