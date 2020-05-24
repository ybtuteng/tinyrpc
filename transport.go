package tinyrpc

import (
	"bufio"
	"encoding/binary"
	"encoding/gob"
	"io"
	"log"
	"net"
	"sync"
)

var transports sync.Map

// TransClient struct
type TransClient struct {
	conn net.Conn
	enc  *gob.Encoder
	dec  *gob.Decoder
	encBuf *bufio.Writer
}

// NewTransClient creates a transport
func NewTransClient(conn net.Conn) *TransClient {
	buf := bufio.NewWriter(conn)
	return &TransClient{conn: conn, enc: gob.NewEncoder(buf), dec: gob.NewDecoder(conn), encBuf: buf}
}

//gob有个坑，一个gob实例接收一个输入输出流，为了提高效率，流里面的第一个元素包含结构体的完整信息，之后的只包含必要信息
//如果一条连接创建了多个gob实例，每个实例的第一个元素都会发送完整信息给服务端，服务端收到多个元素的完整信息就会报错
//所以一条连接只能和一个gob实例绑定，确保不会发送重复的元信息
func GetTransClient(conn net.Conn) *TransClient {
	v, ok :=transports.Load(conn)
	if !ok {
		v = NewTransClient(conn)
		transports.Store(conn, v)
	}

	return v.(*TransClient)
}

//当连接关闭的时候，删掉map里的元素
func CloseTransClient(conn net.Conn) {
	_, ok :=transports.Load(conn)
	if ok {
		transports.Delete(conn)
	}
}

// Send data
func (t *TransClient) Send(req Data) error {
	b, err := encode(req)
	if err != nil {
		log.Printf("encode req error %+v", err)
		return err
	}
	buf := make([]byte, 4+len(b))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(b))) // Set length in Header
	copy(buf[4:], b)                                    // Set Data
	_, err = t.conn.Write(buf)
	return err
}

// Receive data
func (t *TransClient) Receive() (Data, error) {
	header := make([]byte, 4)
	_, err := io.ReadFull(t.conn, header)
	if err != nil {
		return Data{}, err
	}

	//startT := time.Now()
	dataLen := binary.BigEndian.Uint32(header) // Read length in Header
	data := make([]byte, dataLen)              // Read Data of length
	_, err = io.ReadFull(t.conn, data)
	if err != nil {
		log.Printf("read data error %+v", err)
		return Data{}, err
	}

	rsp, err := decode(data) // Decode rsp from bytes
	if err != nil {
		log.Printf("gob decode error %+v", err)
	}
	return rsp, err
}

func (t *TransClient) NewReceive(rsp interface{}) error {
	return t.dec.Decode(rsp)
}

func (t *TransClient) NewSend(req Data) (err error) {
	if err = t.enc.Encode(req); err != nil {
		if t.encBuf.Flush() == nil {
			// Gob couldn't encode the header. Should not happen, so if it does,
			// shut down the connection to signal that the connection is broken.
			log.Println("rpc: gob error encoding response:", err)
		}
		return
	}
	return t.encBuf.Flush()
}
