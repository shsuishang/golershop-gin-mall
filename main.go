package main

import (
	"log"

	"golershop.cn/internal/cmd"
)

func main() {
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
