package tinyrpc

import (
	"errors"
	"log"
	"net"
	"reflect"
)

// Client struct
type Client struct {
	conn net.Conn
	//cliTransport *TransClient
}

// NewClient creates a new client
func NewClient(conn net.Conn) *Client {
	return &Client{conn: conn}
}

func (c *Client) GetConn() net.Conn {
	return c.conn
}

func (c *Client) Call(name string, arg, reply interface{}) error {
	//cliTransport := NewTransClient(c.conn)
	cliTransport := GetTransClient(c.conn)

	errorHandler := func(err error) error {
		log.Printf("error %+v", err)
		return err
	}

	// send request to server
	args := []interface{}{arg}
	err := cliTransport.NewSend(Data{Name: name, Args: args})
	if err != nil { // local network error or encode error
		log.Printf("send error %+v", err)
		return errorHandler(err)
	}

	// receive response from server
	//rsp, err := cliTransport.Receive()
	rsp := &Data{}
	err = cliTransport.NewReceive(rsp)
	if err != nil { // local network error or decode error
		log.Printf("receive error %+v", err)
		return errorHandler(err)
	}

	if rsp.Err != "" { // remote server error
		log.Printf("receive error %s", rsp.Err)
		return errorHandler(errors.New(rsp.Err))
	}

	if len(rsp.Args) < 1 {
		return errorHandler(errors.New("empty response"))
	}

	// set reply
	sPtr := reflect.ValueOf(reply)
	sPtr.Elem().Set(reflect.ValueOf(rsp.Args[0]))

	return nil
}
