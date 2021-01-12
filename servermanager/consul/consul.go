/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package consul

import (
	"context"
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type ConsulInfo struct {
	url        string
	token      string
	healthPort int
	healthType string
	timeout    string
	interval   string
	logout     string
}

type ConsulRegisterInfo struct {
	svrID   string
	svrName string
	svrType string
	svrAddr string
	svrPort int
}

type ConsulService struct {
	exitCh       chan interface{}
	consulInfo   ConsulInfo
	regInfo      ConsulRegisterInfo
	registration *consulapi.AgentServiceRegistration
	client       *consulapi.Client
}

func NewConsulService(info ConsulInfo, regInfo ConsulRegisterInfo) *ConsulService {
	return &ConsulService{
		exitCh:     make(chan interface{}),
		consulInfo: info,
		regInfo:    regInfo,
	}
}

//初始化注册信息
func (s *ConsulService) Open() error {
	config := consulapi.DefaultConfig()
	config.Address = s.consulInfo.url
	if len(s.consulInfo.token) > 0 {
		config.Token = s.consulInfo.token
	} else {
		config.Token = "defaultToken"
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		return err
	}
	s.client = client

	//服务注册信息
	reg := &consulapi.AgentServiceRegistration{
		ID:      strings.ToLower(uuid.NewV4().String()),                // 生成一个唯一当服务ID
		Name:    strings.ToLower(fmt.Sprintf("%s", s.regInfo.svrName)), // 注册服务名
		Tags:    []string{strings.ToLower(s.regInfo.svrType)},          // 标签
		Port:    s.regInfo.svrPort,                                     // 端口号
		Address: s.regInfo.svrAddr,                                     // 所在节点ip地址

	}

	// 健康检测配置信息
	reg.Check = &consulapi.AgentServiceCheck{
		Timeout:                        s.consulInfo.timeout,
		Interval:                       s.consulInfo.interval,
		DeregisterCriticalServiceAfter: s.consulInfo.logout,     // 30秒服务不可达时，注销服务
		Status:                         consulapi.HealthPassing, // 服务启动时，默认正常
	}
	if s.consulInfo.healthType == "http" {
		reg.Check.HTTP = fmt.Sprintf("http://%s:%d%s", reg.Address, s.consulInfo.healthPort, "/health")
		s.RunHealthCheck(reg.Check.HTTP)
	} else {
		reg.Check.TCP = fmt.Sprintf("%s:%d", s.regInfo.svrAddr, s.regInfo.svrPort)
	}
	s.registration = reg

	return nil
}

//退出，关闭健康诊断服务
func (s *ConsulService) Close() {
	if s.exitCh == nil {
		return
	}
	close(s.exitCh)
	s.exitCh = nil
}

func (s *ConsulService) Register() error {
	err := s.client.Agent().ServiceRegister(s.registration)
	if err != nil {
		return err
	}
	return nil
}

func (s *ConsulService) UnRegister() error {
	err := s.client.Agent().ServiceDeregister(s.registration.ID)
	if err != nil {
		return err
	}
	return nil
}

//健康诊断函数
func (s *ConsulService) RunHealthCheck(addr string) error {
	// 实现一个接口类

	uri, err := url.Parse(addr)
	if err != nil {
		return err
	}
	router := http.NewServeMux()

	router.HandleFunc(uri.Path, func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("success"))
	})
	server := &http.Server{
		Addr:         uri.Host,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			fmt.Println(err)
		}
	}()
	//监听退出信号，方便调用http服务退出
	go func() {
		<-s.exitCh
		server.Shutdown(context.Background())
	}()
	return nil
}

//同步版本
type ConsulClient struct {
	url    string
	token  string
	client *consulapi.Client
}

func NewConsulClient(url, token string) *ConsulClient {
	return &ConsulClient{
		url:   url,
		token: token,
	}
}

//根据标签内容随机获取一个服务的ip 端口  Consul无法实时获取到数据的变化通知，只能主动调用接口查看数据是否变化
func (c *ConsulClient) GetAddrs(ServiceType string) (string, int, bool, error) {
	if c.client == nil {
		config := consulapi.DefaultConfig()
		config.Address = c.url
		if len(c.token) > 0 {
			config.Token = c.token
		} else {
			config.Token = "defaultToken"
		}

		client, err := consulapi.NewClient(config)
		if err != nil {
			return "", 0, false, err
		}
		c.client = client
	}

	services, err := c.client.Agent().ServicesWithFilter(ServiceType + " in Tags")
	if err != nil {
		return "", 0, false, err
	}
	if services == nil || len(services) == 0 {
		return "", 0, false, nil
	}
	var serv *consulapi.AgentService
	//go map具有随机性，这里就可以随机选择一个服务地址
	for _, v := range services {
		serv = v
		if serv != nil {
			break
		}
	}
	return serv.Address, serv.Port, true, nil
}

//异步更新版本Consul封装
type addrInfo struct {
	addr string
	port int
}

type ConsulClientASync struct {
	url     string
	token   string
	sType   string
	closeCh chan interface{}
	client  *consulapi.Client
	lock    sync.Mutex
	addrs   []addrInfo
}

func NewConsulClientASync(url, token, sType string) *ConsulClientASync {
	return &ConsulClientASync{
		url:     url,
		token:   token,
		sType:   sType,
		closeCh: make(chan interface{}),
	}
}

func (c *ConsulClientASync) Open(tick time.Duration) error {
	config := consulapi.DefaultConfig()
	config.Address = c.url
	if len(c.token) > 0 {
		config.Token = c.token
	} else {
		config.Token = "defaultToken"
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		return err
	}
	c.client = client
	//开启同步子协程
	go func() {
		for {
			select {
			case <-c.closeCh:
				{
					//收到退出信号，退出更新协程
					return
				}
			default:
				{
					addrs, err := c.getAddrs(c.sType)
					if err != nil {
						//log
					} else {

						c.set(addrs)
					}
				}
			}
			time.Sleep(tick)
		}
	}()

	return nil
}

func (c *ConsulClientASync) Close() {
	if c.closeCh != nil {
		close(c.closeCh)
		c.closeCh = nil
	}

}

func (c *ConsulClientASync) set(addrs *[]addrInfo) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.addrs = *addrs
}

func (c *ConsulClientASync) Get() (string, int, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if len(c.addrs) == 0 {
		return "", 0, false
	}
	index := rand.Intn(len(c.addrs))
	return c.addrs[index].addr, c.addrs[index].port, true
}

//根据标签内容随机获取一个服务的ip 端口  Consul无法实时获取到数据的变化通知，只能主动调用接口查看数据是否变化
func (c *ConsulClientASync) getAddrs(ServiceType string) (*[]addrInfo, error) {
	services, err := c.client.Agent().ServicesWithFilter(ServiceType + " in Tags")
	if err != nil {
		return nil, err
	}
	addrs := []addrInfo{}
	if services == nil || len(services) == 0 {
		return &addrs, nil
	}
	for _, v := range services {
		addrs = append(addrs, addrInfo{
			addr: v.Address,
			port: v.Port,
		})
	}
	return &addrs, nil
}
