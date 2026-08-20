package servidor

import (
	"net/http"
	"proyecto-golang/database"

	"github.com/gin-gonic/gin"
)

// Struct para la validacion del perfil admin
type LoginInput struct {
	Correo     string `json:"correo" binding:"required"`
	Contrasena string `json:"contrasena" binding:"required"`
}

func SetupRoutes(ruta *gin.Engine) {

	ruta.LoadHTMLGlob("templates/*.html") //que solo busque archivos html

	ruta.Static("/static", "./static") //parte del diseño o funcionalidades del front

	//rutas publicas:

	ruta.GET("/app", func(contexto *gin.Context) {

		tasaUSD, err := ObtenerTasaCambioUSD()

		if err != nil {
			tasaUSD = 0.0
		}

		contexto.HTML(http.StatusOK, "index.html", gin.H{
			"Title": "Inicio - Sypago Store",
			"Logo":  "/static/img/sypago_spinner.svg",
			"Tasa":  tasaUSD,
		})
	})

	ruta.GET("/app/productos", func(contexto *gin.Context) {
		productos, err := database.ObtenerProductos()

		if err != nil {
			contexto.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error()})
			return
		}

		tasaUSD, err := ObtenerTasaCambioUSD()

		if err != nil {
			tasaUSD = 0.0
		}

		contexto.HTML(http.StatusOK, "productos.html", gin.H{
			"Title":     "Productos - Sypago Store",
			"Productos": productos,
			"Logo":      "/static/img/sypago_spinner.svg",
			"Tasa":      tasaUSD,
		})

	})

	ruta.GET("/app/login", func(contexto *gin.Context) {
		contexto.HTML(http.StatusOK, "login.html", gin.H{
			"Titulo": "Login - Sypago Store",
			"Logo":   "/static/img/sypago_spinner.svg",
		})
	})

	ruta.GET("/app/admin/login", func(contexto *gin.Context) {
		contexto.HTML(http.StatusOK, "adminLogin.html", gin.H{
			"Titulo": "Login admin - Sypago Store",
			"Logo":   "/static/img/sypago_spinner.svg",
		})
	})

	ruta.POST("/app/admin/login", LoginAdminHandler)

	ruta.GET("/app/admin/logout", LogoutAdminHandler) //llama a la funcion para cerrar sesion (elimina la cookie de sesion)

	//rutas protegidas para los administradores (por ahora solo hay uno)
	adminGrupo := ruta.Group("/app/admin")
	adminGrupo.Use(AuthAdminMiddleware())
	{
		adminGrupo.GET("/ordenes", func(contexto *gin.Context) {
			transacciones, err := database.ObtenerTransaccionBD()

			if err != nil {
				contexto.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error()})
				return
			}

			contexto.HTML(http.StatusOK, "transacciones.html", gin.H{
				"Titulo":        "Panel administrativo - Sypago Store",
				"Transacciones": transacciones,
				"Logo":          "/static/img/sypago_spinner.svg",
			})

		})
	}

	ruta.GET("/api/v1", func(contexto *gin.Context) {
		productos, err := database.ObtenerProductos()
		if err != nil {
			contexto.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error()})
			return
		}

		contexto.JSON(http.StatusOK, gin.H{
			"productos": productos,
		})

	})

	ruta.GET("/api/productos", func(contexto *gin.Context) {
		productos, err := database.ObtenerProductos()
		if err != nil {
			contexto.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error()})
			return
		}

		contexto.JSON(http.StatusOK, gin.H{
			"productos": productos,
		})

	})

	ruta.GET("/saludo/:nombre", func(contexto *gin.Context) {
		nombre := contexto.Param("nombre")
		contexto.String(http.StatusOK, "Hola %s, bienvenido", nombre)
	})

	// ruta.GET("/:pagina", func(contexto *gin.Context) {

	// 	pagina := contexto.Param("pagina")

	// 	if !strings.HasSuffix(pagina, ".html") {
	// 		pagina += ".html"
	// 	}

	// 	if _, err := os.Stat("templates/" + pagina); err == nil {
	// 		contexto.HTML(http.StatusOK, pagina, nil)
	// 	} else {
	// 		contexto.HTML(http.StatusNotFound, "404.html", nil)
	// 	}
	// })

	// ruta.POST("/usuarios", func(contexto *gin.Context) {
	// 	var nuevoUsuario Usuarios

	// 	if err := contexto.BindJSON(&nuevoUsuario); err != nil {
	// 		fmt.Printf("\nError decodificando body: %v", err)
	// 		contexto.JSON(http.StatusBadRequest, gin.H{"error": "Error al decodificar el JSON"})
	// 		return
	// 	}

	// 	if nuevoUsuario.Nombre == "" || nuevoUsuario.Correo == "" {
	// 		contexto.JSON(http.StatusBadRequest, gin.H{"error": "Nombre y correo electronico son campos requeridos"})
	// 		return
	// 	}

	// 	usuario = append(usuario, nuevoUsuario)

	// 	contexto.JSON(http.StatusOK, gin.H{"mensaje": "Usuario registrado", "datos": usuario})
	// })

	// ruta.GET("/usuarios", func(contexto *gin.Context) {
	// 	if usuario == nil {
	// 		contexto.JSON(http.StatusOK, gin.H{
	// 			"datos": "No se encontraron registros de usuarios actualmente",
	// 		})
	// 		return
	// 	}
	// 	contexto.JSON(http.StatusOK, gin.H{
	// 		"datos": usuario,
	// 	})
	// })

}

// funcion para verficar si eres admin registrado en la bd
func LoginAdminHandler(contexto *gin.Context) {
	var input LoginInput

	if err := contexto.ShouldBindJSON(&input); err != nil {
		contexto.JSON(http.StatusBadRequest, gin.H{"message": "Datos incompletos"})
		return
	}

	// 1. Consultar el usuario en la BD
	usuario, err := database.ObtenerUsuarioPorCorreo(input.Correo)
	if err != nil {
		// No revelamos si falló el correo o la clave por seguridad
		contexto.JSON(http.StatusUnauthorized, gin.H{"message": "Credenciales inválidas o sin autorización"})
		return
	}

	// 2. Verificar la contraseña
	// NOTA: Viendo tu captura de pgAdmin, las claves están en texto plano (ej: "admin123").
	// Si estás usando texto plano:
	if usuario.Contrasena != input.Contrasena {
		contexto.JSON(http.StatusUnauthorized, gin.H{"message": "Credenciales inválidas o sin autorización"})
		return
	}

	// 3. Validar si es administrador
	if !usuario.EsAdmin {
		contexto.JSON(http.StatusForbidden, gin.H{"message": "Credenciales inválidas o sin autorización"})
		return
	}

	// 4. Login correcto
	// Parámetros: nombre, valor, maxAge (segundos... 300s = 5min), path, domain, secure, httpOnly
	contexto.SetCookie("admin_session", "true", 300, "/", "", false, true)

	contexto.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Bienvenido al sistema",
	})
}

// autenticacion para el admin (debes iniciar sesion obligatoriamente para observar las transacciones)
func AuthAdminMiddleware() gin.HandlerFunc {
	return func(contexto *gin.Context) {
		// Leemos la cookie del navegador
		cookie, err := contexto.Cookie("admin_session")

		// Si la cookie no existe o no es válida, redirigimos al login
		if err != nil || cookie != "true" {
			contexto.Redirect(http.StatusSeeOther, "/app/admin/login")
			contexto.Abort() // Detiene la ejecución del resto de los handlers
			return
		}

		contexto.Next() // Si tiene la cookie, lo deja pasar
	}
}

// funcion que fuerza el cierre de sesion como admin
func LogoutAdminHandler(c *gin.Context) {
	// Elimina la cookie fijando maxAge en -1
	c.SetCookie("admin_session", "", -1, "/", "", false, true)

	// Redirige al login de inmediato
	c.Redirect(http.StatusSeeOther, "/app")
}
