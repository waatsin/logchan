package main

import (
	"flag"
	"fmt"
       "log"
)

func main() {
         // 按照所需读写权限创建文件
        f, err := os.OpenFile("/var/log/logchan.log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	flag.Parse()
	num, err := ReadFile()

        // 完成后延迟关闭
        defer f.Close()
        //设置日志输出到 f
        log.SetOutput(f)

	if err != nil {
		fmt.Printf("%#v\n", err)
                log.Printf("%#v\n", err)
	} else {
		fmt.Printf("send msg line : %v\n", num)
                log.Printf("send msg line : %v\n", num)
	}
       
}
