package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	file, err := os.Create("test.csv")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	now := time.Now()

	file.WriteString("email\n")

	for i := 0; i < 100000; i++ {
		now = now.Add(time.Minute * 1)
		content := fmt.Sprintf("%v@webook.com", now.Format("200601021504")) //,Ew333W#23fget

		_, err = file.WriteString(content + "\n")
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println("文件写入成功。")
}
