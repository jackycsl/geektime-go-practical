package service

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestServerShutdown(t *testing.T) {
	s1 := NewServer("server_A", "8081")
	s2 := NewServer("server_B", "8082")

	s := NewApp([]*Server{s1, s2}, WithShutdownCallbacks(func(ctx context.Context) {
		log.Println("callback...")
		time.Sleep(2 * time.Second)
	}))

	s.StartAndServe()
}
