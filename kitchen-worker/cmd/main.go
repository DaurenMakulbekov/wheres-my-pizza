package main

import (
	"flag"
	"wheres-my-pizza/kitchen-worker/cmd/app"
)

func main() {
	var workerName = flag.String("worker-name", "", "Unique name for the worker")
	var orderTypes = flag.String("order-types", "", "Comma-separated list of order types the worker can handle")
	var heartbeatInterval = flag.Int("heartbeat-interval", 30, "Interval (seconds) between heartbeats")
	var prefetch = flag.Int("prefetch", 1, "RabbitMQ prefetch count, limiting how many messages the worker receives at once")
	flag.Parse()

	app.Run(*workerName, *orderTypes, *heartbeatInterval, *prefetch)
}
