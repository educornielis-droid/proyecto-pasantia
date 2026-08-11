package servidor

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"proyecto-golang/database"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

/* ============================================================
   CHECKOUT: crear la transacción a partir del carrito y
   mostrar checkout.html con esos datos.
   (Esto es lógica propia de la tienda, no de la API de Sypago,
   por eso vive separado de sypagoService.go)
   ============================================================ */

type productoSolicitado struct {
	Nombre   string `json:"nombre"`
	Cantidad int    `json:"cantidad"`
}

type solicitudCheckout struct {
	Productos []productoSolicitado `json:"productos"`
}

type productoTransaccion struct {
	Nombre    string
	ImagenURL string
	Cantidad  int
	Precio    float64
	Subtotal  float64
}

type transaccion struct {
	ID          string
	Productos   []productoTransaccion
	Total       float64
	CreadaEn    time.Time
	DatosDebito *datosDebitoPendiente
}

// Datos del pagador guardados al solicitar el OTP, para reutilizarlos
// al confirmar el código sin volver a confiar en el navegador.
type datosDebitoPendiente struct {
	NombreCompleto string
	DocumentInfo   documentoInfo
	DebitorAccount cuentaSypago
}

/* ------------------------------------------------------------
   ALMACENAMIENTO TEMPORAL EN MEMORIA
   (mañana esto se reemplaza por una tabla en PostgreSQL)
------------------------------------------------------------- */

var (
	almacenTransacciones = make(map[string]transaccion)
	mutexTransacciones   sync.Mutex
)

func guardarTransaccion(nuevaTransaccion transaccion) {
	mutexTransacciones.Lock()
	defer mutexTransacciones.Unlock()
	almacenTransacciones[nuevaTransaccion.ID] = nuevaTransaccion
}

func obtenerTransaccion(idTransaccion string) (transaccion, bool) {
	mutexTransacciones.Lock()
	defer mutexTransacciones.Unlock()
	transaccionEncontrada, existe := almacenTransacciones[idTransaccion]
	return transaccionEncontrada, existe
}

func guardarDatosDebitoPendiente(idTransaccion string, datos datosDebitoPendiente) bool {
	mutexTransacciones.Lock()
	defer mutexTransacciones.Unlock()

	transaccionExistente, existe := almacenTransacciones[idTransaccion]
	if !existe {
		return false
	}

	transaccionExistente.DatosDebito = &datos
	almacenTransacciones[idTransaccion] = transaccionExistente
	return true
}

/* ------------------------------------------------------------
   Checkout registra las 2 rutas de este archivo, siguiendo el
   mismo patrón que TdC() en sypagoService.go.
------------------------------------------------------------- */

func Checkout(ruta *gin.Engine) {

	// PASO 1: el carrito inicia la transacción
	ruta.POST("/api/checkout/iniciar", func(contexto *gin.Context) {
		var solicitud solicitudCheckout

		if err := contexto.ShouldBindJSON(&solicitud); err != nil {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "El cuerpo de la solicitud no es válido"})
			return
		}

		if len(solicitud.Productos) == 0 {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "El carrito está vacío"})
			return
		}

		var productosTransaccion []productoTransaccion
		montoTotal := 0.0

		for _, itemSolicitado := range solicitud.Productos {
			if itemSolicitado.Cantidad < 1 {
				contexto.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Cantidad inválida para %s", itemSolicitado.Nombre)})
				return
			}

			// Recalculamos SIEMPRE desde la BD, nunca confiamos en un
			// precio que pudiera venir del navegador.
			productoBD, err := database.ObtenerProductoPorNombre(itemSolicitado.Nombre)
			if err != nil {
				contexto.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Producto no encontrado: %s", itemSolicitado.Nombre)})
				return
			}

			if itemSolicitado.Cantidad > productoBD.Stock {
				contexto.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("No hay suficiente stock de %s", itemSolicitado.Nombre)})
				return
			}

			subtotal := productoBD.Precio * float64(itemSolicitado.Cantidad)
			montoTotal += subtotal

			productosTransaccion = append(productosTransaccion, productoTransaccion{
				Nombre:    productoBD.Nombre,
				ImagenURL: productoBD.ImagenURL,
				Cantidad:  itemSolicitado.Cantidad,
				Precio:    productoBD.Precio,
				Subtotal:  subtotal,
			})
		}

		if montoTotal <= 0 {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "El monto total debe ser mayor a cero"})
			return
		}

		idTransaccion, err := generarIDTransaccion()
		if err != nil {
			contexto.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar la transacción"})
			return
		}

		nuevaTransaccion := transaccion{
			ID:        idTransaccion,
			Productos: productosTransaccion,
			Total:     montoTotal,
			CreadaEn:  time.Now(),
		}

		guardarTransaccion(nuevaTransaccion)

		contexto.JSON(http.StatusOK, gin.H{
			"redirect_url": "/checkout/" + idTransaccion,
		})
	})

	// PASO 2: mostrar checkout.html con los datos guardados
	ruta.GET("/checkout/:idTransaccion", func(contexto *gin.Context) {
		idTransaccion := contexto.Param("idTransaccion")

		transaccionEncontrada, existe := obtenerTransaccion(idTransaccion)
		if !existe {
			contexto.Redirect(http.StatusFound, "/app/productos")
			return
		}

		contexto.HTML(http.StatusOK, "checkout.html", gin.H{
			"Title":          "Checkout - Sypago Store",
			"NombreComercio": "Sypago Store",
			"Rif":            "J-3090842500",
			"Productos":      transaccionEncontrada.Productos,
			"Total":          transaccionEncontrada.Total,
			"IDTransaccion":  transaccionEncontrada.ID,
		})
	})
}

/* ------------------------------------------------------------
   UTILIDAD: ID único de 12 caracteres hexadecimales
   (la usan tanto checkout.go como sypagoService.go)
------------------------------------------------------------- */

func generarIDTransaccion() (string, error) {
	bytesAleatorios := make([]byte, 6)
	if _, err := rand.Read(bytesAleatorios); err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(bytesAleatorios)), nil
}
