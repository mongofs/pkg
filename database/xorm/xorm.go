/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package xorm

import (
	"context"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"time"
	"xorm.io/xorm"
)

var (
	Mysql   = "mysql"
	Mariadb = "mariadb"
)

//xorm简单封装，目前只支持mysql
type XormEngine struct {
	host     string
	dbName   string
	charSet  string
	user     string
	passWord string
	driver   string
	debug    bool
	*xorm.Engine
}

func NewXorm(host string, dbName, charSet, user, password string, debug bool) *XormEngine {
	return &XormEngine{
		host:     host,
		dbName:   dbName,
		charSet:  charSet,
		user:     user,
		passWord: password,
		debug:    debug,
	}
}

func (x *XormEngine) Start(ctx context.Context) error {
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

		dbUrl := ""
		if x.driver == "" {
			x.driver = Mysql
		}
		if x.driver == Mysql {
			//host := fmt.Sprintf("%s:%d", x.addr, x.port)
			dbUrl = fmt.Sprintf("%v:%v@(%v)/%v?charset=%v",
				x.user,
				x.passWord,
				x.host,
				x.dbName,
				x.charSet)

		} else if x.driver == Mariadb {
			sigErr <- fmt.Errorf("no support driver driverName:" + x.driver)
		} else {
			sigErr <- fmt.Errorf("no support driver driverName:" + x.driver)
		}

		dbEngine, err := xorm.NewEngine("mysql", dbUrl)
		if err != nil {
			sigErr <- err
		}

		time.LoadLocation("Asia/Shanghai")
		dbEngine.DatabaseTZ = time.Local
		dbEngine.TZLocation = time.Local
		dbEngine.ShowSQL(x.debug)
		dbEngine.SetConnMaxLifetime(60 * time.Second)
		dbEngine.SetMaxOpenConns(100)
		x.Engine = dbEngine
		sigSuc <- struct{}{}
	}()

	select {
	case <-sigSuc:
		fmt.Println("xorm : start service success")
		return nil
	case err := <-sigErr:
		return err
	case <-ctx.Done():
		return fmt.Errorf("service starting is timeout")
	}
}


func (x *XormEngine) Stop(ctx context.Context) error {
	if x.Engine == nil {
		return nil
	}
	return x.Engine.Close()
}
