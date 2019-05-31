package log

import (
	"testing"
)

func TestMyLog_Init(t *testing.T) {
	logger := NewLogger(StruLog{DebugFileName: "Debug/debug.log"})
	defer logger.Flush()
	/*
		logger.Debug.Trace("111111111")
		m := make(map[string]int)
		m["1"] = 1
		logger.Debug.Debug("%d %s %v", 1, "sssss", m)
		err := errors.New("error")
		logger.Debug.Debug("错误是：", err)
		logger.Debug.Debug("错误是：%v", err)
		logger.Debug.Debug("这是第一个数%v", 1, "这是第二个数%v", 2)
		logger.Debug.Debug("这是第一个数%v,这是第二个数%v", 1, 2)
		b := []byte("11111111111111111111")
		c := []byte("22222222")
		logger.Debug.Debug(string(b))
		logger.Debug.Debug(string(b), string(c))
	*/
	logger.Debug.Trace("11111111")
	logger.Debug.Debug("222222222222")
	logger.Debug.Info("3333333")
	logger.Debug.Notice("444444444")
	logger.Debug.Warn("55555555")
	logger.Debug.Error("66666666")
	logger.Debug.Critical("7777777")

}
