package servidor // AJUSTA si tu paquete real tiene otro nombre

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"proyecto-golang/database"
)

/* ============================================================
   REEMBOLSO — POST /admin/reembolso/:idTransaccion
   ============================================================
   Manda una solicitud de "crédito inmediato" a Sypago
   (POST /api/v1/transaction/credit), invirtiendo el sentido del
   dinero respecto al cobro original:
   - account (origen)          -> ahora es la cuenta del COMERCIO
   - receiving_user.account    -> ahora es la cuenta del CLIENTE
     original (la que quedó guardada en la transacción)
   El monto es el que se guardó en su momento (monto_final_ves),
   no el total recalculado hoy — es literalmente lo que se cobró.
   ============================================================ */

type receivingUserCredito struct {
	Name         string        `json:"name"`
	DocumentInfo documentoInfo `json:"document_info"`
	Account      cuentaSypago  `json:"account"`
}

type solicitudCreditoSypago struct {
	InternalID       string               `json:"internal_id"`
	GroupID          string               `json:"group_id,omitempty"`
	Account          cuentaSypago         `json:"account"`
	SubProduct       string               `json:"sub_product"`
	Amount           montoSypago          `json:"amount"`
	Concept          string               `json:"concept"`
	NotificationUrls notificationUrlsOTP  `json:"notification_urls"`
	ReceivingUser    receivingUserCredito `json:"receiving_user"`
}

func Reembolso(ruta *gin.Engine) {
	ruta.POST("/admin/reembolso/:idTransaccion", func(contexto *gin.Context) {
		idTransaccion := contexto.Param("idTransaccion")

		transaccionBD, err := database.ObtenerTransaccionPorID(idTransaccion)
		if err != nil {
			contexto.JSON(http.StatusNotFound, gin.H{"error": "Transacción no encontrada"})
			return
		}

		// Solo se puede reembolsar lo que realmente se cobró
		if transaccionBD.EstadoTransaccion != "ACCP" {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "Solo se pueden reembolsar transacciones aceptadas"})
			return
		}

		// Protección extra a nivel de aplicación (la definitiva vive en
		// database.IniciarReembolso, con la condición SQL). Esta evita
		// gastar una llamada a Sypago si ya sabemos que no corresponde.
		if transaccionBD.EstadoReembolso != "" {
			contexto.JSON(http.StatusBadRequest, gin.H{"error": "Esta transacción ya tiene un reembolso en curso o completado"})
			return
		}

		token, err := obtenerTokenValido()
		if err != nil {
			fmt.Println("[Sypago Reembolso] Error al obtener token:", err)
			contexto.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo autenticar con Sypago"})
			return
		}

		idInterno, err := generarIDTransaccion()
		if err != nil {
			contexto.JSON(http.StatusInternalServerError, gin.H{"error": "Error al generar identificador de operación"})
			return
		}

		cuerpoSypago := solicitudCreditoSypago{
			InternalID: idInterno,
			GroupID:    os.Getenv("SYPAGO_GROUP_ID"),
			Account: cuentaSypago{
				// Origen del dinero: la cuenta del comercio
				BankCode: os.Getenv("SYPAGO_MERCHANT_BANK_CODE"),
				Type:     "CNTA",
				Number:   os.Getenv("SYPAGO_MERCHANT_ACCOUNT_NUMBER"),
			},
			SubProduct: "220",
			Amount: montoSypago{
				// El monto que se cobró EN SU MOMENTO, no el de hoy
				Amt:      transaccionBD.MontoFinalVES,
				Currency: "VES",
			},
			Concept: "Reembolso Sypago Store - Orden " + idTransaccion,
			NotificationUrls: notificationUrlsOTP{
				WebHookEndpoint: os.Getenv("SYPAGO_WEBHOOK_URL"), // puede quedar vacío, Sypago lo permite
			},
			ReceivingUser: receivingUserCredito{
				// Destino del dinero: el cliente original de esta transacción
				Name: nombreGenericoPagador,
				DocumentInfo: documentoInfo{
					Type:   transaccionBD.TipoDocumento,
					Number: transaccionBD.NumeroDocumento,
				},
				Account: cuentaSypago{
					BankCode: transaccionBD.BancoOrigen,
					Type:     transaccionBD.TipoCuenta,
					Number:   transaccionBD.CuentaOTelefono,
				},
			},
		}

		respuestaSypago, err := llamarCreditoSypago(token, cuerpoSypago)
		if err != nil {
			fmt.Println("[Sypago Reembolso] Error:", err)
			contexto.JSON(http.StatusBadGateway, gin.H{"error": "No se pudo iniciar el reembolso con el proveedor"})
			return
		}

		if err := database.IniciarReembolso(idTransaccion, respuestaSypago.TransactionID); err != nil {
			fmt.Println("[BD] Error al registrar inicio de reembolso:", err)
			contexto.JSON(http.StatusInternalServerError, gin.H{"error": "El reembolso se envió a Sypago pero no se pudo registrar. Contacta soporte."})
			return
		}

		contexto.JSON(http.StatusOK, gin.H{
			"mensaje":        "Reembolso en proceso",
			"transaction_id": respuestaSypago.TransactionID,
		})
	})
}

func llamarCreditoSypago(token string, cuerpo solicitudCreditoSypago) (respuestaDebitoOTP, error) {
	var resultadoVacio respuestaDebitoOTP

	cuerpoJSON, err := json.Marshal(cuerpo)
	if err != nil {
		return resultadoVacio, fmt.Errorf("error al preparar la solicitud de crédito: %w", err)
	}

	urlCompleta := os.Getenv("SYPAGO_API_BASE_URL") + "/api/v1/transaction/credit"
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

	fmt.Println("[Sypago] Respuesta cruda de /api/v1/transaction/credit :", string(cuerpoRespuesta))

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
