package request

import (
	"MyLoadGen/lib"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

//创建简单的http get请求
type HttpGETRequest struct {
	url string
}

//为了json序列化，添加一个字段
type clientRequest struct {
	Url string
}

//只记录服务器的返回状态码
type serverResponse struct {
	StatusCode int
}

func (request *HttpGETRequest) BuildReq() lib.RawReq {
	id := time.Now().UnixNano()
	sreq := clientRequest{request.url}
	bytes, err := json.Marshal(sreq)
	if err != nil {
		panic(err)
	}
	rawReq := lib.RawReq{ID: id, Req: bytes}
	return rawReq
}

func (getRequest *HttpGETRequest) Caller(req []byte, timeoutNs time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeoutNs}
	resp, err := client.Get(getRequest.url)
	if err != nil {
		return nil, err
	}

	sresp := serverResponse{}
	sresp.StatusCode = resp.StatusCode
	bytes, err := json.Marshal(sresp)
	return bytes, err
}

func (request *HttpGETRequest) CheckResp(req lib.RawReq, resp lib.RawResp) *lib.CallResult {

	var commResult lib.CallResult
	commResult.ID = req.ID
	commResult.Req = req
	commResult.Resp = resp

	var sreq clientRequest
	err := json.Unmarshal(req.Req, &sreq)
	if err != nil {
		commResult.Code = lib.RET_CODE_FATAL_CALL
		commResult.Msg =
			fmt.Sprintf("Incorrectly formatted Req: %s!\n", string(req.Req))
		return &commResult
	}
	var sresp serverResponse
	err = json.Unmarshal(resp.Resp, &sresp)
	if err != nil {
		commResult.Code = lib.RET_CODE_ERROR_RESPONSE
		commResult.Msg =
			fmt.Sprintf("Incorrectly formatted Resp: %s!\n", string(resp.Resp))
		return &commResult
	}

	if sresp.StatusCode == 408 {
		commResult.Code = lib.RET_CODE_WARNING_CALL_TIMEOUT
		commResult.Msg =
			fmt.Sprintf("Waring Call Timeout: %s!\n", string(resp.Resp))
		return &commResult
	} else if sresp.StatusCode >= 400 && sresp.StatusCode < 500 {
		commResult.Code = lib.RET_CODE_ERROR_RESPONSE
		commResult.Msg =
			fmt.Sprintf("Response Error: %s!\n", string(resp.Resp))
		return &commResult
	}

	if sresp.StatusCode >= 500 {
		commResult.Code = lib.RET_CODE_ERROR_CALEE
		commResult.Msg =
			fmt.Sprintf("Waring Call Timeout: %s!\n", string(resp.Resp))
		return &commResult
	}

	commResult.Code = lib.RET_CODE_SUCCESS
	commResult.Msg = fmt.Sprintf("Success! (%s)\n ", string(resp.Resp))
	return &commResult
}

func NewGetRequest(url string) lib.Caller {
	return &HttpGETRequest{url: url}
}
