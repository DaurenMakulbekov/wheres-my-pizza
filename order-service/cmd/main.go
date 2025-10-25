package main

import (
	"flag"
	"log"
	"strconv"
	"wheres-my-pizza/order-service/cmd/app"
)

func main() {
	var port = flag.String("port", "3000", "The HTTP port for the API")
	var maxConcurrent = flag.Int("max-concurrent", 50, "Maximum number of concurrent orders to process")
	flag.Parse()

	number, err := strconv.Atoi(*port)
	if err != nil || number < 1024 || number > 49151 {
		log.Fatal("Error: incorrect port number")
	}

	app.Run(*port, *maxConcurrent)
}
