![github](https://img.shields.io/github/license/gufeijun/hgen) [![Build Status](https://app.travis-ci.com/gufeijun/hgen.svg?branch=master)](https://app.travis-ci.com/gufeijun/hgen)

# 相关

搭配以下RPC库使用：

+ Go：https://github.com/gufeijun/rpch-go
+ C：https://github.com/gufeijun/rpch-c
+ Node：https://github.com/gufeijun/rpch-node

# 介绍

类似于Grpc的跨编程语言RPC框架，此为中立接口描述语言(IDL)的编译器部分。

未采用第三方IDL如ProtoBuf，而是自行设计一套简单语法，此编译器的作用就是对我们语法进行解析，生成目标语言的桩代码，并与上述三个RPC框架接口进行对接。目前已支持Go、C、Node三门语言之间的相互RPC调用。

LL(1)文法，编译器实现中前端是手写递归下降，后端使用Go模板引擎进行代码生成，文法见[bnf.txt](https://github.com/gufeijun/hgen/blob/master/bnf.txt)。

仅为个人学习项目，误用于任何生产环境。

关于RPC框架的协议设计部分见[rpch-go](https://github.com/gufeijun/rpch-go#协议设计)

# IDL语法

新建`math.gfj`文件：

```protobuf
service Math{
	// 两数相加
    uint32 Add(uint32,uint32)	// 可以有多个传入参数，且传入参数可以为基础类型
   	// 两数相减
    int32 Sub(int32,int32)
    // 两数相乘
    int32 Multiply(TwoNum)
    // 两数相除
    Quotient Divide(uint64,uint64)
    void NoReturnFunc(void)		// 服务定义可以无返回值和请求参数
}

message Quotient{
    uint64 Quo	// 商
    uint64 Rem	// 余数
}

message TwoNum{
    int32 A
    int32 B
}

// message可以拥有message成员
message ComplexStruct {
	TwoNum Nums
	Quotient noConcern
}
```

仅两个关键字：

+ message用来定义复合结构。
+ service用来定义服务集合。

风格类似与C语言，相较于ProtoBuf具有更为直观的定义和灵活性。

**基础类型**：int8、uint8、int16、uint16、int32、uint32、int64、uint64、float32、float64、string、void、stream、istream、ostream。stream为流传输类型，仅rpch-go支持，详见[rpch-go](https://github.com/gufeijun/rpch-go)。

# 压测

除了三种语言实现的rpch外，还引入了golang的rpc标准库以及grpc框架来进行横向的对比。

测试案例选择简单的Add服务(两数相加)，本框架可以用两种IDL定义方式实现此Add服务：

> 第一种实现

```go
service Math{
	int32 Add(int32,int32)
}
```

不需要json序列化，数据全部用小端传输即可，效率更高。

> 第二种实现

```go
service Math{
	Response Add(Request)
}
message Response{
	int32 Result
}
message Request{
	int32 A
	int32 B
}
```

因为需要传输message这种复合结构，需要采用序列化，效率低，更能代表普遍的应用场景。google推出grpc只能使用此方式。显然第二种实现更贴近我们平时

的业务场景，所以对我们的框架压测除了采用第一种方式外，也测试了第二种实现。

在不同数量客户端连接下测量吞吐率和延时的时间，当客户端并发数小于等于500时，每个客户端发送10000个rpc请求。当客户端并发数大于500时，为了节省测

试用时，每个客户端发送1000请求。

采用reply-response同步请求方式，在未获取请求的响应之前，不允许再发送下一个请求。所有案例测试多次，取平均值。

测试环境：

+ 系统：Ubuntu 20.04。

+ 终端利用 ulimit -n 65535 命令，将允许打开的文件描述符个数改为最大。

+ CPU：AMD Ryzen 5 3550H 移动端。

+ 内存：16GB。

### 原始数据

rpch-c代表使用c语言实现的rpch框架，rpch-c(json)代表使用了第二种json序列化的方式。

不同框架在不同并发客户端数目下的吞吐率(每秒处理的请求数)：

|      | rpch-c | rpch-c(json) | rpch-go | rpch-go(json) | rpch-node | stdrpc | grpc-go |
| ---- | ------ | ------------ | ------- | ------------- | --------- | ------ | ------- |
| 1    | 9980   | 9320         | 11792   | 9355          | 1039      | 6993   | 3686    |
| 10   | 74239  | 72727        | 91491   | 78740         | 8702      | 71994  | 29420   |
| 100  | 178922 | 144928       | 148456  | 115380        | 12649     | 110803 | 37926   |
| 200  | 167940 | 130387       | 139860  | 111988        | 15759     | 107718 | 39118   |
| 300  | 158587 | 125697       | 134735  | 109601        | 17549     | 104102 | 39425   |
| 500  | 150299 | 119927       | 131144  | 107794        | 17370     | 101161 | 39538   |
| 1000 | 151604 | 120460       | 130371  | 111089        | 16420     | 102162 | 39757   |
| 2000 | 145041 | 1209575      | 129514  | 109922        | 15709     | 99789  | 37434   |
| 5000 | 143661 | 113254       | 123066  | 102303        | 14772     | 90283  | 36342   |

不同框架在不同并发客户端数目下的每个请求的延时(单位ms)：

|      | rpch-c  | rpch-c(json) | rpch-go | rpch-go(json) | rpch-node | stdrpc  | grpc-go  |
| ---- | ------- | ------------ | ------- | ------------- | --------- | ------- | -------- |
| 1    | 0.1002  | 0.1073       | 0.0848  | 0.1069        | 0.9628    | 0.143   | 0.2713   |
| 10   | 0.1347  | 0.1375       | 0.1093  | 0.127         | 1.1491    | 0.1389  | 0.3399   |
| 100  | 0.5589  | 0.69         | 0.6736  | 0.8667        | 7.9056    | 0.9025  | 2.6367   |
| 200  | 1.1909  | 1.5339       | 1.43    | 1.7859        | 12.691    | 1.8567  | 5.1127   |
| 300  | 1.8917  | 2.3867       | 2.2266  | 2.7372        | 17.0947   | 2.8818  | 7.6093   |
| 500  | 3.3267  | 4.1692       | 3.8126  | 4.6385        | 28.785    | 4.9426  | 12.6462  |
| 1000 | 6.5961  | 8.3015       | 7.6704  | 9.0018        | 60.903    | 9.7884  | 25.1527  |
| 2000 | 13.7892 | 16.5323      | 15.4424 | 18.1948       | 127.3164  | 20.0422 | 53.427   |
| 5000 | 34.8042 | 44.1485      | 40.6287 | 48.8744       | 338.4807  | 55.3812 | 137.5815 |

### 吞吐率

![image.png](https://s2.loli.net/2022/03/15/56AbshY1Rkuxm9L.png)

横坐标为客户端并发数，纵坐标为吞吐率即每秒处理的rpc请求数。吞吐率越高越好！

stdrpc代表golang标准库，grpc是google推出的grpc(本例使用Go语言实现rpc服务)框架，包含_json的为上述框架使用第二种实现(采用json序列化)的方式，其他

为使用第一种实现的方式。

rpch-c > rpch-go > rpch-c(json序列化) > rpch-go(json序列化) > stdrpc >> grpc > rpch-node

使用c语言开发的rpch框架效率最高，其次为golang实现。当采用json序列化后，效率会有一定的损失，但依旧优于golang的rpc标准库，grpc的性能很低，仅能

到达rpch-node的两倍。

### 延时

![image.png](https://s2.loli.net/2022/03/15/tczvKSjaNE54iVh.png)

横坐标为并发数，纵坐标为每个请求的延时，单位为毫秒。延时越低越好。

rpch-c < rpch-go < rpch-c(json序列化) < rpch-go(json序列化) < stdrpc << grpc < rpch-node

各框架服务都比较稳定，延时随并发数增长而线性增长，排名和吞吐率指标中一致，c语言的实现性能最优，nodejs实现最差。

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

目前已支持go语言、c语言以及Nodejs，lang参数用于指定语言。dir参数用于指定生成的代码文件存放路径。

