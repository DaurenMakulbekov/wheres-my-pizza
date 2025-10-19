package main

import (
	"flag"
	"wheres-my-pizza/order-service/cmd/app"
)

func main() {
	var mode = flag.String("mode", "", "")
	//var port = flag.String("port", "3000", "The HTTP port for the API")
	//var concurrent = flag.Int("max-concurrent", 50, "Maximum number of concurrent orders to process")
	flag.Parse()

	if *mode == "order-service" {
		app.Run()
	}
}
