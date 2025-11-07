package main

import (
	"fmt"
	"log"

	"github.com/Vadim-Makhnev/quickenv"
)

func main() {
	count, err := quickenv.Load(&quickenv.LoadOptions{Overwrite: true, Debug: true})
	if err != nil {
		log.Fatal(err)
	}

	str := quickenv.GetEnv("CONFIG_PATH", "config/local.env")
	port := quickenv.GetEnv("DB_PORT", "8000")

	fmt.Println(count, str, port)
}
