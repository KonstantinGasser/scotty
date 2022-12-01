package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
)

func main() {
	fmt.Println("Scotty, reporting for duty!")
	defer fmt.Println("\nScotty, signing off!")

	info, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	if info.Mode()&os.ModeCharDevice == os.ModeCharDevice || info.Size() <= 0 {
		fmt.Println("Program requires input through pipes\n\tUsage: cat logs.log | beam")
		return
	}

	f, err := os.Create("temp.log")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	reader := bufio.NewReader(os.Stdin)
	var i = 1
	for {

		log, _, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		line := fmt.Sprintf("%v: %s\n", time.Now(), string(log))
		if _, err := f.WriteString(line); err != nil {
			panic(err)
		}
		fmt.Printf("\rLogs send(%d)", i)
		i++
	}
}
