package routes

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

type Usuarios struct {
	Nombre string `json:"nombre"`
	Correo string `json:"correo"`
}

var usuario []Usuarios

func SetupRoutes(ruta *gin.Engine) {

	ruta.LoadHTMLGlob("templates/*.html") //que solo busque archivos html

	ruta.Static("/static", "./static") //parte del diseño o funcionalidades del front

	ruta.GET("/", func(contexto *gin.Context) {
		contexto.HTML(http.StatusOK, "index.html", gin.H{
			"Heading": "Pagina principal",
		})
	})
	// ruta.GET("/", func(contexto *gin.Context) {
	// 	contexto.String(http.StatusOK, "Hola, mundo!")
	// })

	ruta.GET("/:pagina", func(contexto *gin.Context) {

		pagina := contexto.Param("pagina")

		if !strings.HasSuffix(pagina, ".html") {
			pagina += ".html"
		}

		if _, err := os.Stat("templates/" + pagina); err == nil {
			contexto.HTML(http.StatusOK, pagina, nil)
		} else {
			contexto.HTML(http.StatusNotFound, "404.html", nil)
		}
	})

	ruta.GET("/saludo/:nombre", func(contexto *gin.Context) {
		nombre := contexto.Param("nombre")
		contexto.String(http.StatusOK, "Hola %s, bienvenido", nombre)
	})

	ruta.POST("/usuarios", func(contexto *gin.Context) {
		var nuevoUsuario Usuarios

		if err := contexto.BindJSON(&nuevoUsuario); err != nil {
			fmt.Printf("\nError decodificando body: %v", err)
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "Error al decodificar el JSON"})
			return
		}

		if nuevoUsuario.Nombre == "" || nuevoUsuario.Correo == "" {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "Nombre y correo electronico son campos requeridos"})
			return
		}

		usuario = append(usuario, nuevoUsuario)

		contexto.JSON(http.StatusOK, gin.H{"mensaje": "Usuario registrado", "datos": usuario})
	})

	ruta.GET("/usuarios", func(contexto *gin.Context) {
		if usuario == nil {
			contexto.JSON(http.StatusOK, gin.H{
				"datos": "No se encontraron registros de usuarios actualmente",
			})
			return
		}
		contexto.JSON(http.StatusOK, gin.H{
			"datos": usuario,
		})
	})

}
