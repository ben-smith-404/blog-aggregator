package main

import (
	"fmt"

	config "github.com/ben-smith-404/blog-aggregator/internal/config"
)

func main() {
	path, err := config.Read()
	if err != nil {
		fmt.Printf("Error: %v/n", err)
	}
	fmt.Println(path)
}
