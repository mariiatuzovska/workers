package main

import (
	"fmt"
	"os"
	"time"

	"github.com/mariiatuzovska/workers/models"
	"github.com/mariiatuzovska/workers/service"
)

func main() {
	srv := service.NewService(10, -1, os.Stdout)
	srv.Start()
	for i := 0; i < 1000; i++ {
		srv.Query() <- &models.Message{Text: fmt.Sprintf("%d", i), Tag: "", CreatedAt: time.Now()}
	}
	srv.Clear()
}
