package request

import (
	. "MyLoadGen/lib"
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/astaxie/beego/logs"
	"math/rand"
	"net"
	"strconv"
	"sync/atomic"
	"time"
)

const (
	DELIM = '\n' // 分隔符。
)

// operators 代表操作符切片。
var operators = []string{"+", "-", "*", "/"}

// TCPComm 表示TCP通讯器的结构。
type TCPComm struct {
	addr string
}

// ServerReq 表示服务器请求的结构。
type ServerReq struct {
	ID       int64  `json:"id"`
	Operands []int  `json:"numbers"`
	Operator string `json:"operator"`
}

// ServerResp 表示服务器响应的结构。
type ServerResp struct {
	ID      int64  `json:"id"`
	Formula string `json:"formula"`
	Result  int    `json:"result"`
	Err     error  `json:"err"`
}

var logger = logs.NewLogger(10000) // 创建一个日志记录器，参数为缓冲区的大小

func init() {
	logger.SetLogger("console", "")  // 设置日志记录方式：控制台记录
	logger.SetLevel(logs.LevelDebug) // 设置日志写入缓冲区的等级：Debug级别（最低级别，所以所有log都会输入到缓冲区）
	logger.EnableFuncCallDepth(true) // 输出log时能显示输出文件名和行号（非必须）
}

// NewTCPComm 会新建一个TCP通讯器。
func NewTCPComm(addr string) Caller {
	return &TCPComm{addr: addr}
}

// BuildReq 会构建一个请求。
func (comm *TCPComm) BuildReq() RawReq {
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

// Call 会发起一次通讯。
func (comm *TCPComm) Caller(req []byte, timeoutNS time.Duration) ([]byte, error) {
	conn, err := net.DialTimeout("tcp", comm.addr, timeoutNS)
	if err != nil {
		return nil, err
	}
	_, err = write(conn, req, DELIM)
	if err != nil {
		return nil, err
	}
	return read(conn, DELIM)
}

// CheckResp 会检查响应内容。
func (comm *TCPComm) CheckResp(rawReq RawReq, rawResp RawResp) *CallResult {
	var commResult CallResult
	commResult.ID = rawResp.ID
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

// read 会从连接中读数据直到遇到参数delim代表的字节。
func read(conn net.Conn, delim byte) ([]byte, error) {
	readBytes := make([]byte, 1)
	var buffer bytes.Buffer
	for {
		_, err := conn.Read(readBytes)
		if err != nil {
			return nil, err
		}
		readByte := readBytes[0]
		if readByte == delim {
			break
		}
		buffer.WriteByte(readByte)
	}
	return buffer.Bytes(), nil
}

// write 会向连接写数据，并在最后追加参数delim代表的字节。
func write(conn net.Conn, content []byte, delim byte) (int, error) {
	writer := bufio.NewWriter(conn)
	n, err := writer.Write(content)
	if err == nil {
		writer.WriteByte(delim)
	}
	if err == nil {
		err = writer.Flush()
	}
	return n, err
}

func op(operands []int, operator string) int {
	var result int
	switch {
	case operator == "+":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result += v
			}
		}
	case operator == "-":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result -= v
			}
		}
	case operator == "*":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result *= v
			}
		}
	case operator == "/":
		for _, v := range operands {
			if result == 0 {
				result = v
			} else {
				result /= v
			}
		}
	}
	return result
}

// genFormula 会根据参数生成字符串形式的公式。
func genFormula(operands []int, operator string, result int, equal bool) string {
	var buff bytes.Buffer
	n := len(operands)
	for i := 0; i < n; i++ {
		if i > 0 {
			buff.WriteString(" ")
			buff.WriteString(operator)
			buff.WriteString(" ")
		}

		buff.WriteString(strconv.Itoa(operands[i]))
	}
	if equal {
		buff.WriteString(" = ")
	} else {
		buff.WriteString(" != ")
	}
	buff.WriteString(strconv.Itoa(result))
	return buff.String()
}

// reqHandler 会把参数sresp代表的请求转换为数据并发送给连接。
func reqHandler(conn net.Conn) {
	var errMsg string
	var sresp ServerResp
	req, err := read(conn, DELIM)
	if err != nil {
		errMsg = fmt.Sprintf("Server: Req Read Error: %s", err)
	} else {
		var sreq ServerReq
		err := json.Unmarshal(req, &sreq)
		if err != nil {
			errMsg = fmt.Sprintf("Server: Req Unmarshal Error: %s", err)
		} else {
			sresp.ID = sreq.ID
			sresp.Result = op(sreq.Operands, sreq.Operator)
			sresp.Formula =
				genFormula(sreq.Operands, sreq.Operator, sresp.Result, true)
		}
	}
	if errMsg != "" {
		sresp.Err = errors.New(errMsg)
	}
	bytes, err := json.Marshal(sresp)
	if err != nil {
		logger.Emergency("Server: Resp Marshal Error: %s", err)
	}
	_, err = write(conn, bytes, DELIM)
	if err != nil {
		logger.Emergency("Server: Resp Write error: %s", err)
	}
}

// TCPServer 表示基于TCP协议的服务器。
type TCPServer struct {
	listener net.Listener
	active   uint32 // 0-未激活；1-已激活。
}

// NewTCPServer 会新建一个基于TCP协议的服务器。
func NewTCPServer() *TCPServer {
	return &TCPServer{}
}

// init 会初始化服务器。
func (server *TCPServer) init(addr string) error {
	if !atomic.CompareAndSwapUint32(&server.active, 0, 1) {
		return nil
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		atomic.StoreUint32(&server.active, 0)
		return err
	}
	server.listener = ln
	return nil
}

// Listen 会启动对指定网络地址的监听。
func (server *TCPServer) Listen(addr string) error {
	err := server.init(addr)
	if err != nil {
		return err
	}
	go func() {
		for {
			if atomic.LoadUint32(&server.active) != 1 {
				break
			}
			conn, err := server.listener.Accept()
			if err != nil {
				if atomic.LoadUint32(&server.active) == 1 {
					logger.Emergency("Server: Request Acception Error: %s\n", err)
				} else {
					logger.Warning("Server: Broken acception because of closed network connection.")
				}
				continue
			}
			go reqHandler(conn)
		}
	}()
	return nil
}

// Close 会关闭服务器。
func (server *TCPServer) Close() bool {
	if !atomic.CompareAndSwapUint32(&server.active, 1, 0) {
		return false
	}
	server.listener.Close()
	return true
}
