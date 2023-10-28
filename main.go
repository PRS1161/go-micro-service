package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/PRS1161/go-micro-service/application"
)

func main() {
	app := application.StartServer()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	defer cancel()

	err := app.Start(ctx)

	if err != nil {
		fmt.Println("THERE IS A ISSUE WHILE STARTING THE SERVER", err)
	}
}
