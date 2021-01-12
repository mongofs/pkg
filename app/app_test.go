package app

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestApp_Run(t *testing.T) {
	serv := &serverTest{}
	app := New(SetStartTimeOut(2 * time.Second))
	app.Append(serv)
	app.Append(serv)
	err := app.Run()
	fmt.Println(err)
}


type serverTest struct{}

func (s *serverTest) Start(ctx context.Context) error {
	sig := make(chan struct{})
	go func() {
		defer func() {
			sig <- struct{}{}
		}()
		time.Sleep(50 * time.Second)
		fmt.Println("im ready")
	}()

	select {
	case <-sig:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("service starting is timeout")
	}

}
func (s *serverTest) Stop(ctx context.Context) error {
	fmt.Println("im finished ")
	return nil
}
