package main

import (
	"encoding/gob"
	"fmt"
	"github.com/ybtuteng/tinyrpc"
	"github.com/ybtuteng/tinyrpc/example/protocol"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type HelloServer struct {

}

func (hs *HelloServer) SayHello(req protocol.HelloReq) (protocol.HelloRsp, error){
	fmt.Println(req)
	time.Sleep(time.Microsecond * 30)
	return protocol.HelloRsp{Content: "hello too"}, nil
}

func (hs *HelloServer) SayHello2(req *protocol.HelloReq, rsp *protocol.HelloRsp) error{
	fmt.Println(req)
	rsp.Content = "hello too"
	time.Sleep(time.Microsecond * 30)
	return nil
}

func main() {
	gob.Register(protocol.HelloRsp{})
	gob.Register(protocol.HelloReq{})

	addr := "0.0.0.0:2333"
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("listen on %s err: %v\n", addr, err)
		return
	}
	srv := tinyrpc.NewServer()
	query := HelloServer{}
	srv.Register(&query)
	log.Println("service is running")
	go goRpcServer()
	srv.Accept(l)
}

func goRpcServer() {
	query := new(HelloServer)
	rpc.Register(query)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}
