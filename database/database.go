package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// type Productos struct {
// 	ProductoID      int     `json:"producto_id"`
// 	NombreCategoria string  `json:"nombre_categoria"`
// 	Nombre          string  `json:"nombre"`
// 	Descripcion     string  `json:"descripcion"`
// 	Precio          float64 `json:"precio"`
// 	Stock           int     `json:"stock"`
// }

func Main() {
	fmt.Println("Hola, desde el archivo de database")

	conexion, err := pgx.Connect(
		context.Background(),
		"postgres://postgres:123456@localhost:5432/bd_proyecto",
	)

	if err != nil {
		panic(err)
	}

	rows, err := conexion.Query(context.Background(), "SELECT producto_id, nombre_categoria, nombre, descripcion, precio, stock FROM v_productos")
	if err != nil {
		panic(err)
	}

	for rows.Next() {
		var producto_id int
		var nombre_categoria string
		var nombre string
		var descripcion string
		var precio float64
		var stock int

		rows.Scan(&producto_id, &nombre_categoria, &nombre, &descripcion, &precio, &stock)
		fmt.Printf("\nID: %d, \nNombre de la categoria: %s, \nNombre del producto: %s, \nDescripcion del producto: %s, \nPrecio: %.2f, \nStock disponible: %d\n", producto_id, nombre_categoria, nombre, descripcion, precio, stock)
	}

	defer conexion.Close(context.Background())
	fmt.Println("\nConectado correctamente")
}
