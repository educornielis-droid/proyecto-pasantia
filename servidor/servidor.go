package servidor

import (
	"proyecto-golang/servidor/routes"

	"github.com/gin-gonic/gin"
)

func Main() {
	ruta := gin.Default()

	routes.SetupRoutes(ruta)
	TdC(ruta)
	Checkout(ruta)
	SolicitarOTP(ruta)
	ConfirmarOTP(ruta)
	EstadoTransaccion(ruta)
	ruta.Run(":8080")
}
