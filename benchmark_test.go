package tinyrpc

import (
	"encoding/gob"
	"fmt"
	"github.com/ybtuteng/tinyrpc/example/protocol"
	"log"
	"net"
	"net/rpc"
	"testing"
)

func BenchmarkGoRpc(b *testing.B) {
	cli, err := rpc.DialHTTP("tcp", "127.0.0.1:1234")
	if err != nil {
		fmt.Println("dialing:", err)
	}

	req := protocol.HelloReq{
		Content: "Hello",
	}
	rsp := protocol.HelloRsp{}
	for i := 0; i < b.N; i++ {
		err = cli.Call("HelloServer.SayHello2", req, &rsp)
		if err != nil {
			fmt.Printf("call sayhello error %+v", err)
		}
	}
}

func BenchmarkTinyrpc(b *testing.B) {
	gob.Register(protocol.HelloRsp{})
	gob.Register(protocol.HelloReq{})

	conn, err := net.Dial("tcp", "127.0.0.1:2333")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	cli := NewClient(conn)
	req := protocol.HelloReq{
		Content: "Hello",
	}
	rsp := protocol.HelloRsp{}
	for i := 0; i < b.N; i++ {
		err = cli.Call("SayHello", req, &rsp)
		if err != nil {
			log.Printf("call hellp error %+v", err)
		}
	}
	fmt.Println(rsp.Content)
}
