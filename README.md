### Generator Load 载荷发生器

>简单的web负载生成器，模拟用户并发地对网站进行测试,可用来测试web应用性能
	
### 名词解释

| 变量名  | 中文  | 解释 |
|:-------------:|:---------------:| :-------------:|
| timeoutNS     | 超时时间(单位NS) |  用来衡量响应是否超时 |
| lps      | 单位时间载荷发生量   |  可以理解为每秒发起的请求次数 |
| concurrency | 并发量        | 可以理解为同时访问的用户数量 |


>计算公式
>
>1e9/lps 可以得到发起请求的时间间隔
>
>>concurrency = timeoutNS/(1e9/lps)+1
>
>上述公式可以得到并发量


### 获取可执行文件
	在release中可以找到对应的系统的二进制文件
### 使用方法

>用法举例
>>手动设置参数
>>>./MyLoadGen  -url http://127.0.0.1:8080 -lps 1000 -timeOut 1000ms -tm 10s
>>
>>使用默认参数(也可以单独设置某一个参数)
>>>./MyLoadGen  
>
>获取帮助
>>./MyLoadGen -help
>>>
>>>

```bash

-> % ./MyLoadGen -help 
Usage of ./MyLoadGen:
  -lps int
        每秒载荷发送量 (default 1000)
  -t string
        测试时长(单位: s) (default "10s")
  -timeOut string
        响应超时时间(单位: ms,s 等) (default "1000ms")
  -url string
        测试地址 (default "http://127.0.0.1:8080")


```

### 手动编译

	需要安装好go语言环境
```bash
go get github.com/astaxie/beego
cd $GOPATH/src
git clone https://github.com/sinksmell/MyLoadGen.git
cd MyLoadGen
go build
```
