/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/FZambia/sentinel"
	"github.com/gomodule/redigo/redis"
	"log"
	"math/rand"
	"os"
	"strings"
	//"strconv"
	"time"
)

//哨兵版redis
//后期可以为其增加一些操作的封装或者替换其他的redis库
type RedisDriver struct {
	dbName     string
	addrs      []string
	auth       bool
	passWord   string
	masterPool *redis.Pool
	slavePool  *redis.Pool
}

func NewRedisDriver(dbName, password string, addrs []string, auth bool) *RedisDriver {
	return &RedisDriver{
		dbName:   dbName,
		addrs:    addrs,
		auth:     auth,
		passWord: password,
	}
}

func (driver *RedisDriver) Start(ctx context.Context) error {
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

		sntnl := &sentinel.Sentinel{
			Addrs:      driver.addrs,
			MasterName: driver.dbName,
			Dial: func(addr string) (redis.Conn, error) {
				timeout := 500 * time.Millisecond
				c, err := redis.DialTimeout("tcp", addr, timeout, timeout, timeout)
				if err != nil {
					return nil, err
				}
				return c, nil
			},
		}

		driver.masterPool = &redis.Pool{
			MaxIdle:     100,
			MaxActive:   100,
			Wait:        true,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				masterAddr, err := sntnl.MasterAddr()
				if err != nil {
					return nil, err
				}
				c, err := redis.Dial("tcp", masterAddr)
				if err != nil {
					return nil, err
				}
				if driver.auth {
					_, err = c.Do("auth", driver.passWord)
					if err != nil {
						return nil, err
					}
				}
				return c, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if !sentinel.TestRole(c, "master") {
					return errors.New("Role check failed")
				} else {
					return nil
				}
			},
		}

		driver.slavePool = &redis.Pool{
			MaxIdle:     100,
			MaxActive:   100,
			Wait:        true,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				slaveAddr, err := sntnl.SlaveAddrs()
				if err != nil {
					return nil, err
				}
				rand.Seed(time.Now().Unix())
				c, err := redis.Dial("tcp", slaveAddr[rand.Intn(len(slaveAddr)-1)])
				if err != nil {
					return nil, err
				}
				if driver.auth {
					_, err = c.Do("auth", driver.auth)
					if err != nil {
						return nil, err
					}
				}
				return c, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if !sentinel.TestRole(c, "master") {
					return errors.New("Role check failed")
				} else {
					return nil
				}
			},
		}
		driver.passWord = driver.passWord
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

func (driver *RedisDriver) GetMasterConn() redis.Conn {
	if driver.masterPool == nil {
		return nil
	}
	return driver.masterPool.Get()
}

func (driver *RedisDriver) GetSlaveConn() redis.Conn {
	if driver.slavePool == nil {
		return nil
	}
	return driver.slavePool.Get()
}

func (driver *RedisDriver) Stop(ctx context.Context) error {
	if driver.masterPool != nil {
		driver.masterPool.Close()
	}
	if driver.slavePool != nil {
		driver.slavePool.Close()
	}
	return nil
}

//集合 使用多参数传递
func (driver *RedisDriver) SetAddMany(key string, value ...string) (int64, error) {
	conn := driver.GetMasterConn()
	defer conn.Close()
	res, err := redis.Int64(conn.Do("SADD", redis.Args{}.Add(key).AddFlat(value)...))
	if err != nil {
		return 0, err
	}
	return res, nil
}

//集合删除 使用多参数传递
func (driver *RedisDriver) SetDelMany(key string, value ...string) (int64, error) {
	conn := driver.GetMasterConn()
	defer conn.Close()
	res, err := redis.Int64(conn.Do("SREM", redis.Args{}.Add(key).AddFlat(value)...))
	if err != nil {
		return 0, err
	}
	return res, nil
}

//集合 添加
func (driver *RedisDriver) SetAddOne(key string, value string) (int64, error) {
	conn := driver.GetMasterConn()
	defer conn.Close()
	res, err := redis.Int64(conn.Do("SADD", key, value))
	if err != nil {
		return 0, err
	}
	return res, nil
}

//hash 设置
func (driver *RedisDriver) HSet(key string, value string, value2 string) error {

	conn := driver.GetMasterConn()
	defer conn.Close()

	_, err := conn.Do("hset", key, value, value2)

	if err != nil {
		return err
	}
	return nil
}

//hash获取
func (driver *RedisDriver) HGetOne(key, field string) (string, error) {

	conn := driver.GetSlaveConn()
	defer conn.Close()
	res, err := conn.Do("hget", key, field)
	if res == nil {
		return "", err
	}
	return string(res.([]byte)), err
}

//字符串设置
func (driver *RedisDriver) Set(key, value string) error {
	conn := driver.GetMasterConn()
	defer conn.Close()
	_, err := redis.Values(conn.Do("set", key, value))
	return err
}

//字符串获取
func (driver *RedisDriver) Get(key string) (string, error) {
	conn := driver.GetSlaveConn()
	defer conn.Close()
	res, err := redis.String((conn.Do("get", key)))
	if err != nil {
		return "", err
	}
	if res == "" {
		return "", nil
	}
	return res, nil
}

//判断key是否存在 存在返回true 否则false
func (driver *RedisDriver) Exists(key string) (bool, error) {
	if key == "" {
		return false, nil
	}
	conn := driver.GetSlaveConn()
	defer conn.Close()
	res, err := conn.Do("exists", key)
	if err != nil {
		return false, err
	}
	if res.(int64) == 0 {
		return false, nil
	}
	return true, nil
}

//获取set的数量
func (driver *RedisDriver) Scard(key string) (int, error) {
	if key == "" {
		return 0, nil
	}
	conn := driver.GetSlaveConn()
	defer conn.Close()
	res, err := redis.Int64((conn.Do("scard", key)))
	if err != nil {
		return 0, err
	}
	return int(res), nil
}

//设置key过期时间
func (driver *RedisDriver) SetExpire(key string, ExpireTime int) error {

	conn := driver.GetMasterConn()

	defer conn.Close()

	_, err := conn.Do("Expire", key, ExpireTime)

	if err != nil {
		return err
	}
	return nil
}

//单机版redis
type RedisDriverAlone struct {
	addr      string
	auth      bool
	passWord  string
	redisPool *redis.Pool
}

func NewRedisDriverAlone( password, addr string, auth bool) *RedisDriverAlone {
	return &RedisDriverAlone{
		addr:     addr,
		passWord: password,
		auth:     auth,
	}
}

func (driver *RedisDriverAlone) Start(ctx context.Context) error {

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
		driver.redisPool = &redis.Pool{
			MaxIdle:     100,
			MaxActive:   100,
			Wait:        true,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", driver.addr)
				if err != nil {
					fmt.Println("connect master addr errors [", err, "]")
					return nil, err
				}
				if driver.auth {
					_, err = c.Do("auth", driver.passWord)
					if err != nil {
						return nil, err
					}
				}
				return c, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if !sentinel.TestRole(c, "master") {
					return errors.New("Role check failed")
				} else {
					return nil
				}
			},
		}
		sigSuc <- struct{}{}
	}()

	select {
	case <-sigSuc:
		return nil
	case err := <-sigErr:
		return err
	case <-ctx.Done():
		return fmt.Errorf("service starting is timeout")
	}
}

func (driver *RedisDriverAlone) GetMasterConn() redis.Conn {
	if driver.redisPool == nil {
		return nil
	}
	return driver.redisPool.Get()
}

func (driver *RedisDriverAlone) Stop(ctx context.Context) error {
	if driver.redisPool != nil {
		driver.redisPool.Close()
	}

	return nil
}

///////封装的一些redis方法
//获取到
func (r *RedisDriverAlone) Get(key string) (interface{}, error) {
	conn := r.redisPool.Get()
	res, err := conn.Do("get", key)
	defer conn.Close()
	if res != nil {
		return string(res.([]uint8)), nil
	}
	return res, err
}

// incr 命令将 key 中储存的数字值增一。
func (r *RedisDriverAlone) Incr(key string) bool {
	conn := r.redisPool.Get()
	_, err := conn.Do("incr", key)
	defer conn.Close()
	if err != nil {
		return false
	}
	return true
}

//Decr 命令将 key 中储存的数字值减一。
func (r *RedisDriverAlone) Decr(key string) bool {
	conn := r.redisPool.Get()
	_, err := conn.Do("decr", key)
	defer conn.Close()
	if err != nil {
		return false
	}
	return true
}

//获取到 string 缓存
func (r *RedisDriverAlone) GetString(key string) (interface{}, error) {
	conn := r.redisPool.Get()
	res, err := conn.Do("get", key)
	defer conn.Close()
	if res != nil {
		return string(res.([]uint8)), nil
	}
	return "", err
}

//获取到 []byte] 缓存
func (r *RedisDriverAlone) GetBytes(key string) ([]byte, error) {
	conn := r.redisPool.Get()
	res, err := conn.Do("get", key)
	defer conn.Close()
	if res != nil {
		return res.([]uint8), nil
	}
	return nil, err
}

//设置redis 带时间 缓存
func (r *RedisDriverAlone) Setex(key string, ttl int, value interface{}) error {
	conn := r.redisPool.Get()
	_, err := conn.Do("setex", key, ttl, value)
	defer conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisDriverAlone) SetString(key string, value interface{}) error {
	conn := r.redisPool.Get()
	_, err := conn.Do("set", key, value)
	defer conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// 设置字符串 过期时间
func (r *RedisDriverAlone) Expireat(key string, value interface{}) error {
	conn := r.redisPool.Get()
	_, err := conn.Do("Expire", key, value)
	defer conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisDriverAlone) DelAndSet(key string, value interface{}) error {
	conn := r.redisPool.Get()
	_, err := conn.Do("del", key)
	defer conn.Close()
	if err != nil {
		return err
	}
	_, err = conn.Do("set", key, value)
	return err
}

//获取redis的hgetall
func (r *RedisDriverAlone) HgetAll(do, key string) (map[string]string, error) {
	if !strings.Contains(do, "hget") {
		return nil, fmt.Errorf("this older must includes hget")
	}

	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	res, err := redisConn.Do(do, key)
	if err != nil {
		return nil, err
	}

	keys := []string{}
	values := []string{}

	result := make(map[string]string)
	temRes := res.([]interface{})
	for k, v := range temRes {
		if k%2 == 0 {
			keys = append(keys, string(v.([]uint8)))
		} else {
			values = append(values, string(v.([]uint8)))
		}
	}
	if len(keys) != len(values) {
		return nil, fmt.Errorf("this func keys must eq values")
	}

	for i := 0; i < len(keys); i++ {
		result[keys[i]] = values[i]
	}
	return result, nil
}

//获取到redis的hget
func (r *RedisDriverAlone) Hget(keyName, fieldName string) (string, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	res, err := redisConn.Do("hget", keyName, fieldName)
	if err != nil || res == nil {
		return "", err
	}

	if _, ok := res.([]uint8); !ok {
		return "", fmt.Errorf("redis查询数据断言失败")
	}
	return string(res.([]uint8)), nil
}

//获取到redis的hget
func (r *RedisDriverAlone) HmGet(keyName string, fieldOne string, fieldTwo string, fieldThree string) ([]interface{}, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	/*	res, err := redisConn.Do("hmget", keyName, fieldName, two, three)
		if err != nil || res == nil {
			return nil, err
		}*/
	value, err := redis.Values(redisConn.Do("hmget", keyName, fieldOne, fieldTwo, fieldThree))
	if err != nil {
		return nil, err
	}
	return value, nil
}

//获取当前的hash 所有值
func (r *RedisDriverAlone) HVals(key string) (interface{}, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	res, err := redisConn.Do("hVals", key)
	if err != nil || res == nil {
		return nil, err
	}
	return res, nil
}

//添加redis hash
func (r *RedisDriverAlone) Hset(key, field string, insertData interface{}) (bool, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	_, err := redisConn.Do("hset", key, field, insertData)
	if err != nil {
		fmt.Println(fmt.Sprintf("redis存入数据hset失败，err:%s,key:%s,field:%s", err, key, field))
		return false, err
	}
	return true, nil
}

//添加redis hash过期时间（顶级）
func (r *RedisDriverAlone) Expire(key string, num int) (bool, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	_, err := redisConn.Do("Expire", key, num)
	if err != nil {
		return false, err
	}
	return true, nil
}

//redis 查看时间
func (r *RedisDriverAlone) TTL(key string) (bool, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	_, err := redisConn.Do("ttl", key)
	if err != nil {
		return false, err
	}
	return true, nil
}

//添加redis hash todo 改变传入值field 类型
func (r *RedisDriverAlone) NHset(key string, field interface{}, insertData interface{}) (bool, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	_, err := redisConn.Do("hset", key, field, insertData)
	if err != nil {
		return false, err
	}
	return true, nil
}

//获取到redis的hdel
func (r *RedisDriverAlone) Hdel(keyName, fieldName string) (bool, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	res, err := redisConn.Do("hdel", keyName, fieldName)
	if err != nil || res == nil {
		return false, err
	}
	return true, nil
}

//添加redis hash
func (r *RedisDriverAlone) Rpush(key string, value interface{}) (bool, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	_, err := redisConn.Do("rpush", key, value)
	if err != nil {
		return false, err
	}
	return true, nil
}


//移除并返回列表的第一个元素。
func (r *RedisDriverAlone) Lpop(key string) ([]uint8, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	res, err := redisConn.Do("lpop", key)
	if err != nil {
		return nil, err
	}
	if vv, ok := res.([]uint8); ok {
		return vv, nil
	}
	return nil, nil
}

//移除 最后一个 redis hash
func (r *RedisDriverAlone) Rpop(key string) (bool, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	_, err := redisConn.Do("rpop", key)
	if err != nil {
		return false, err
	}
	return true, nil
}

//删除redis key
func (r *RedisDriverAlone) Dels(key string) error {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	_, err := redisConn.Do("del", key)
	if err != nil {
		return err
	}
	return nil
}

//模糊删除 key
func (r *RedisDriverAlone) DelVagues(key string) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	res, err := redisConn.Do("keys", key)
	if err != nil {
		fmt.Println("获取模糊查询数据合集失败")
		return
	}
	Res := res.([]interface{})
	for _, v := range Res {
		if vv, ok := v.([]uint8); ok {
			_, err := redisConn.Do("del", string(vv))
			if err != nil {
				fmt.Println(fmt.Sprintf("删除键失败，key:%s", v))
			}
		}
	}
	return
}

// 获取list指定范围
func (r *RedisDriverAlone) LRange(key string, start int, end int) ([]interface{}, error) {
	redisConn := r.redisPool.Get() //获取redis实例
	defer redisConn.Close()
	list, err := redis.Values(redisConn.Do("Lrange", key, start, end))
	if err != nil {
		return nil, err
	}
	return list, err
}
