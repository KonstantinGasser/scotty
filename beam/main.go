package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("beam me up, Scotty")

	for i := 0; i < 5; i++ {
		fmt.Printf("[INFO] msg=I am log #%d\n", i)
		time.Sleep(time.Second * 1)
	}

	fmt.Println("[INFO] msg=shutting down server")
}
