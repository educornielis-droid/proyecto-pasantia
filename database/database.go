package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

func Main() {
	fmt.Println("Hola, desde el archivo de database")

	conexion, err := pgx.Connect(
		context.Background(),
		"postgres://postgres:123456@localhost:5432/bd_proyecto",
	)

	if err != nil {
		panic(err)
	}

	defer conexion.Close(context.Background())
	fmt.Println("Conectado correctamente")
}
