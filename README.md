![github](https://img.shields.io/github/license/gufeijun/hgen) [![Build Status](https://app.travis-ci.com/gufeijun/hgen.svg?branch=master)](https://app.travis-ci.com/gufeijun/hgen)

# 介绍

hgen是分布式系统与中间件课程的大作业(自己开发一个rpc框架)的编译器部分。

不同编程语言有不同的语法规则，甚至有的语言有自己独有的序列化方式，如golang的gob、python的pickle等。如果想开发一个跨语言的RPC框架，就必须屏蔽掉不同语言之间的区别，提供一种统一的网络协议格式、对象序列化方法以及接口描述方式。

因此IDL(接口描述语言)的概念应运而生，比如google推出的protobuf。IDL就需要做到统一描述我们欲提供的服务接口，让所有语言都遵循接口定义的规则，由编译器完成IDL到某个具体语言的本土化编译工作。

本着造轮子学习的态度，我并没有为自己的框架选择已经开源成熟的IDL，而是自己设计了一个简单的IDL语言，并为其开发了一个简陋的编译器，也就是本项目hgen，目前已经完成了由IDL编译到golang的工作。编译器前端使用有限状态机，后端采用了golang的模板引擎。目前还未给自己设计的IDL语言命名，暂且每个IDL文件以gfj(作者名字缩写)为文件后缀。

目前hgen生成的代码和其服务的rpc框架rpch存在极高的耦合，后期应该让各个功能模块细度更高，达到解耦。

# IDL语法

### quickstart

新建math.gfj文件：

```protobuf
service Math{
    int32 Add(int32,int32)
}
```

我们定义了一个接口服务Math，这个服务目前仅提供一个Add方法。

利用hgen编译：`hgen -dir gfj -lang go math.gfj`即可在gfj目录下生成golang语言的源码文件`math.rpch.go`：

```go
// This is code generated by hgen. DO NOT EDIT!!!
// hgen version: v0.1.0
// source: math.gfj

package gfj

import (
    rpch "github.com/gufeijun/rpch-go"
)

type MathService interface{
	Add(int32, int32) (int32, error)
}

func RegisterMathService(impl MathService, svr *rpch.Server) {
	methods := map[string]*rpch.MethodDesc {
        "Add": rpch.BuildMethodDesc(impl, "Add", "int32"),
	}
	service := &rpch.Service{
		Impl:    impl,
        Name:    "Math",
		Methods: methods,
	}
	svr.Register(service)
}

type MathServiceClient struct{
    client *rpch.Client
}

func NewMathServiceClient(client *rpch.Client) *MathServiceClient {
    return &MathServiceClient{
		client: client,
	}
}

func (c *MathServiceClient) Add(arg1 int32, arg2 int32) (res int32, err error) {
    resp, err := c.client.Call("Math", "Add",
		&rpch.RequestArg{
            TypeKind: 0,
            TypeName: "int32",
            Data:     arg1,
		},
		&rpch.RequestArg{
            TypeKind: 0,
            TypeName: "int32",
            Data:     arg2,
		})
	if resp == nil {
		return
	}
	return resp.(int32),err
}
```

在该生成的文件中定义了我们服务的接口以及注册服务的函数，除此之外还定义了客户端调用的方法。我们只需要为`MathService`实现具体的Add方法就可以快速实现rpc服务端以及客户端：

```go
package main

import (
	"example/quickstart/gfj"
	"fmt"
	"log"
	"time"

	rpch "github.com/gufeijun/rpch-go"
)

type mathService struct{}

func (*mathService) Add(a int32, b int32) (int32, error) {
	return a + b, nil
}

func startServer() {
	svr := rpch.NewServer()
	gfj.RegisterMathService(new(mathService), svr)
	panic(svr.ListenAndServe("127.0.0.1:8080"))
}

func main() {
	go startServer()
	time.Sleep(time.Second)

	//客户端
	conn, err := rpch.Dial("127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
    defer conn.Close()
	client := gfj.NewMathServiceClient(conn)
	result, err := client.Add(2, 3)
	if err != nil {
		panic(err)
	}
	if result != 5 {
		log.Panicf("want %d, but got %d\n", 5, result)
	}
	fmt.Println("test success!")
}
```

详细实现代码见[链接](https://github.com/gufeijun/rpch-go/tree/master/examples/quickstart)。

### 内置类型

IDL内置了丰富的基本类型：int8、uint8、int16、uint16、int32、uint32、int64、uint64、float32、float64、string、void、stream、istream、ostream。

大部分类型比较好理解，主要讲讲几个特殊的类型。

其中void代表无返回值或者无参数，当接口方法的参数是void，则代表这个方法不需要传入参数。当接口方法的返回值为void，则代表这个方法没有返回值。这点很类似于c语言。

stream类型是我的rpc框架的原创类型，代表了一个流。istream是读取流，对应golang的io.Reader；ostream是写入流，对应golang的io.Writer；stream是读写流，对应golang的io.ReadWriter。不论是哪种stream都可以作为参数传入方法，或者返回参数返回。比如服务端open一个文件得到一个stream流，它可以直接将此stream流返回给客户端，客户端可以如同读写本地文件一样对服务器的这个文件进行读写操作。当然客户端拿到的是经过封装后的，客户端看似在直接读取服务端文件，实际上底层还是在读写tcp连接，只是因为封装给你屏蔽了传输细节，你难以感知而已。

流的实现用到了http协议中的chunk编码。具体使用案例：

比如我们定义IDL文件：

```protobuf
service File{
    stream OpenFile(string)
}
```

服务端实现该接口，其功能就是把这个文件流返回给客户端：

```go
func (*fileService) OpenFile(filepath string) (stream io.ReadWriter, onFinish func(), err error) {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return
	}
	return file, func() {
		file.Close()
	}, nil
}
```

客户端调用：

```go
func readSomething(client *gfj.FileServiceClient) error {
	file, err := client.OpenFile("test.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	io.Copy(os.Stdout, file)
	return err
}

func writeSomething(client *gfj.FileServiceClient) error {
	file, err := client.OpenFile("test.txt")
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write([]byte("hello world\n"))
	if err != nil {
	}
	return err
}
```

FileServiceClient是hgen编译器自动生成的类型，我们可以使用NewFileServiceClient就可以将rpch.Client类型转化为此类型，这些代码这里省略。

客户端调用的OpenFile方法隐藏了底层的网络传输细节，客户端并无法感知我们读写的file是一个远在服务器上的文件，而感觉操纵的是一个本地文件。

具体使用可以查看：[链接](https://github.com/gufeijun/rpch-go/tree/master/examples/fileserver)。

###  关键字

目前仅存在两个关键字：message以及service。

service在前面已经有了初识，它用于定义我们的服务接口，服务端代码对照接口的每个方法予以具体实现，客户端根据这个接口的定义进行远程调用。

message用于定义我们通信过程中传输的参数，案例如下 :

```protobuf
service Math{
    uint32 Add(TwoNum)
    Quotient Divide(uint32,uint32)
    //void PrintServiceName(void)
}

message Quotient{
    uint64 Quo	//商
    uint64 Rem	//余数
}

message TwoNum{
    int32 A
    int32 B
}
```

我们定义了两个复合结构Quotient以及TwoNum，复合结构既可作为传入参数又可以作为返回值。除此之外，复合结构内部成员也可以是复合结构，没有限制规则。

### 注意事项

#### 接口方法的定义

1. 对于Go语言来说，最好将服务名、复合结构名以及复合结构成员首字母大小表示，否则你只能在所生成的代码文件包下编写对应的实现，会带来一定麻烦。
2. 多个service尽量放在不同的.gfj后缀的IDL文件中。
3. 如果方法无返回值，则让返回值为void；如果方法不需要传入参数，则必须填入void显示地表示该方法无参数。如`void Print(void)`，该方法就无参数且无返回值。

# 安装与使用

两种安装方式：

+ 源码编译安装。

  ```shell
  # for linux
  git clone github.com/gufeijun/hgen
  cd hgen
  go build -o hgen main.go
  ```

  即可生成名叫hgen的可执行文件。

+ 二进制下载：[releases](https://github.com/gufeijun/hgen/releases)

请自行配置PATH。

使用：

```
Usage of hgen:
	hgen [options] <IDLfiles...>
options：
  -dir string
    	the dirpath where the generated source code files will be placed (default "gfj")
  -lang string
    	the target languege the IDL will be compliled to (default "c")
```

目前已支持go语言以及c语言，lang参数用于指定语言。dir参数用于指定生成的代码文件存放路径，默认为gfj。

