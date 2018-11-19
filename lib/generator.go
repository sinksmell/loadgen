package lib

//TODO define interface Generator

//使用面向接口编程
type Generator interface {
	//启动载荷发生器
	//返回值表示结果
	Start() bool
	//停止
	//
	Stop() bool
	//
	//获取状态
	Status() uint32
	//
	//获取调用计数，获取后会重置该计数
	CallCount() int64
}