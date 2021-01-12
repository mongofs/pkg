/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package mongo

//import (
//	"flag"
//	"fmt"
//	"log"
//	"strconv"
//	"sync"
//	"testing"
//)
//
////go test -v mongo.go mongo_test.go -args 10.0.0.174 27017
//
//type TestData struct {
//	User   string
//	Passwd string
//}
//
////增
//func TestMongoAdd(t *testing.T) {
//	param := flag.Args()
//	if len(param) == 0 {
//		log.Println("Please Enter Addr")
//		return
//	}
//	addr := param[0]
//	port, _ := strconv.Atoi(param[1])
//
//	driver := NewMongoDriver(addr, port)
//	//"10.0.0.174:27017"
//	err := driver.Open("", "", false)
//	if err != nil {
//		log.Panic(err)
//	}
//	defer driver.Close()
//
//	data := TestData{
//		User:   "user",
//		Passwd: "passwd",
//	}
//
//	var wg sync.WaitGroup
//	//测试并发安全 没有没有并发问题
//	for i := 0; i < 100; i++ {
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//			for j := 0; j < 100; j++ {
//				driver.InsertOne("test", "demo", &data)
//				if err != nil {
//					log.Println(err)
//				}
//			}
//		}()
//	}
//	wg.Wait()
//	log.Println("insert count :10000")
//}
//
////查
//func TestMongoFind(t *testing.T) {
//	param := flag.Args()
//	if len(param) == 0 {
//		log.Println("Please Enter Addr")
//		return
//	}
//	addr := param[0]
//	port, _ := strconv.Atoi(param[1])
//
//	driver := NewMongoDriver(addr, port)
//	//"10.0.0.174:27017"
//	err := driver.Open("", "", false)
//	if err != nil {
//		log.Panic(err)
//	}
//	defer driver.Close()
//
//	//find
//	res, err := driver.FindAll("test", "demo", map[string]string{
//		"user": "user"}, func() interface{} {
//		return &TestData{}
//	})
//	result := []TestData{}
//	for k := range *res {
//		result = append(result, *(*res)[k].(*TestData))
//	}
//
//	if err != nil {
//		log.Println(err)
//	} else {
//		log.Println("find count:", len(result))
//	}
//
//}
//
////改
//func TestMongoUpdate(t *testing.T) {
//	param := flag.Args()
//	if len(param) == 0 {
//		log.Println("Please Enter Addr")
//		return
//	}
//	addr := param[0]
//	port, _ := strconv.Atoi(param[1])
//
//	driver := NewMongoDriver(addr, port)
//	//"10.0.0.174:27017"
//	err := driver.Open("", "", false)
//	if err != nil {
//		log.Panic(err)
//	}
//	defer driver.Close()
//
//	count, err := driver.Update("test", "demo", map[string]string{"user": "user"}, map[string]string{"passwd": "passwd1111"})
//	if err != nil {
//		fmt.Println(err)
//	} else {
//		log.Println("update count :", count)
//	}
//}
//
////删
//func TestMongoDel(t *testing.T) {
//	param := flag.Args()
//	if len(param) == 0 {
//		log.Println("Please Enter Addr")
//		return
//	}
//	addr := param[0]
//	port, _ := strconv.Atoi(param[1])
//
//	driver := NewMongoDriver(addr, port)
//	//"10.0.0.174:27017"
//	err := driver.Open("", "", false)
//	if err != nil {
//		log.Panic(err)
//	}
//	defer driver.Close()
//
//	//del
//	count, err := driver.DeleteAll("test", "demo", map[string]string{
//		"user": "user",
//	})
//	if err != nil {
//		log.Println(err)
//	} else {
//		log.Println("del count:", count)
//	}
//}
//
//func TestCollection(t *testing.T) {
//	param := flag.Args()
//	if len(param) == 0 {
//		log.Println("Please Enter Addr")
//		return
//	}
//	addr := param[0]
//	port, _ := strconv.Atoi(param[1])
//
//	driver := NewMongoDriver(addr, port)
//	//"10.0.0.174:27017"
//	err := driver.Open("", "", false)
//	if err != nil {
//		log.Panic(err)
//	}
//	defer driver.Close()
//	//程序拿到coll  执行批量操作，避免使用driver生成大量的db和coll对象
//	coll := NewMongoMongoCollection(driver, "test", "demo")
//	data := []TestData{
//		{
//			User:   "user",
//			Passwd: "passwd",
//		},
//		{
//			User:   "user",
//			Passwd: "passwd1",
//		},
//	}
//	//  []T 无法隐式转化为[]interface{}
//	insert := []interface{}{}
//	for _, v := range data {
//		insert = append(insert, v)
//	}
//
//	count, err := coll.InsertAll(insert)
//	if err != nil {
//		log.Println(err)
//	} else {
//		log.Println("insert count:", count)
//	}
//
//	count, err = coll.DeleteAll(map[string]string{"user": "user"})
//	if err != nil {
//		log.Println(err)
//	} else {
//		log.Println("del count:", count)
//	}
//
//}
