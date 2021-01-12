/**
  @author: $(huangjiangming)
  @data:$(2020-11-18)
  @note:mongo封装驱动
**/
package mongo

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"

	//"time"
)

//mgo停更不使用， mongo-driver驱动接口比较原始
//mongo官方驱动简易封装   由于mongo提供一套基于bson独特的查询方式，
//并且在nosql环境中比sql功能还要强大 难以封装成redis那样简化的api,只能暴露原生接口供使用
type MongoDriver struct {
	host     string
	dbName   string
	user     string
	passWord string
	login    bool
	*mongo.Client
}

func NewMongoDriver(host, user, passWord string, login bool) *MongoDriver {
	return &MongoDriver{
		user:     user,
		passWord: passWord,
		login:    login,
	}
}

//避免MongoDriver操作是产生大量的中间对象，可以生成该对象，再执行操作
type mongoCollection struct {
	*mongo.Collection
}

//参数不能为空
func NewMongoMongoCollection(driver *MongoDriver, dbName, col string) *mongoCollection {
	mongoCol := &mongoCollection{driver.Database(dbName).Collection(col)}
	return mongoCol
}

func (d *MongoDriver) Start(ctx context.Context) (err error) {
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
		clientoptions := options.Client()
		//clientoptions.SetConnectTimeout(time.Duration(int(time.Second) * 10))
		//host := fmt.Sprintf("%s:%d", d.addr, d.port)
		//查看是否需要登录验证
		if d.login {
			mongoUrl := "mongodb://" + d.user + ":" + d.passWord + "@" + d.host + "/"
			clientoptions.ApplyURI(mongoUrl)

		} else {
			mongoUrl := "mongodb://" + d.host + "/"
			clientoptions.ApplyURI(mongoUrl)
		}

		client, err := mongo.Connect(context.Background(), clientoptions)
		if err != nil {
			sigErr <- err
			return
		}
		err = client.Ping(nil, nil)
		if err != nil {
			sigErr <- err
			return
		}
		d.Client = client
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

func (d *MongoDriver) Stop(ctx context.Context) error {
	if d.Client == nil {
		//log
		return nil
	}
	return d.Client.Disconnect(nil)
}

func (d *MongoDriver) GetDB() *mongo.Database {
	if d.Client == nil {
		return nil
	}
	return d.Client.Database(d.dbName)
}

func (d *MongoDriver) GetCollection(dbName, colName string) *mongo.Collection {
	if d.Client == nil {
		return nil
	}
	return d.Client.Database(dbName).Collection(colName)
}

func (d *MongoDriver) InsertOne(dbName string, colName string, obj interface{}) error {
	_, err := d.Client.Database(dbName).Collection(colName).InsertOne(context.Background(), obj)
	if err != nil {
		return err
	}
	return nil
}

func (d *MongoDriver) InsertAll(dbName string, colName string, obj []interface{}) (int64, error) {
	res, err := d.Client.Database(dbName).Collection(colName).InsertMany(context.Background(), obj)
	if err != nil {
		return 0, err
	}
	count := int64(len(res.InsertedIDs))
	return count, nil
}

//删除，目前只支持键值匹配
func (d *MongoDriver) DeleteOne(dbName, colName string, filter map[string]string) error {
	info := bson.D{}
	for k, v := range filter {
		info = append(info, bson.E{k, v})
	}
	_, err := d.Client.Database(dbName).Collection(colName).DeleteOne(context.Background(), info)
	if err != nil {
		return err
	}
	return nil
}

//删除，目前只支持键值匹配
func (d *MongoDriver) DeleteAll(dbName, colName string, filter map[string]string) (int64, error) {
	info := bson.D{}
	for k, v := range filter {
		info = append(info, bson.E{k, v})
	}
	res, err := d.Client.Database(dbName).Collection(colName).DeleteMany(context.Background(), info)
	if err != nil {
		return 0, err
	}

	return res.DeletedCount, nil
}

//删除，目前只支持键值匹配
func (d *MongoDriver) Update(dbName, colName string, filter map[string]string, update map[string]string) (int64, error) {
	//组织成bson格式
	info := bson.D{}
	for k, v := range filter {
		info = append(info, bson.E{k, v})
	}

	up := bson.D{}
	for k, v := range update {
		up = append(up, bson.E{k, v})
	}
	res, err := d.Client.Database(dbName).Collection(colName).UpdateMany(context.Background(), info, bson.D{{"$set", up}})
	if err != nil {
		return 0, err
	}

	return res.UpsertedCount, nil
}

func (d *MongoDriver) FindOne(dbName string, colName string, filter map[string]string, f func() interface{}) (interface{}, error) {
	info := bson.D{}
	for k, v := range filter {
		info = append(info, bson.E{k, v})
	}

	res := d.Client.Database(dbName).Collection(colName).FindOne(context.Background(), info)
	if res.Err() != nil {
		return nil, res.Err()
	}

	obj := f()
	err := res.Decode(obj)
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (d *MongoDriver) FindAll(dbName string, colName string, filter map[string]string, f func() interface{}) (*[]interface{}, error) {
	info := bson.D{}
	for k, v := range filter {
		info = append(info, bson.E{k, v})
	}

	res, err := d.Client.Database(dbName).Collection(colName).Find(context.Background(), info)
	if err != nil {
		return nil, err
	}
	data := make([]interface{}, 0)
	for res.Next(context.Background()) {
		obj := f()
		err = res.Decode(obj)
		if err != nil {
			return nil, err
		} else {
			data = append(data, obj)
		}
	}
	return &data, nil
}

func (c *mongoCollection) InsertAll(obj []interface{}) (int64, error) {
	res, err := c.InsertMany(context.Background(), obj)
	if err != nil {
		return 0, err
	}
	count := int64(len(res.InsertedIDs))
	return count, nil
}

//删除，目前只支持键值匹配
func (c *mongoCollection) DeleteAll(filter map[string]string) (int64, error) {
	info := bson.D{}
	for k, v := range filter {
		info = append(info, bson.E{k, v})
	}
	res, err := c.DeleteMany(context.Background(), info)
	if err != nil {
		return 0, err
	}

	return res.DeletedCount, nil
}

//删除，目前只支持键值匹配
func (c *mongoCollection) UpdateAll(filter map[string]string, update map[string]string) (int64, error) {
	//组织成bson格式
	info := bson.D{}
	for k, v := range filter {
		info = append(info, bson.E{k, v})
	}

	up := bson.D{}
	for k, v := range update {
		up = append(up, bson.E{k, v})
	}
	res, err := c.UpdateMany(context.Background(), info, bson.D{{"$set", up}})
	if err != nil {
		return 0, err
	}

	return res.UpsertedCount, nil
}

func (c *mongoCollection) FindAll(filter map[string]string, f func() interface{}) (*[]interface{}, error) {
	info := bson.D{}
	for k, v := range filter {
		info = append(info, bson.E{k, v})
	}

	res, err := c.Find(context.Background(), info)
	if err != nil {
		return nil, err
	}
	data := make([]interface{}, 0)
	for res.Next(context.Background()) {
		obj := f()
		err = res.Decode(obj)
		if err != nil {
			return nil, err
		} else {
			data = append(data, obj)
		}
	}
	return &data, nil
}
