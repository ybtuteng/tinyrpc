一个简易的rpc框架
* 用法参考golang内置库net/rpc，客户端调用示例如下
```
func main() {
    gob.Register(ResponseQueryUser{})
	addr := "0.0.0.0:2333"
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("dial error: %v\n", err)
	}
	cli := myrpc.NewClient(conn)
	reply := ResponseQueryUser{}

	err = cli.Call("QueryUser", 1, &reply)
	if err != nil {
		log.Printf("query error: %v\n", err)
	} 
}
```
* 网络传输基于tcp协议，序列化方法采用内置库gob
* codec编码解码方法使用简单的len+data拼接
* 整体性能与go rpc相当

机型：macbook pro 2017 i7 16G

|测试接口|  协程数   | 用户数  | qps |
|----|  ----  | ----  | ----    |
|myrpc|200|random|9000|
|gorpc|200|random|10000|