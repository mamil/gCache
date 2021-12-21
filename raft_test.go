package main

import (
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	addr := "http://localhost:8081"
	go receive(addr)
	time.Sleep(time.Duration(2) * time.Second)
	send("test from send", addr)
}
