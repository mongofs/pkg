/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package zap

import (
	"errors"
	"go.uber.org/zap"
	"testing"
)

func TestZap(t *testing.T) {
	logger := InitZapLog("test", false)
	logger.Debug("测试", zap.Error(errors.New("test errors")))
	logger.Info("测试", zap.Error(errors.New("test errors")))
	logger.Warn("测试", zap.Error(errors.New("test errors")))
	logger.Error("测试", zap.Error(errors.New("test errors")))
}

func TestZapDebug(t *testing.T) {
	logger := InitZapLog("test", true)
	logger.Debug("测试", zap.Error(errors.New("debug errors")))
	logger.Info("测试", zap.Error(errors.New("debug errors")))
	logger.Warn("测试", zap.Error(errors.New("debug errors")))
	logger.Error("测试", zap.Error(errors.New("debug errors")))
}
