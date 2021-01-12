/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package redis

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"testing"
)

var testKey = "demokey"

// 本地没有配置主从模式的redis  使用开发环境哨兵主从的redis
var testDnName = "mymaster"
var testAddrs = []string{"10.0.3.196:26379", "10.0.3.196:26380", "10.0.3.196:26381"}
var testAuth = false
var testPassWord = "czn"

func TestRedisAdd(t *testing.T) {
	redisPoll := NewRedisDriver(testDnName, testAddrs)
	err := redisPoll.Open(testAuth, testPassWord)
	if err != nil {
		fmt.Println("TestRedisAdd errors:", err)
		return
	}
	defer redisPoll.Close()
	conn := redisPoll.GetMasterConn()
	_, err = conn.Do("set", testKey, "1")
	if err != nil {
		fmt.Println("TestRedisAdd errors:", err)
	} else {
		fmt.Println("TestRedisAdd OK")
	}
}

func TestRedisFind(t *testing.T) {
	redisPoll := NewRedisDriver(testDnName, testAddrs)
	err := redisPoll.Open(testAuth, testPassWord)
	if err != nil {
		fmt.Println("TestRedisAdd errors:", err)
		return
	}
	defer redisPoll.Close()
	conn := redisPoll.GetSlaveConn()
	res, err := conn.Do("get", testKey)
	if err != nil {
		fmt.Println("TestRedisFind errors:", err)
	}

	str, err := redis.String(res, err)
	if err != nil {
		fmt.Println("TestRedisFind errors:", err)
	} else {
		fmt.Println("TestRedisFind OK result:", str)
	}

}

func TestRedisDel(t *testing.T) {
	redisPoll := NewRedisDriver(testDnName, testAddrs)
	err := redisPoll.Open(testAuth, testPassWord)
	if err != nil {
		fmt.Println("TestRedisAdd errors:", err)
		return
	}
	defer redisPoll.Close()
	conn := redisPoll.GetMasterConn()
	_, err = conn.Do("del", testKey)
	if err != nil {
		fmt.Println("TestRedisDel errors:", err)
	} else {
		fmt.Println("TestRedisDel OK")
	}
}
