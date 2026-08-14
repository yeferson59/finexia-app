package main

import (
	"fmt"

	"github.com/google/uuid"
)

func main() {
	fmt.Println(uuid.NewV7())
	fmt.Println(uuid.New())
}
