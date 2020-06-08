package main

import (
	"bufio"
	"fmt"
	"nasu/config"
	"os"
)

func main() {
	pass := config.PASSWORD{}

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("New Password: ")
	p, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	pass.Set(p[:len(p)-1])
	fmt.Println()
	fmt.Println("Here's the entry to use in your config:")
	fmt.Println(pass.Serialize())
}
