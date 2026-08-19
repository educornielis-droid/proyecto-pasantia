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
	ruta.Run(":8080")
}
