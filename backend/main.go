package main

import (
	"fmt"
	"uuid"
)

func main() {
	fmt.Println(uuid.NewV7())
	fmt.Println(uuid.New())
}
