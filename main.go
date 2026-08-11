package main

import (
	"os"
	"os/signal"
	"proyecto-golang/database"
	"proyecto-golang/servidor"
	"syscall"

	"github.com/joho/godotenv"
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	godotenv.Load()

	go servidor.Main()

	database.ConectarDB()
	defer database.CerrarDB()

	<-sigs
}

/*
import "proyecto-golang/servidor"

func main() {
	servidor.TdC()
}
*/
