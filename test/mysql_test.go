package test

import (
	"fmt"
	"github.com/DMingGu/go_common/lib"
	"testing"
	"time"
)

type Channel struct {
	Id        int64     `json:"id" gorm:"primary_key"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}
func Test_GORM(t *testing.T) {
	SetUp()
	defer lib.Destroy()
	//获取链接池
	db, err := lib.GetGormPool("default")
	if err != nil {
		t.Fatal(err)
	}
	var channel,channel1 Channel
	db.Where("id",1).First(&channel)
	db.Where("id",2).First(&channel1)
	fmt.Println("test==>",channel,channel1)


}
