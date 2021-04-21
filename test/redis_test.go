package test

import (
	"fmt"
	"github.com/DMingGu/go_common/lib"
	"testing"
)

func Test_Redis(t *testing.T) {
	SetUp()
	defer lib.Destroy()
	//获取链接池

	replay,err:=lib.RedisConfDo(lib.NewTrace(),"default","hvals", "dm")
	fmt.Println(replay,err)

}
