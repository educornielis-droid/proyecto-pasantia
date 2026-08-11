package servidor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

/* ============================================================
   SYPAGO SERVICE
   Todo lo relacionado con la API de Sypago vive en este único
   archivo: tasa de cambio, autenticación (token), solicitud de
   OTP y confirmación de débito con OTP.
   ============================================================ */

/* ------------------------------------------------------------
   1. TASA DE CAMBIO (tal cual ya la tenías funcionando)
------------------------------------------------------------- */

type TasaCambio struct {
	Codigo               string    `json:"code"`
	FechaCarga           time.Time `json:"load_date"`
	Tasa                 float64   `json:"rate"`
	TasaDeFuncionamiento bool      `json:"is_operation_rate"`
}

func TdC(ruta *gin.Engine) {
	ruta.GET("/api/tasa", func(contexto *gin.Context) {
		url := os.Getenv("SYPAGO_API_BASE_URL") + "/api/v1/bank/bcv/rate?use_date_rate=true"

		respuesta, err := http.Get(url)
		if err != nil {
			contexto.JSON(http.StatusBadGateway, gin.H{
				"error": "Respuesta al comunicar con la API externa",
			})
			return
		}
		defer respuesta.Body.Close()

		if respuesta.StatusCode != http.StatusOK {
			contexto.JSON(respuesta.StatusCode, gin.H{
				"error": "Respuesta no exitosa de la API externa",
			})
			return
		}

		var listaTasas []TasaCambio

		if err := json.NewDecoder(respuesta.Body).Decode(&listaTasas); err != nil {
			contexto.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error al decodificar la respuesta",
			})
			return
		}

		contexto.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"payload": listaTasas,
		})
	})
}

/* ------------------------------------------------------------
   2. AUTENTICACIÓN: GESTOR DE TOKEN
   (nadie fuera de este archivo llama a esto directamente,
   todas las funciones de aquí abajo usan obtenerTokenValido())
------------------------------------------------------------- */

type gestorTokenSypago struct {
	tokenActual string
	expiraEn    time.Time
	mutex       sync.Mutex
}

var gestorToken = &gestorTokenSypago{}

// Margen de seguridad: refrescamos el token un poco ANTES de que
// venza de verdad, para nunca usar uno que expire a mitad de una petición.
const margenSeguridadToken = 60 * time.Second

func obtenerTokenValido() (string, error) {
	gestorToken.mutex.Lock()
	defer gestorToken.mutex.Unlock()

	if gestorToken.tokenActual != "" && time.Now().Before(gestorToken.expiraEn.Add(-margenSeguridadToken)) {
		return gestorToken.tokenActual, nil
	}

	nuevoToken, duracionSegundos, err := solicitarNuevoTokenASypago()
	if err != nil {
		return "", err
	}

	gestorToken.tokenActual = nuevoToken
	gestorToken.expiraEn = time.Now().Add(time.Duration(duracionSegundos) * time.Second)

	fmt.Println("[Sypago] Nuevo token obtenido, expira:", gestorToken.expiraEn.Format(time.RFC3339))

	return gestorToken.tokenActual, nil
}

type solicitudTokenSypago struct {
	ClientID string `json:"client_id"`
	Secret   string `json:"secret"`
}

type respuestaTokenSypago struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

func solicitarNuevoTokenASypago() (token string, duracionSegundos int, err error) {
	urlAutenticacion := os.Getenv("SYPAGO_API_BASE_URL") + "/api/v1/auth/token"

	cuerpoSolicitud := solicitudTokenSypago{
		ClientID: os.Getenv("SYPAGO_CLIENT_ID"),
		Secret:   os.Getenv("SYPAGO_API_KEY"),
	}

	cuerpoJSON, err := json.Marshal(cuerpoSolicitud)
	if err != nil {
		return "", 0, fmt.Errorf("error al preparar la solicitud de token: %w", err)
	}

	peticion, err := http.NewRequest(http.MethodPost, urlAutenticacion, bytes.NewBuffer(cuerpoJSON))
	if err != nil {
		return "", 0, fmt.Errorf("error al crear la petición de token: %w", err)
	}
	peticion.Header.Set("Content-Type", "application/json")

	cliente := &http.Client{Timeout: 15 * time.Second}
	respuesta, err := cliente.Do(peticion)
	if err != nil {
		return "", 0, fmt.Errorf("error al contactar el endpoint de autenticación: %w", err)
	}
	defer respuesta.Body.Close()

	if respuesta.StatusCode < 200 || respuesta.StatusCode >= 300 {
		cuerpoError, _ := io.ReadAll(respuesta.Body)
		return "", 0, fmt.Errorf("Sypago respondió %d al autenticar: %s", respuesta.StatusCode, string(cuerpoError))
	}

	var cuerpoRespuesta respuestaTokenSypago
	if err := json.NewDecoder(respuesta.Body).Decode(&cuerpoRespuesta); err != nil {
		return "", 0, fmt.Errorf("error al leer la respuesta de autenticación: %w", err)
	}

	if cuerpoRespuesta.AccessToken == "" {
		return "", 0, fmt.Errorf("la respuesta de Sypago no incluyó access_token")
	}

	return cuerpoRespuesta.AccessToken, cuerpoRespuesta.ExpiresIn, nil
}

/* ------------------------------------------------------------
   3. TIPOS COMPARTIDOS DE LOS ENDPOINTS DE TRANSACCIÓN
------------------------------------------------------------- */

type cuentaSypago struct {
	BankCode string `json:"bank_code"`
	Type     string `json:"type"`
	Number   string `json:"number"`
}

type documentoInfo struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

type montoSypago struct {
	Amt      float64 `json:"amt"`
	Currency string  `json:"currency"`
}

/* ------------------------------------------------------------
   4. SOLICITAR OTP
   POST /api/v1/request/otp
------------------------------------------------------------- */

type solicitudOTPSypago struct {
	CreditorAccount     cuentaSypago  `json:"creditor_account"`
	DebitorDocumentInfo documentoInfo `json:"debitor_document_info"`
	DebitorAccount      cuentaSypago  `json:"debitor_account"`
	Amount              montoSypago   `json:"amount"`
}

type solicitudOTPDesdeFrontend struct {
	NombreCompleto  string `json:"nombre_completo"`
	TipoCuenta      string `json:"tipo_cuenta"`
	CodigoBanco     string `json:"codigo_banco"`
	NumeroCuenta    string `json:"numero_cuenta"`
	TipoDocumento   string `json:"tipo_documento"`
	NumeroDocumento string `json:"numero_documento"`
}

// SolicitarOTP registra la ruta que pide el código OTP a Sypago.
func SolicitarOTP(ruta *gin.Engine) {
	ruta.POST("/api/checkout/:idTransaccion/otp", func(contexto *gin.Context) {
		idTransaccion := contexto.Param("idTransaccion")

		transaccionActual, existe := obtenerTransaccion(idTransaccion)
		if !existe {
			contexto.JSON(http.StatusNotFound, gin.H{"error": "La transacción no existe o ya expiró"})
			return
		}

		var datosFormulario solicitudOTPDesdeFrontend
		if err := contexto.ShouldBindJSON(&datosFormulario); err != nil {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "Datos del formulario inválidos"})
			return
		}

		if datosFormulario.TipoCuenta != "CELE" && datosFormulario.TipoCuenta != "CNTA" {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "Tipo de cuenta inválido"})
			return
		}

		if datosFormulario.CodigoBanco == "" || datosFormulario.NumeroCuenta == "" ||
			datosFormulario.TipoDocumento == "" || datosFormulario.NumeroDocumento == "" {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "Faltan datos del formulario"})
			return
		}

		token, err := obtenerTokenValido()
		if err != nil {
			fmt.Println("[Sypago OTP] Error al obtener token:", err)
			contexto.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo autenticar con Sypago"})
			return
		}

		cuerpoSypago := solicitudOTPSypago{
			CreditorAccount: cuentaSypago{
				BankCode: os.Getenv("SYPAGO_MERCHANT_BANK_CODE"),
				Type:     "CNTA",
				Number:   os.Getenv("SYPAGO_MERCHANT_ACCOUNT_NUMBER"),
			},
			DebitorDocumentInfo: documentoInfo{
				Type:   datosFormulario.TipoDocumento,
				Number: datosFormulario.NumeroDocumento,
			},
			DebitorAccount: cuentaSypago{
				BankCode: datosFormulario.CodigoBanco,
				Type:     datosFormulario.TipoCuenta,
				Number:   datosFormulario.NumeroCuenta,
			},
			Amount: montoSypago{
				Amt:      transaccionActual.Total, // monto real, calculado por el servidor
				Currency: "VES",
			},
		}

		respuestaSypago, err := llamarSolicitarOTPSypago(token, cuerpoSypago)
		if err != nil {
			fmt.Println("[Sypago OTP] Error al solicitar OTP:", err)
			contexto.JSON(http.StatusBadGateway, gin.H{"error": "No se pudo solicitar el código OTP con el proveedor"})
			return
		}

		guardarDatosDebitoPendiente(idTransaccion, datosDebitoPendiente{
			NombreCompleto: datosFormulario.NombreCompleto,
			DocumentInfo: documentoInfo{
				Type:   datosFormulario.TipoDocumento,
				Number: datosFormulario.NumeroDocumento,
			},
			DebitorAccount: cuentaSypago{
				BankCode: datosFormulario.CodigoBanco,
				Type:     datosFormulario.TipoCuenta,
				Number:   datosFormulario.NumeroCuenta,
			},
		})

		contexto.JSON(http.StatusOK, gin.H{
			"mensaje": "Código OTP solicitado. Revisa SMS, correo o tu app bancaria.",
			"detalle": respuestaSypago,
		})
	})
}

func llamarSolicitarOTPSypago(token string, cuerpo solicitudOTPSypago) (map[string]interface{}, error) {
	return llamarSypagoConToken(token, "/api/v1/request/otp", cuerpo)
}

/* ------------------------------------------------------------
   5. CONFIRMAR DÉBITO CON OTP
   POST /api/v1/transaction/otp
------------------------------------------------------------- */

type notificationUrlsOTP struct {
	WebHookEndpoint string `json:"web_hook_endpoint"`
}

type receivingUserOTP struct {
	Name         string        `json:"name"`
	Otp          string        `json:"otp"`
	DocumentInfo documentoInfo `json:"document_info"`
	Account      cuentaSypago  `json:"account"`
}

type solicitudDebitoOTP struct {
	InternalID       string              `json:"internal_id"`
	GroupID          string              `json:"group_id"`
	Account          cuentaSypago        `json:"account"`
	Amount           montoSypago         `json:"amount"`
	Concept          string              `json:"concept"`
	NotificationUrls notificationUrlsOTP `json:"notification_urls"`
	ReceivingUser    receivingUserOTP    `json:"receiving_user"`
}

type solicitudConfirmarOTPDesdeFrontend struct {
	Otp string `json:"otp"`
}

// ConfirmarOTP registra la ruta que envía el código que el usuario
// recibió, para completar el débito.
func ConfirmarOTP(ruta *gin.Engine) {
	ruta.POST("/api/checkout/:idTransaccion/confirmar-otp", func(contexto *gin.Context) {
		idTransaccion := contexto.Param("idTransaccion")

		transaccionActual, existe := obtenerTransaccion(idTransaccion)
		if !existe {
			contexto.JSON(http.StatusNotFound, gin.H{"error": "La transacción no existe o ya expiró"})
			return
		}

		if transaccionActual.DatosDebito == nil {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "Primero debes solicitar el código OTP"})
			return
		}

		var datosFrontend solicitudConfirmarOTPDesdeFrontend
		if err := contexto.ShouldBindJSON(&datosFrontend); err != nil || datosFrontend.Otp == "" {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "Debes indicar el código OTP"})
			return
		}

		token, err := obtenerTokenValido()
		if err != nil {
			fmt.Println("[Sypago Confirmar OTP] Error al obtener token:", err)
			contexto.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo autenticar con Sypago"})
			return
		}

		idInterno, err := generarIDTransaccion()
		if err != nil {
			contexto.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar identificador de operación"})
			return
		}

		cuerpoSypago := solicitudDebitoOTP{
			InternalID: idInterno,
			GroupID:    os.Getenv("SYPAGO_GROUP_ID"),
			Account: cuentaSypago{
				BankCode: os.Getenv("SYPAGO_MERCHANT_BANK_CODE"),
				Type:     "CNTA",
				Number:   os.Getenv("SYPAGO_MERCHANT_ACCOUNT_NUMBER"),
			},
			Amount: montoSypago{
				Amt:      transaccionActual.Total,
				Currency: "VES",
			},
			Concept: "Sypago Store - Orden " + idTransaccion,
			NotificationUrls: notificationUrlsOTP{
				WebHookEndpoint: os.Getenv("SYPAGO_WEBHOOK_URL"),
			},
			ReceivingUser: receivingUserOTP{
				Name:         transaccionActual.DatosDebito.NombreCompleto,
				Otp:          datosFrontend.Otp,
				DocumentInfo: transaccionActual.DatosDebito.DocumentInfo,
				Account:      transaccionActual.DatosDebito.DebitorAccount,
			},
		}

		respuestaSypago, err := llamarSypagoConToken(token, "/api/v1/transaction/otp", cuerpoSypago)
		if err != nil {
			fmt.Println("[Sypago Confirmar OTP] Error:", err)
			contexto.JSON(http.StatusBadGateway, gin.H{"error": "No se pudo confirmar el pago con el proveedor"})
			return
		}

		// TODO: marcar la transacción como "pagada" en tu futura tabla de
		// órdenes en PostgreSQL, y descontar el stock real de cada producto
		// (puedes usar database.ObtenerProductoPorNombre + un UPDATE de stock).

		contexto.JSON(http.StatusOK, gin.H{
			"mensaje": "Pago confirmado",
			"detalle": respuestaSypago,
		})
	})
}

/* ------------------------------------------------------------
   6. FUNCIÓN GENÉRICA PARA LLAMAR A SYPAGO CON TOKEN
   (la reutilizan solicitar-otp y confirmar-otp; evita repetir
   la misma lógica de armar el request HTTP dos veces)
------------------------------------------------------------- */

func llamarSypagoConToken(token string, rutaEndpoint string, cuerpo interface{}) (map[string]interface{}, error) {
	cuerpoJSON, err := json.Marshal(cuerpo)
	if err != nil {
		return nil, fmt.Errorf("error al preparar la solicitud: %w", err)
	}

	urlCompleta := os.Getenv("SYPAGO_API_BASE_URL") + rutaEndpoint
	peticion, err := http.NewRequest(http.MethodPost, urlCompleta, bytes.NewBuffer(cuerpoJSON))
	if err != nil {
		return nil, fmt.Errorf("error al crear la petición: %w", err)
	}

	peticion.Header.Set("Content-Type", "application/json")
	peticion.Header.Set("Authorization", "Bearer "+token)

	cliente := &http.Client{Timeout: 20 * time.Second}
	respuesta, err := cliente.Do(peticion)
	if err != nil {
		return nil, fmt.Errorf("error al contactar a Sypago: %w", err)
	}
	defer respuesta.Body.Close()

	cuerpoRespuesta, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return nil, fmt.Errorf("error al leer la respuesta de Sypago: %w", err)
	}

	if respuesta.StatusCode < 200 || respuesta.StatusCode >= 300 {
		return nil, fmt.Errorf("Sypago respondió %d: %s", respuesta.StatusCode, string(cuerpoRespuesta))
	}

	var datosRespuesta map[string]interface{}
	if err := json.Unmarshal(cuerpoRespuesta, &datosRespuesta); err != nil {
		return nil, fmt.Errorf("la respuesta de Sypago no es un JSON válido: %w", err)
	}

	return datosRespuesta, nil
}
