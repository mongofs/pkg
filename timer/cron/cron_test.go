/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package cron

import (
	"fmt"
	"testing"
	"time"
)

func TestCron(t *testing.T) {
	cron := NewCron()
	cron.AddFunc("*/1 * * * * ?", func() {
		fmt.Println("handler")
	})

	cron.Start()
	defer cron.Stop()
	time.Sleep(time.Second * 10)
}
