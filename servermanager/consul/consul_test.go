/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package consul

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"strconv"
	"sync"
	"testing"
	"time"
)

//使用开发环境的Consul完成测试
var test_url = "10.0.3.196:8500"
var test_token = ""
var test_serviceType = "test"
var test_localAddr = "10.0.0.85"

//正常退出流程
func testConsulService(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(6)
	for i := 0; i < 5; i++ {
		go func(i int) {
			defer wg.Done()
			consulSvr := NewConsulService(ConsulInfo{
				url:        test_url,
				token:      test_token,
				healthPort: 1000 + i,
				healthType: "http",
				timeout:    "1s",
				interval:   "15s",
				logout:     "30s",
			}, ConsulRegisterInfo{
				svrID:   uuid.NewV4().String(),
				svrName: "test",
				svrType: test_serviceType,
				svrAddr: test_localAddr,
				svrPort: i,
			})
			err := consulSvr.Open()
			if err != nil {
				fmt.Println(err)
				return
			}
			defer consulSvr.Close()

			err = consulSvr.Register()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("register OK")
			defer consulSvr.UnRegister()
			time.Sleep(time.Minute * 10)
		}(i)
	}
	//等待注册完成
	time.Sleep(time.Second)
	go func() {
		defer wg.Done()
		for {
			consulCli := NewConsulClient(test_url, test_token)
			addr, port, flag, err := consulCli.GetAddrs(test_serviceType)
			if err != nil {
				fmt.Println("consulCli.GetAddrs errors:", err)
				return
			} else if !flag {
				fmt.Println("consulCli.GetAddrs  Not Service")
			} else {
				fmt.Println("GetService Addr :", addr, " port:", port)
			}
			time.Sleep(time.Second)
		}

	}()
	wg.Wait()
}

//注册服务异常退出时的测试
func testConsulServiceTimout(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(6)
	for i := 0; i < 5; i++ {
		go func(i int) {
			defer wg.Done()
			consulSvr := NewConsulService(ConsulInfo{
				url:        test_url,
				token:      test_token,
				healthPort: 1000 + i,
				healthType: "http",
				timeout:    "1s",
				interval:   "5s",
				logout:     "15s",
			}, ConsulRegisterInfo{
				svrID:   uuid.NewV4().String(),
				svrName: "test",
				svrType: test_serviceType,
				svrAddr: test_localAddr,
				svrPort: i,
			})
			err := consulSvr.Open()
			if err != nil {
				fmt.Println(err)
				return
			}
			defer consulSvr.Close()

			err = consulSvr.Register()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("register OK")
			//模拟异常退出，不向consul注销信息
			//defer consulSvr.UnRegister()
			time.Sleep(time.Second * 10)
		}(i)
	}
	//等待注册完成
	time.Sleep(time.Second)
	go func() {
		defer wg.Done()
		for {
			consulCli := NewConsulClient(test_url, test_token)
			addr, port, flag, err := consulCli.GetAddrs(test_serviceType)
			if err != nil {
				fmt.Println("consulCli.GetAddrs errors:", err)
				return
			} else if !flag {
				fmt.Println("consulCli.GetAddrs  Not Service")
			} else {
				fmt.Println("GetService Addr :", addr, " port:", port)
			}
			time.Sleep(time.Second)
		}

	}()
	wg.Wait()
}

//异步Consul测试
func TestConsulServiceAsync(t *testing.T) {

	wg := sync.WaitGroup{}
	wg.Add(6)
	for i := 0; i < 5; i++ {
		go func(i int) {
			defer wg.Done()
			consulSvr := NewConsulService(ConsulInfo{
				url:        test_url,
				token:      test_token,
				healthPort: 1000 + i,
				healthType: "http",
				timeout:    "1s",
				interval:   "5s",
				logout:     "15s",
			}, ConsulRegisterInfo{
				svrID:   uuid.NewV4().String(),
				svrName: "test",
				svrType: test_serviceType,
				svrAddr: test_localAddr,
				svrPort: i,
			})
			err := consulSvr.Open()
			if err != nil {
				fmt.Println(err)
				return
			}
			defer consulSvr.Close()

			err = consulSvr.Register()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("register OK")
			defer fmt.Println("unregister OK prot:" + strconv.Itoa(i))
			defer consulSvr.UnRegister()
			time.Sleep((time.Duration)((int)(time.Second*10) * (i + 1)))
		}(i)
	}

	//等待注册完成
	time.Sleep(time.Second)
	go func() {
		defer wg.Done()
		consulCli := NewConsulClientASync(test_url, test_token, test_serviceType)

		err := consulCli.Open(time.Second * 10)
		if err != nil {
			fmt.Println("consulCli.Open errors:", err)
			return
		}
		//内部开启了子协程，手动退出
		defer consulCli.Close()
		//等待客户端初次获取完成
		time.Sleep(time.Second * 3)
		//随着注册协程依次注销，能获取的注册信息范围会越来越小，直到最后没有注册服务存在，退出执行
		for {
			addr, port, flag := consulCli.Get()
			if !flag {
				fmt.Println("consulCli.Get 获取失败")
				return
			} else {
				fmt.Println("GetService Addr :", addr, " port:", port)
			}
			time.Sleep(time.Second)
		}

	}()
	wg.Wait()
}
