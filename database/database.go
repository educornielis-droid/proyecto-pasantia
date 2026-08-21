package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Struct de Productos
type Productos struct {
	ProductoID      int     `json:"producto_id"`
	Nombre          string  `json:"nombre"`
	Descripcion     string  `json:"descripcion"`
	NombreCategoria string  `json:"nombre_categoria"`
	Precio          float64 `json:"precio"`
	Stock           int     `json:"stock"`
	ImagenURL       string  `json:"imagen_url"`
}

// Struct de Transaccion
type Transacciones struct {
	TransaccionID          string    `json:"transaccion_id"`
	TipoDocumento          string    `json:"tipo_documento"`
	NumeroDocumento        string    `json:"numero_documento"`
	TipoCuenta             string    `json:"tipo_cuenta"`
	CuentaOTelefono        string    `json:"cuenta_o_telefono"`
	BancoOrigen            string    `json:"banco_origen"`
	MontoFinalUSD          float64   `json:"monto_final_usd"`
	MontoFinalVES          float64   `json:"monto_final_ves"`
	TasaCambio             float64   `json:"tasa_cambio"`
	EstadoTransaccion      string    `json:"estado_transaccion"`
	ReferenciaSypago       string    `json:"referencia_sypago"`
	FechaCreacion          time.Time `json:"fecha_creacion"`
	FechaActualizacion     time.Time `json:"fecha_actualizacion"`
	CodigoRechazo          string    `json:"codigo_rechazo"`
	EstadoReembolso        string    `json:"estado_reembolso"`     // "" = sin reembolso, PEND/PROC/ACCP/RJCT/CANC = en curso o resuelto
	ReferenciaReembolso    string    `json:"referencia_reembolso"` // transaction_id de Sypago para el crédito
	CodigoRechazoReembolso string    `json:"codigo_rechazo_reembolso"`
}

// Una línea de detalle: qué producto y cuántas unidades tenía la transacción
type DetalleTransaccion struct {
	ProductoID       int `json:"producto_id"`
	CantidadProducto int `json:"cantidad_producto"`
}

// Struct de Usuarios
type Usuarios struct {
	UsuarioID     int       `json:"usuario_id"`
	Nombre        string    `json:"nombre"`
	Apellido      string    `json:"apellido"`
	Correo        string    `json:"correo"`
	Contrasena    string    `json:"contrasena"`
	FechaCreacion time.Time `json:"fecha_creacion"`
	EsAdmin       bool      `json:"es_admin"`
}

// Variable global del paquete github.com/jackc/pgx/v5 para reutilizar la conexión
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

// 3. Función que consulta la vista y RETORNA los productos.
func ObtenerProductos() ([]Productos, error) {
	rows, err := DB.Query(
		context.Background(),
		"SELECT producto_id, nombre, descripcion, nombre_categoria, precio, stock, COALESCE(imagen_url, '') FROM v_productos ORDER BY producto_id ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("error al consultar v_productos: %w", err)
	}
	defer rows.Close()

	var listaProductos []Productos

	for rows.Next() {
		var p Productos
		err := rows.Scan(&p.ProductoID, &p.Nombre, &p.Descripcion, &p.NombreCategoria, &p.Precio, &p.Stock, &p.ImagenURL)
		if err != nil {
			return nil, fmt.Errorf("error al escanear fila: %w", err)
		}
		listaProductos = append(listaProductos, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error durante la iteración: %w", err)
	}

	return listaProductos, nil
}

// 4. Función que consulta UN SOLO producto por nombre.
func ObtenerProductoPorNombre(nombre string) (Productos, error) {
	var p Productos

	err := DB.QueryRow(
		context.Background(),
		"SELECT producto_id, nombre, descripcion, nombre_categoria, precio, stock, COALESCE(imagen_url, '') FROM v_productos WHERE nombre = $1",
		nombre,
	).Scan(&p.ProductoID, &p.Nombre, &p.Descripcion, &p.NombreCategoria, &p.Precio, &p.Stock, &p.ImagenURL)

	if err != nil {
		return Productos{}, fmt.Errorf("producto no encontrado: %w", err)
	}

	return p, nil
}

// 5. Inserta la cabecera de una transacción de pago en la tabla "transacciones".
func InsertarTransaccion(
	transaccionID string,
	tipoDocumento string,
	numeroDocumento string,
	tipoCuenta string,
	cuentaOTelefono string,
	bancoOrigen string,
	montoFinalUSD float64,
	montoFinalVES float64,
	tasaCambio float64,
	estadoTransaccion string,
) error {
	_, err := DB.Exec(context.Background(), `
		INSERT INTO transacciones
			(transaccion_id, tipo_documento, numero_documento, tipo_cuenta,
			 cuenta_o_telefono, banco_origen, monto_final_usd, monto_final_ves,
			 tasa_cambio, estado_transaccion)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		transaccionID, tipoDocumento, numeroDocumento, tipoCuenta,
		cuentaOTelefono, bancoOrigen, montoFinalUSD, montoFinalVES,
		tasaCambio, estadoTransaccion,
	)
	if err != nil {
		return fmt.Errorf("error al insertar transacción %s: %w", transaccionID, err)
	}
	return nil
}

// 6. Inserta una línea de detalle (un producto dentro de una transacción).
func InsertarDetalle(transaccionID string, productoID int, cantidadProducto int) error {
	_, err := DB.Exec(
		context.Background(),
		"INSERT INTO detalles (transaccion_id, producto_id, cantidad_producto) VALUES ($1, $2, $3)",
		transaccionID, productoID, cantidadProducto,
	)
	if err != nil {
		return fmt.Errorf("error al insertar detalle de %s (producto %d): %w", transaccionID, productoID, err)
	}
	return nil
}

// 7. Guarda la referencia de Sypago (transaction_id) apenas la tenemos.
func ActualizarReferenciaSypago(transaccionID string, referenciaSypago string) error {
	_, err := DB.Exec(
		context.Background(),
		"UPDATE transacciones SET referencia_sypago = $1, fecha_actualizacion = now() WHERE transaccion_id = $2",
		referenciaSypago, transaccionID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar referencia_sypago de %s: %w", transaccionID, err)
	}
	return nil
}

// 8. Actualiza el estado final de una transacción (ACCP, RJCT, CANC, etc.)
func ActualizarEstadoTransaccion(transaccionID string, estado string, codigoRechazo string) error {
	_, err := DB.Exec(
		context.Background(),
		"UPDATE transacciones SET estado_transaccion = $1, codigo_rechazo = NULLIF($2, ''), fecha_actualizacion = now() WHERE transaccion_id = $3",
		estado, codigoRechazo, transaccionID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar estado de %s: %w", transaccionID, err)
	}
	return nil
}

// 9. Descuenta stock real cuando una transacción queda ACEPTADA (ACCP).
func DescontarStock(productoID int, cantidad int) error {
	resultado, err := DB.Exec(
		context.Background(),
		"UPDATE productos SET stock = stock - $1 WHERE producto_id = $2 AND stock >= $1",
		cantidad, productoID,
	)
	if err != nil {
		return fmt.Errorf("error al descontar stock del producto %d: %w", productoID, err)
	}
	if resultado.RowsAffected() == 0 {
		return fmt.Errorf("no se pudo descontar stock del producto %d (sin stock suficiente o no existe)", productoID)
	}
	return nil
}

// 10. Trae todas las transacciones guardadas, la más reciente primero.
func ObtenerTransaccionBD() ([]Transacciones, error) {
	rows, err := DB.Query(
		context.Background(),
		`SELECT transaccion_id, tipo_documento, numero_documento, tipo_cuenta,
		        cuenta_o_telefono, banco_origen, monto_final_usd, monto_final_ves,
		        tasa_cambio, estado_transaccion, COALESCE(referencia_sypago, ''),
		        fecha_creacion, fecha_actualizacion, COALESCE(codigo_rechazo, ''),
		        COALESCE(estado_reembolso, ''), COALESCE(referencia_reembolso, ''),
		        COALESCE(codigo_rechazo_reembolso, '')
		 FROM transacciones
		 ORDER BY fecha_creacion DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("error al consultar transacciones: %w", err)
	}
	defer rows.Close()

	var listaTransacciones []Transacciones

	for rows.Next() {
		var t Transacciones
		err := rows.Scan(
			&t.TransaccionID, &t.TipoDocumento, &t.NumeroDocumento, &t.TipoCuenta,
			&t.CuentaOTelefono, &t.BancoOrigen, &t.MontoFinalUSD, &t.MontoFinalVES,
			&t.TasaCambio, &t.EstadoTransaccion, &t.ReferenciaSypago,
			&t.FechaCreacion, &t.FechaActualizacion, &t.CodigoRechazo,
			&t.EstadoReembolso, &t.ReferenciaReembolso, &t.CodigoRechazoReembolso,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear fila: %w", err)
		}
		listaTransacciones = append(listaTransacciones, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error durante la iteración: %w", err)
	}

	return listaTransacciones, nil
}

// 11. Trae UNA sola transacción por su ID. La usa el flujo de reembolso
// para recuperar todos los datos del pagador original (cuenta, banco,
// documento) y el monto exacto que se cobró en su momento.
func ObtenerTransaccionPorID(transaccionID string) (Transacciones, error) {
	var t Transacciones

	err := DB.QueryRow(
		context.Background(),
		`SELECT transaccion_id, tipo_documento, numero_documento, tipo_cuenta,
		        cuenta_o_telefono, banco_origen, monto_final_usd, monto_final_ves,
		        tasa_cambio, estado_transaccion, COALESCE(referencia_sypago, ''),
		        fecha_creacion, fecha_actualizacion, COALESCE(codigo_rechazo, ''),
		        COALESCE(estado_reembolso, ''), COALESCE(referencia_reembolso, ''),
		        COALESCE(codigo_rechazo_reembolso, '')
		 FROM transacciones
		 WHERE transaccion_id = $1`,
		transaccionID,
	).Scan(
		&t.TransaccionID, &t.TipoDocumento, &t.NumeroDocumento, &t.TipoCuenta,
		&t.CuentaOTelefono, &t.BancoOrigen, &t.MontoFinalUSD, &t.MontoFinalVES,
		&t.TasaCambio, &t.EstadoTransaccion, &t.ReferenciaSypago,
		&t.FechaCreacion, &t.FechaActualizacion, &t.CodigoRechazo,
		&t.EstadoReembolso, &t.ReferenciaReembolso, &t.CodigoRechazoReembolso,
	)

	if err != nil {
		return Transacciones{}, fmt.Errorf("transacción no encontrada: %w", err)
	}

	return t, nil
}

// 12. Trae el detalle (productos + cantidades) de una transacción.
// La usa el reembolso para saber qué stock hay que reponer.
func ObtenerDetallePorTransaccion(transaccionID string) ([]DetalleTransaccion, error) {
	rows, err := DB.Query(
		context.Background(),
		"SELECT producto_id, cantidad_producto FROM detalles WHERE transaccion_id = $1",
		transaccionID,
	)
	if err != nil {
		return nil, fmt.Errorf("error al consultar detalle de %s: %w", transaccionID, err)
	}
	defer rows.Close()

	var listaDetalle []DetalleTransaccion

	for rows.Next() {
		var d DetalleTransaccion
		if err := rows.Scan(&d.ProductoID, &d.CantidadProducto); err != nil {
			return nil, fmt.Errorf("error al escanear detalle: %w", err)
		}
		listaDetalle = append(listaDetalle, d)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error durante la iteración: %w", err)
	}

	return listaDetalle, nil
}

// 13. Marca el inicio de un reembolso. La condición "estado_reembolso IS NULL"
// evita que se pueda iniciar dos veces el mismo reembolso (protección
// contra doble clic o dos peticiones casi simultáneas).
func IniciarReembolso(transaccionID string, referenciaReembolso string) error {
	resultado, err := DB.Exec(
		context.Background(),
		"UPDATE transacciones SET estado_reembolso = 'PEND', referencia_reembolso = $1, fecha_actualizacion = now() WHERE transaccion_id = $2 AND estado_reembolso IS NULL",
		referenciaReembolso, transaccionID,
	)
	if err != nil {
		return fmt.Errorf("error al iniciar reembolso de %s: %w", transaccionID, err)
	}
	if resultado.RowsAffected() == 0 {
		return fmt.Errorf("la transacción %s no existe o ya tiene un reembolso en curso", transaccionID)
	}
	return nil
}

// 14. Actualiza el estado del reembolso (PEND -> ACCP/RJCT/CANC) según
// lo que vaya reportando el polling contra Sypago.
func ActualizarEstadoReembolso(transaccionID string, estado string, codigoRechazo string) error {
	_, err := DB.Exec(
		context.Background(),
		"UPDATE transacciones SET estado_reembolso = $1, codigo_rechazo_reembolso = NULLIF($2, ''), fecha_actualizacion = now() WHERE transaccion_id = $3",
		estado, codigoRechazo, transaccionID,
	)
	if err != nil {
		return fmt.Errorf("error al actualizar estado de reembolso de %s: %w", transaccionID, err)
	}
	return nil
}

// 15. Repone stock cuando un reembolso queda ACEPTADO (ACCP). Es el
// inverso exacto de DescontarStock.
func AumentarStock(productoID int, cantidad int) error {
	_, err := DB.Exec(
		context.Background(),
		"UPDATE productos SET stock = stock + $1 WHERE producto_id = $2",
		cantidad, productoID,
	)
	if err != nil {
		return fmt.Errorf("error al aumentar stock del producto %d: %w", productoID, err)
	}
	return nil
}

// Función que consulta la bd y RETORNA los usuarios, en este caso por el correo.
func ObtenerUsuarioPorCorreo(correo string) (Usuarios, error) {
	var u Usuarios

	err := DB.QueryRow(
		context.Background(),
		`SELECT usuario_id, nombre, apellido, correo, contraseña, fecha_creacion, es_admin 
		 FROM usuarios 
		 WHERE correo = $1`,
		correo,
	).Scan(&u.UsuarioID, &u.Nombre, &u.Apellido, &u.Correo, &u.Contrasena, &u.FechaCreacion, &u.EsAdmin)

	if err != nil {
		return Usuarios{}, fmt.Errorf("usuario no encontrado: %w", err)
	}

	return u, nil
}
