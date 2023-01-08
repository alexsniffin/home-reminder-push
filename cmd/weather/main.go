package main

import (
	"home-reminder-push/internal/weather/clients/darksky"
	"home-reminder-push/internal/weather/clients/firebase"
	"home-reminder-push/internal/weather/process"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	args := os.Args
	done := make(chan struct{})

	darkskyClient := darksky.New(args[1])
	firebaseClient, err := firebase.New()
	if err != nil {
		log.Fatal(err)
	}

	p := process.New(darkskyClient, firebaseClient)
	ch := p.Start(done)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	signal.Notify(stop, syscall.SIGTERM)
	<-stop
	done <- struct{}{}
	<-ch
}
