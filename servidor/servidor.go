package servidor

import (
	"github.com/gin-gonic/gin"
)

func Main() {
	ruta := gin.Default()

	SetupRoutes(ruta)
	TdC(ruta)
	Checkout(ruta)
	SolicitarOTP(ruta)
	ConfirmarOTP(ruta)
	EstadoTransaccion(ruta)
	Reembolso(ruta)

	ruta.GET("/ping", func(contexto *gin.Context) {
		contexto.JSON(200, gin.H{
			"message": "pong",
		})
	})

	ruta.Run("0.0.0.0:8080")
}
