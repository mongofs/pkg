/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package rabbitMQ

import (
	"fmt"
	"testing"
)

// 测试建立接受消息
// 基于连接的可能出现rabbitmq宕机 等现象，或者网络错误需要断开重连
func TestReceive(t *testing.T) {

	receive,err:= NewWrapper(10,
		SetUser("guest"),
		SetPassword("guest"),
		SetHost("127.0.0.1:5672"),
		SetExchangeName("demo"),
		SetQueueName("demo"))
	if err != nil{
		 panic(err)
	}
	// 将监控断开重连go出去
	go receive.MonitorAndStarReceiver()
	// 用户只关心这里就好
	res:=receive.Receive()
	for v:=range res{
		fmt.Printf("收到消息：%v \n",v)
	}
}


// 测试推送消息
// 如果是持续不断的推送就不要把连接关闭，
func Test_PublishPub(t *testing.T) {
	wrapper,err:= NewWrapper(10,
		SetUser("guest"),
		SetPassword("guest"),
		SetHost("127.0.0.1:5672"),
		SetExchangeName("demo"),
		SetQueueName("demo"))

	if err != nil{
		panic(err)
	}
	defer wrapper.Close()
	wrapper.Push("hahhaha")
}