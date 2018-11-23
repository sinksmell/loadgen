package generator

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestJavaJson(t *testing.T) {
	buf := bytes.NewBuffer([]byte{})
	request, err := http.NewRequest("GET", "http://127.0.0.1:8080/json", buf)
	if err != nil {
		t.Error(err)
	}
	// req.Header.Set("X-Custom-Header", "myvalue")
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(request)

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	t.Log(string(data))
}
