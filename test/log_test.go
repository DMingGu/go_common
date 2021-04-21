package test

import (
	"github.com/DMingGu/go_common/lib"
	"testing"
	"time"
)

//测试日志打点
func TestDefaultLog(t *testing.T) {
	SetUp()
	lib.Log.TagInfo(lib.NewTrace(), lib.DLTagMySqlSuccess, map[string]interface{}{
		"sql": "sql",
	})
	time.Sleep(time.Second)
}

//测试日志实例打点
func TestLogInstance(t *testing.T) {

	time.Sleep(time.Second)
}