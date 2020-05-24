package main

import (
	"encoding/gob"
	"fmt"
	"github.com/ybtuteng/tinyrpc"
	"github.com/ybtuteng/tinyrpc/example/protocol"
	"log"
	"net"
)

func main() {
	gob.Register(protocol.HelloRsp{})
	gob.Register(protocol.HelloReq{})

	conn, err := net.Dial("tcp", "127.0.0.1:2333")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	cli := tinyrpc.NewClient(conn)
	req := protocol.HelloReq{
		Content: "Hello",
	}
	rsp := protocol.HelloRsp{}
	err = cli.Call("SayHello", req, &rsp)
	if err != nil {
		log.Printf("call hellp error %+v", err)
	}
	fmt.Println(rsp.Content)
}
