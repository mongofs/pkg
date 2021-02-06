/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package rabbitMQ

import (
	"fmt"
	"github.com/streadway/amqp"
	"time"
	"os"
	"log"
	"context"
)

var (
	defaultQueueName    = "default_queue"
	defaultExchangeName = "default_exchange"
	defaultExchangeType = "fanout"
	defaultKey          = ""
)


type mq struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	host         string
	user         string
	password     string
	queueName    string //队列名称
	exchange     string //交换机
	exchangeType string // 交换机类型
	key          string //key
	mqurl        string //连接信息
}

type MqOption func(r *mq)

func NewRabbitMQ(opts ...MqOption) (*mq, error) {
	defaultOption := &mq{
		queueName:    defaultQueueName,
		exchange:     defaultExchangeName,
		key:          defaultKey,
		exchangeType: defaultExchangeType,
	}
	for _, opt := range opts {
		opt(defaultOption)
	}
	defaultOption.mqurl = fmt.Sprintf("amqp://%v:%v@%v/", defaultOption.user, defaultOption.password, defaultOption.host)
	fmt.Println(defaultOption.mqurl)
	var err error
	defaultOption.conn, err = amqp.Dial(defaultOption.mqurl)
	if err != nil {
		return nil, err
	}
	defaultOption.channel, err = defaultOption.conn.Channel()
	if err != nil {
		return nil, err
	}
	return defaultOption, nil
}

const (
	FANOUT = 1
	HEADER = 2
	TOPIC  = 3
	DIRECT = 4
)

var ExchangeType = map[int]string{
	FANOUT: "fanout",
	HEADER: "header",
	TOPIC:  "topic",
	DIRECT: "direct",
}

func SetExchangeType(exchangeType int) MqOption {
	return func(r *mq) {
		switch exchangeType {
		case FANOUT:
			r.exchangeType = ExchangeType[FANOUT]
		case HEADER:
			r.exchangeType = ExchangeType[HEADER]
		case TOPIC:
			r.exchangeType = ExchangeType[TOPIC]
		case DIRECT:
			r.exchangeType = ExchangeType[DIRECT]
		default:
			r.exchangeType = ExchangeType[FANOUT]
		}
	}
}

func SetQueueName(queue string) MqOption {
	return func(r *mq) {
		r.queueName = queue
	}
}

func SetExchangeName(exchange string) MqOption {
	return func(r *mq) {
		r.exchange = exchange
	}
}

func SetKey(key string) MqOption {
	return func(r *mq) {
		r.key = key
	}
}

func SetUser(user string) MqOption {
	return func(r *mq) {
		r.user = user
	}
}

func SetPassword(password string) MqOption {
	return func(r *mq) {
		r.password = password
	}
}

func SetHost(host string) MqOption {
	return func(r *mq) {
		r.host = host
	}
}

func (r *mq) Destroy() {
	r.channel.Close()
	r.conn.Close()
}

// 订阅来自作为消费者订阅来自channel的信息
// 第一个channe 收消息 收取其中的消息使用range message.body
// 第二个参数监听和tcp连接
func (r *mq) ReceiveSub() (<-chan amqp.Delivery, chan *amqp.Error, error) {
	cc := make(chan *amqp.Error)

	//1.创建交换机，如果消息发送方没有创建则需要我们进行发送
	err := r.channel.ExchangeDeclare(
		r.exchange,
		r.exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	//2.试探性创建队列,如果存在则不声明
	q, err := r.channel.QueueDeclare(
		r.queueName,
		false, // 消息持久化
		false, // 是否自动删除
		false, // 是否独享数据，排他
		false, //
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	//3.绑定队列到exchange中，在pub/sub模式下,这里的key要为空
	err = r.channel.QueueBind(
		q.Name,
		"",
		r.exchange,
		false,
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	//4.消费消息
	messages, err := r.channel.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, nil, err
	}
	// 注册断开重连的参数
	r.channel.NotifyClose(cc)

	return messages, cc, nil
}

//	这里是推送消息
func (r *mq) PublishPub(message string) error {
	err := r.channel.ExchangeDeclare(
		r.exchange,
		r.exchangeType,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}
	err = r.channel.Publish(
		r.exchange,
		"",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	return err
}

type Wrapper struct {
	dataChan chan []byte
	mq       *mq
	options  []MqOption
}

func NewWrapper(buffer int, opts ...MqOption) (*Wrapper, error) {
	r := &Wrapper{
		dataChan: make(chan []byte, buffer),
		options:  opts,
	}
	r.mq = createMq(r.options...)
	return r, nil
}

//  创建mq连接
func createMq(opts ...MqOption) *mq {
	rabbitMQ, err := NewRabbitMQ(opts...)
	if err != nil || rabbitMQ == nil {
		time.Sleep(time.Second)
		return createMq(opts...)
	}
	return rabbitMQ
}

// 在调用receive之前需要调用此方法
func (r *Wrapper) MonitorAndStarReceiver() {
	if l := len(r.options); l == 0 {
		panic("there is no connection param")
	}
	if r.mq == nil {
		r.mq = createMq(r.options...)
	}

	dataCh, chClose, err := r.mq.ReceiveSub()
	if err != nil {
		// recreate connection
		r.mq = createMq(r.options...)
		r.MonitorAndStarReceiver()
		return
	}
	go func() {
		t := time.NewTicker(3 * time.Second)
	label:
		for {
			select {
			case data := <-dataCh:
				r.dataChan <- data.Body
			case  <-chClose:
				break label
			case <-t.C:
				fmt.Println("Rabbitmq : this thread is alive")
			}
		}
		// receiver errors and retry connection rabbitMQ
		r.mq = createMq(r.options...)
		r.MonitorAndStarReceiver()
	}()

}

//调用此方法可以获取到接收rabbitmq的channel
func (r *Wrapper) Receive() <-chan []byte {
	return r.dataChan
}

//调用此方法可以推送消息
func (r *Wrapper) Push(msg string) error {
	if r.mq == nil {
		r.mq = createMq(r.options...)
	}
	return r.mq.PublishPub(msg)
}

func (r *Wrapper) Close() {
	r.mq.Destroy()
}

func (r *Wrapper) Start(ctx context.Context) error {
	sigSuc := make(chan struct{})
	sigErr := make(chan error)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				if temerr, ok := err.(*os.PathError); ok {
					sigErr <- temerr
				} else {
					log.Fatal(err)
				}
			}
		}()
		sigSuc <- struct{}{}
	}()

	select {
	case <-sigSuc:
		fmt.Println("redis-sentinal : start service success")
		return nil
	case err := <-sigErr:
		return err
	case <-ctx.Done():
		return fmt.Errorf("service starting is timeout")
	}
}

func (r *Wrapper) Stop(ctx context.Context) error {
	r.Close()
	return nil
}