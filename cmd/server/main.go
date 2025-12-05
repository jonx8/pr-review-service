package main

import (
	"os"

	"github.com/jonx8/pr-review-service/internal/app"
)

func main() {
	if err := app.RunApplication(); err != nil {
		os.Exit(1)
	}
}
