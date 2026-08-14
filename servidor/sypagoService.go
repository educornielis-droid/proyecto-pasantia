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

/* ------------------------------------------------------------
   1. TASA DE CAMBIO
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
------------------------------------------------------------- */

type gestorTokenSypago struct {
	tokenActual string
	expiraEn    time.Time
	mutex       sync.Mutex
}

var gestorToken = &gestorTokenSypago{}

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
	TipoCuenta      string `json:"tipo_cuenta"`
	CodigoBanco     string `json:"codigo_banco"`
	NumeroCuenta    string `json:"numero_cuenta"`
	TipoDocumento   string `json:"tipo_documento"`
	NumeroDocumento string `json:"numero_documento"`
}

const nombreGenericoPagador = "Cliente Sypago Store"

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
			NombreCompleto: nombreGenericoPagador,
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

func llamarSolicitarOTPSypago(token string, cuerpo solicitudOTPSypago) (interface{}, error) {
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
	GroupID          string              `json:"group_id,omitempty"`
	Account          cuentaSypago        `json:"account"`
	Amount           montoSypago         `json:"amount"`
	Concept          string              `json:"concept"`
	NotificationUrls notificationUrlsOTP `json:"notification_urls"`
	ReceivingUser    receivingUserOTP    `json:"receiving_user"`
}

type solicitudConfirmarOTPDesdeFrontend struct {
	Otp string `json:"otp"`
}

// Tipado real de la respuesta de éxito de /api/v1/transaction/otp
type respuestaDebitoOTP struct {
	TransactionID   string `json:"transaction_id"`
	OperationSecret string `json:"operation_secret"`
}

// Confirmacion de que Sypago aceptó procesar la solicitud de débito.
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
			GroupID:    os.Getenv("SYPAGO_GROUP_ID"), // vacío = se omite del JSON, no manejan lotes
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

		respuestaSypago, err := llamarConfirmarDebitoOTPSypago(token, cuerpoSypago)
		if err != nil {
			fmt.Println("[Sypago Confirmar OTP] Error:", err)
			contexto.JSON(http.StatusBadGateway, gin.H{"error": "No se pudo enviar la solicitud de pago al proveedor"})
			return
		}

		// guardar la orden en PostgreSQL con estado "PEND" aquí.

		guardarResultadoPago(idTransaccion, resultadoPago{
			TransactionID:   respuestaSypago.TransactionID,
			OperationSecret: respuestaSypago.OperationSecret,
			Estado:          "PEND",
		})

		// OJO: devolvemos 202 (Accepted), no 200 con "éxito". La compra
		// TODAVÍA no está confirmada, apenas se puso en proceso.
		contexto.JSON(http.StatusAccepted, gin.H{
			"mensaje":        "Solicitud de pago enviada, verificando con el banco...",
			"transaction_id": respuestaSypago.TransactionID,
			"estado":         "PEND",
		})
	})
}

func llamarConfirmarDebitoOTPSypago(token string, cuerpo solicitudDebitoOTP) (respuestaDebitoOTP, error) {
	var resultadoVacio respuestaDebitoOTP

	cuerpoJSON, err := json.Marshal(cuerpo)
	if err != nil {
		return resultadoVacio, fmt.Errorf("error al preparar la solicitud de débito: %w", err)
	}

	urlCompleta := os.Getenv("SYPAGO_API_BASE_URL") + "/api/v1/transaction/otp"
	peticion, err := http.NewRequest(http.MethodPost, urlCompleta, bytes.NewBuffer(cuerpoJSON))
	if err != nil {
		return resultadoVacio, fmt.Errorf("error al crear la petición: %w", err)
	}

	peticion.Header.Set("Content-Type", "application/json")
	peticion.Header.Set("Authorization", "Bearer "+token)

	cliente := &http.Client{Timeout: 20 * time.Second}
	respuesta, err := cliente.Do(peticion)
	if err != nil {
		return resultadoVacio, fmt.Errorf("error al contactar a Sypago: %w", err)
	}
	defer respuesta.Body.Close()

	cuerpoRespuesta, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return resultadoVacio, fmt.Errorf("error al leer la respuesta de Sypago: %w", err)
	}

	fmt.Println("[Sypago] Respuesta cruda de /api/v1/transaction/otp :", string(cuerpoRespuesta))

	if respuesta.StatusCode < 200 || respuesta.StatusCode >= 300 {
		return resultadoVacio, fmt.Errorf("Sypago respondió %d: %s", respuesta.StatusCode, string(cuerpoRespuesta))
	}

	var datosRespuesta respuestaDebitoOTP
	if err := json.Unmarshal(cuerpoRespuesta, &datosRespuesta); err != nil {
		return resultadoVacio, fmt.Errorf("la respuesta de Sypago no es un JSON válido: %w", err)
	}

	if datosRespuesta.TransactionID == "" {
		return resultadoVacio, fmt.Errorf("la respuesta de Sypago no incluyó transaction_id")
	}

	return datosRespuesta, nil
}

/* ============================================================
   6. CONSULTA DE ESTADO (POLLING)
   GET /api/checkout/:idTransaccion/estado
   ============================================================ */

// Estados que ya no van a cambiar más - el polling se detiene aquí.
func esEstadoDefinitivo(estado string) bool {
	switch estado {
	case "ACCP", "RJCT", "CANC":
		return true
	default:
		return false
	}
}

// Tabla de códigos de rechazo de Sypago, para poder explicarle al usuario por qué falló un pago.
var descripcionesCodigoRechazo = map[string]string{
	"AB01":  "Proceso cancelado debido al tiempo de espera.",
	"AB07":  "El agente del mensaje no está en línea.",
	"AB08":  "SyCloud no puede comunicarse con el Gateway de la IBP.",
	"AC00":  "Operación en espera de respuesta del receptor.",
	"AC01":  "El número de cuenta no es válido o falta.",
	"AC04":  "El número de cuenta se encuentra cancelado por parte del Banco Receptor.",
	"AC06":  "La cuenta especificada está bloqueada.",
	"AC09":  "Moneda no válida o no existe.",
	"ACCP":  "Operación aceptada.",
	"AG01":  "Transacción restringida en este tipo de cuenta.",
	"AG09":  "Pago no recibido.",
	"AG10":  "El agente de mensaje está suspendido del sistema de pago nacional.",
	"AM02":  "El monto de la transacción no cumple con el acuerdo establecido.",
	"AM03":  "El monto especificado se encuentra en una moneda no definida en los acuerdos establecidos.",
	"AM04":  "Fondos insuficientes para cubrir el monto especificado.",
	"AM05":  "Operación duplicada.",
	"BE01":  "Datos del cliente emisor o receptor no se corresponden.",
	"BE20":  "La longitud del nombre supera el máximo permitido.",
	"CANC":  "Operación cancelada por el usuario.",
	"CH20":  "Número de decimales supera el máximo permitido.",
	"CUST":  "Cancelación solicitada por el deudor.",
	"DS02":  "Operación cancelada por usuario autorizado.",
	"DT03":  "Fecha de procesamiento no bancaria o no válida.",
	"DU01":  "La identificación de mensaje está duplicada.",
	"ED05":  "La transacción de liquidación ha fallado.",
	"EX01":  "Operación cancelada por expiración.",
	"FF05":  "El código del producto es inválido o no existe.",
	"FF07":  "El código del sub producto es inválido o no existe.",
	"MBE01": "El cliente pagador no se encuentra afiliado al servicio de cobro inmediato.",
	"MD01":  "El cliente acreedor no está afiliado por el cliente deudor.",
	"MD09":  "El cliente acreedor se encuentra en estado inactivo en la lista del cliente deudor.",
	"MD15":  "La cantidad a cobrar supera el monto establecido por el cliente deudor.",
	"MD21":  "La transacción a cobrar no cumple con los parámetros establecidos por el deudor.",
	"MD22":  "El cliente acreedor se encuentra suspendido por el cliente deudor.",
	"PEND":  "Operación en estatus pendiente.",
	"PROC":  "Operación en proceso.",
	"RC08":  "El código del banco no existe en el sistema de compensación/liquidación.",
	"RJCT":  "Operación rechazada.",
	"TE11":  "Operación cancelada por error de conexión con el banco. Válida para reintentar.",
	"TE28":  "Rechazo por validación específica de formato.",
	"TE29":  "Rechazo técnico del plugin bancario.",
	"TKCM":  "Código único de operación de aceptación de débito incorrecto.",
	"TM01":  "Mensaje enviado fuera del horario establecido.",
	"US03":  "Operación cancelada por error de conexión con el banco. Válida para reintentar.",
	"VE01":  "Rechazo técnico.",
	"WAIT":  "Operación en espera de validación de código.",
}

func EstadoTransaccion(ruta *gin.Engine) {
	ruta.GET("/api/checkout/:idTransaccion/estado", func(contexto *gin.Context) {
		idTransaccion := contexto.Param("idTransaccion")

		transaccionActual, existe := obtenerTransaccion(idTransaccion)
		if !existe {
			contexto.JSON(http.StatusNotFound, gin.H{"error": "Transacción no encontrada"})
			return
		}

		if transaccionActual.Pago == nil {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "Todavía no se ha enviado la solicitud de pago"})
			return
		}

		// Si ya tenemos un estado definitivo guardado, no hace falta
		// volver a consultar a Sypago - lo devolvemos directo.
		if esEstadoDefinitivo(transaccionActual.Pago.Estado) {
			contexto.JSON(http.StatusOK, gin.H{
				"estado":         transaccionActual.Pago.Estado,
				"transaction_id": transaccionActual.Pago.TransactionID,
			})
			return
		}

		token, err := obtenerTokenValido()
		if err != nil {
			fmt.Println("[Sypago Estado] Error al obtener token:", err)
			contexto.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo autenticar con Sypago"})
			return
		}

		nuevoEstado, codigoRechazo, err := consultarEstadoEnSypago(token, transaccionActual.Pago.TransactionID)
		if err != nil {
			fmt.Println("[Sypago Estado] Error al consultar estado:", err)
			contexto.JSON(http.StatusBadGateway, gin.H{"error": "No se pudo consultar el estado con el proveedor"})
			return
		}

		guardarResultadoPago(idTransaccion, resultadoPago{
			TransactionID:   transaccionActual.Pago.TransactionID,
			OperationSecret: transaccionActual.Pago.OperationSecret,
			Estado:          nuevoEstado,
		})

		descripcion := descripcionesCodigoRechazo[codigoRechazo]
		if descripcion != "" {
			fmt.Println("[Sypago Estado] Código:", codigoRechazo, "-", descripcion)
		}

		contexto.JSON(http.StatusOK, gin.H{
			"estado":         nuevoEstado,
			"codigo_rechazo": codigoRechazo,
			"descripcion":    descripcion,
			"transaction_id": transaccionActual.Pago.TransactionID,
		})
	})
}

// Consulta el estado real de una transacción en Sypago.
// GET /api/v1/transaction/{id}

func consultarEstadoEnSypago(token string, transactionID string) (estado string, codigoRechazo string, err error) {
	urlCompleta := os.Getenv("SYPAGO_API_BASE_URL") + "/api/v1/transaction/" + transactionID

	peticion, err := http.NewRequest(http.MethodGet, urlCompleta, nil)
	if err != nil {
		return "", "", fmt.Errorf("error al crear la petición de estado: %w", err)
	}
	peticion.Header.Set("Authorization", "Bearer "+token)

	cliente := &http.Client{Timeout: 15 * time.Second}
	respuesta, err := cliente.Do(peticion)
	if err != nil {
		return "", "", fmt.Errorf("error al contactar a Sypago: %w", err)
	}
	defer respuesta.Body.Close()

	cuerpoRespuesta, err := io.ReadAll(respuesta.Body)
	if err != nil {
		return "", "", fmt.Errorf("error al leer la respuesta de Sypago: %w", err)
	}

	fmt.Println("[Sypago Estado] Respuesta cruda de /api/v1/transaction/"+transactionID+":", string(cuerpoRespuesta))

	if respuesta.StatusCode < 200 || respuesta.StatusCode >= 300 {
		return "", "", fmt.Errorf("Sypago respondió %d: %s", respuesta.StatusCode, string(cuerpoRespuesta))
	}

	var datosCrudos map[string]interface{}
	if err := json.Unmarshal(cuerpoRespuesta, &datosCrudos); err != nil {
		return "", "", fmt.Errorf("la respuesta de Sypago no es un JSON válido: %w", err)
	}

	estadoEncontrado := extraerPrimerCampoString(datosCrudos, "status", "estado", "state")
	codigoEncontrado := extraerPrimerCampoString(datosCrudos, "rejected_code", "RejectedCode", "reject_code", "codigo_rechazo")

	if estadoEncontrado == "" {
		return "", "", fmt.Errorf("no se encontró el campo de estado en la respuesta: %s", string(cuerpoRespuesta))
	}

	return estadoEncontrado, codigoEncontrado, nil
}

func extraerPrimerCampoString(datos map[string]interface{}, posiblesClaves ...string) string {
	for _, clave := range posiblesClaves {
		if valor, existe := datos[clave]; existe {
			if valorTexto, esTexto := valor.(string); esTexto {
				return valorTexto
			}
		}
	}
	return ""
}

/* ------------------------------------------------------------
   6. FUNCIÓN GENÉRICA PARA LLAMAR A SYPAGO CON TOKEN
   (la reutilizan solicitar-otp y confirmar-otp; evita repetir
   la misma lógica de armar el request HTTP dos veces)
------------------------------------------------------------- */

func llamarSypagoConToken(token string, rutaEndpoint string, cuerpo interface{}) (interface{}, error) {
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

	fmt.Println("[Sypago] Respuesta cruda de", rutaEndpoint, ":", string(cuerpoRespuesta))

	if respuesta.StatusCode < 200 || respuesta.StatusCode >= 300 {
		return nil, fmt.Errorf("Sypago respondió %d: %s", respuesta.StatusCode, string(cuerpoRespuesta))
	}

	var datosRespuesta interface{}
	if err := json.Unmarshal(cuerpoRespuesta, &datosRespuesta); err != nil {
		return string(cuerpoRespuesta), nil
	}

	return datosRespuesta, nil
}
