package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	file, err := os.Create("test.sql")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	now := time.Now()

	for i := 0; i < 10000; i++ {
		now = now.Add(time.Minute * 1)
		content := fmt.Sprintf("insert into users(`email`, `password`) values ('%v@webook.com','$2a$10$yjQUau/dcrSk0qdZnk8mxO.8cfB6/3o/7atZ/kZ0E5FgxC4bkx4dq');", now.Format("200601021504")) //,Ew333W#23fget

		_, err = file.WriteString(content + "\n")
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println("文件写入成功。")
}
