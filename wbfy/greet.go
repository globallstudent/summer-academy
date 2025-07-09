// Greeting program that asks for the user's name
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	fmt.Print("What is your name? ")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)
	fmt.Printf("Hello, %s!\n", name)
}
