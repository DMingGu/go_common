package main

import (
	"github.com/DMingGu/go_common/lib"
	"os"
)
func main() {
	dir,_:=os.Getwd()
	lib.InitModule(dir+"/config/dev/", []string{"base","mysql","redis"})
	defer lib.Destroy()
}
