package main

import (
	"os"
	"os/signal"
	"proyecto-golang/database"
	"proyecto-golang/servidor"
	"syscall"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go servidor.Main()

	database.ConectarDB()
	defer database.CerrarDB()

	<-sigs
}
