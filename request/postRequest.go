package request

import (
	. "MyLoadGen/lib"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

type MyCaller interface {
	//构建请求
	BuildReq() RawReq
	//调用
	Caller(req []byte, timeoutNs time.Duration) ([]byte, error)
	//检查响应
	CheckResp(req RawReq, resp RawResp) *CallResult
}

type HttpPOSTRequest struct {
	url string
}

func NewPostRequest(url string) Caller {
	return &HttpPOSTRequest{url: url}
}

func (httpreq *HttpPOSTRequest) BuildReq() RawReq {
	id := time.Now().UnixNano()
	sreq := ServerReq{
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
		panic(err)
	}
	rawReq := RawReq{ID: id, Req: bytes}
	return rawReq
}

func (httpreq *HttpPOSTRequest) Caller(req []byte, timeoutNs time.Duration) ([]byte, error) {
	request, err := http.NewRequest("POST", httpreq.url, bytes.NewBuffer(req))
	// req.Header.Set("X-Custom-Header", "myvalue")
	request.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: timeoutNs}
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)

	return data, err
}

func (httpreq *HttpPOSTRequest) CheckResp(rawReq RawReq, rawResp RawResp) *CallResult {
	var commResult CallResult
	//commResult.ID = rawResp.ID
	commResult.ID = rawReq.ID
	commResult.Req = rawReq
	commResult.Resp = rawResp
	var sreq ServerReq
	err := json.Unmarshal(rawReq.Req, &sreq)
	if err != nil {
		commResult.Code = RET_CODE_FATAL_CALL
		commResult.Msg =
			fmt.Sprintf("Incorrectly formatted Req: %s!\n", string(rawReq.Req))
		return &commResult
	}
	var sresp ServerResp
	err = json.Unmarshal(rawResp.Resp, &sresp)
	if err != nil {
		commResult.Code = RET_CODE_ERROR_RESPONSE
		commResult.Msg =
			fmt.Sprintf("Incorrectly formatted Resp: %s!\n", string(rawResp.Resp))
		return &commResult
	}
	if sresp.ID != sreq.ID {
		commResult.Code = RET_CODE_ERROR_RESPONSE
		commResult.Msg =
			fmt.Sprintf("Inconsistent raw id! (%d != %d)\n", rawReq.ID, rawResp.ID)
		return &commResult
	}
	if sresp.Err != nil {
		commResult.Code = RET_CODE_ERROR_CALEE
		commResult.Msg =
			fmt.Sprintf("Abnormal server: %s!\n", sresp.Err)
		return &commResult
	}
	if sresp.Result != op(sreq.Operands, sreq.Operator) {
		commResult.Code = RET_CODE_ERROR_RESPONSE
		commResult.Msg =
			fmt.Sprintf(
				"Incorrect result: %s!\n",
				genFormula(sreq.Operands, sreq.Operator, sresp.Result, false))
		return &commResult
	}
	commResult.Code = RET_CODE_SUCCESS
	commResult.Msg = fmt.Sprintf("Success. (%s)", sresp.Formula)
	return &commResult
}
