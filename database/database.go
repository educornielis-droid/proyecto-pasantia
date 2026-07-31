package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// Struct de Productos (mantenido aquí como querías)
type Productos struct {
	// ProductoID      int     `json:"producto_id"`
	Nombre          string  `json:"nombre"`
	Descripcion     string  `json:"descripcion"`
	NombreCategoria string  `json:"nombre_categoria"`
	Precio          float64 `json:"precio"`
	Stock           int     `json:"stock"`
}

// Variable global del paquete database para reutilizar la conexión
var DB *pgx.Conn

// 1. Función para inicializar la conexión (se llama en el main.go al arrancar la app)
func ConectarDB() error {
	var err error
	DB, err = pgx.Connect(
		context.Background(),
		"postgres://postgres:123456@localhost:5432/bd_proyecto",
	)

	if err != nil {
		return fmt.Errorf("error al conectar a la BD: %w", err)
	}

	fmt.Println("Conexión a PostgreSQL establecida con éxito")
	return nil
}

// 2. Función para cerrar la conexión cuando se apague el servidor
func CerrarDB() {
	if DB != nil {
		DB.Close(context.Background())
		fmt.Println("Conexión a PostgreSQL cerrada")
	}
}

// 3. Función que consulta la vista y RETORNA los productos
func ObtenerProductos() ([]Productos, error) {
	// Usamos la conexión global DB
	rows, err := DB.Query(context.Background(), "SELECT nombre, descripcion, nombre_categoria, precio, stock FROM v_productos")
	if err != nil {
		return nil, fmt.Errorf("error al consultar v_productos: %w", err)
	}
	defer rows.Close()

	var listaProductos []Productos

	for rows.Next() {
		var p Productos

		// Escaneamos directamente a las propiedades de nuestro struct
		err := rows.Scan(&p.Nombre, &p.Descripcion, &p.NombreCategoria, &p.Precio, &p.Stock)
		if err != nil {
			return nil, fmt.Errorf("error al escanear fila: %w", err)
		}

		// Añadimos el producto a la lista
		listaProductos = append(listaProductos, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error durante la iteración: %w", err)
	}

	return listaProductos, nil
}
