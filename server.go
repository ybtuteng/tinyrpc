package tinyrpc

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"unicode"
	"unicode/utf8"
)

type Server struct {
	name    string
	rcvr    reflect.Value
	typ     reflect.Type
	methods map[string]reflect.Value
	pending chan net.Conn
}

func NewServer() *Server {
	return &Server{methods: make(map[string]reflect.Value), pending: make(chan net.Conn, 20)}
}

func (server *Server) Register(rcvr interface{}) error {
	server.typ = reflect.TypeOf(rcvr)
	server.rcvr = reflect.ValueOf(rcvr)
	serverName := reflect.Indirect(server.rcvr).Type().Name()

	if serverName == "" {
		s := "rpc.Register: no service name for type " + server.typ.String()
		log.Print(s)
		return errors.New(s)
	}

	if !isExported(serverName) {
		s := "rpc.Register: type " + serverName + " is not exported"
		log.Print(s)
		return errors.New(s)
	}

	// register all methods
	server.name = serverName
	server.methods = registerMethods(server.typ, server.rcvr)

	if len(server.methods) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		server.methods = registerMethods(reflect.PtrTo(server.typ), server.rcvr)
		if len(server.methods) != 0 {
			str = "rpc.Register: type " + serverName + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + serverName + " has no exported methods of suitable type"
		}
		log.Print(str)
		return errors.New(str)
	}

	log.Println("all methods registered!")

	return nil
}

func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

func registerMethods(typ reflect.Type, val reflect.Value) map[string]reflect.Value {
	methods := make(map[string]reflect.Value)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		f := val.Method(m)
		mname := method.Name

		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}

		log.Println(method.Type.NumIn())
		for i:=1; i<method.Type.NumIn(); i++ {
			t := method.Type.In(i)
			if t.Kind() == reflect.Ptr { //指针类型获取真正type需要调用Elem
				t = t.Elem()
			}
			//obj := reflect.New(t)
			//obj.Elem().FieldByName("Content").SetString("sb")
			//log.Println(obj.Elem().FieldByName("Content").String())
			//
			//log.Println(method.Type.In(i))
			gob.Register(t)
		}

		for i:=0; i<method.Type.NumOut(); i++ {
			//t := reflect.TypeOf(method.Type.In(i))
			//if t.Kind() == reflect.Ptr { //指针类型获取真正type需要调用Elem
			//	t = t.Elem()
			//}
			//log.Println(method.Type.In(i))
			gob.Register(reflect.New(method.Type.Out(i)))
		}

		methods[mname] = f
		log.Printf("method %s has been registerd", mname)
	}
	return methods
}

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Print("rpc.Serve: accept:", err.Error())
			return
		}
		go server.ServeConn(conn)
	}
}

func (server *Server) ServeConn(conn net.Conn) {
	//log.Printf("conn %s serve", conn.RemoteAddr())
	svrTransport := NewTransClient(conn)
	defer func() { CloseTransClient(conn); conn.Close() }()
	for {
		// read request from client
		//req, err := svrTransport.Receive()
		req := new(Data)

		err := svrTransport.NewReceive(req)
		if err != nil {
			if err != io.EOF {
				log.Printf("read err: %v\n", err)
			}
			//log.Printf("rpc server receive eof error %+v", err)
			return
		}

		// get method by name
		f, ok := server.methods[req.Name]
		if !ok { // if method requested does not exist
			e := fmt.Sprintf("func %s does not exist", req.Name)
			log.Println(e)
			if err = svrTransport.NewSend(Data{Name: req.Name, Err: e}); err != nil {
				log.Printf("transport write err: %v\n", err)
			}
			continue
		}

		if len(req.Args) < 1 {
			e := fmt.Sprintf("empty in args %v", req.Args)
			if err = svrTransport.NewSend(Data{Name: req.Name, Err: e}); err != nil {
				log.Printf("transport write err: %v\n", err)
			}
			continue
		}

		// unpackage request arguments
		inArgs := make([]reflect.Value, 1)
		inArgs[0] = reflect.ValueOf(req.Args[0])

		// invoke requested method
		out := f.Call(inArgs)

		// check out
		if len(out) < 1 {
			e := fmt.Sprintf("empty out args %v", out)
			if err = svrTransport.NewSend(Data{Name: req.Name, Err: e}); err != nil {
				log.Printf("transport write err: %v\n", err)
			}
			continue
		}

		outArgs := make([]interface{}, 1)
		outArgs[0] = out[0].Interface()

		// package error argument
		var e string
		if len(out) > 1 {
			if _, ok := out[1].Interface().(error); !ok {
				e = ""
			} else {
				e = out[1].Interface().(error).Error()
			}
		}

		// send response to client
		err = svrTransport.NewSend(Data{Name: req.Name, Args: outArgs, Err: e})
		if err != nil {
			log.Printf("transport send err: %+v\n", err)
		}
	}
}
