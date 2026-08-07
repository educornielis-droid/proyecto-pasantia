package servidor

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type TasaCambio struct {
	Codigo               string    `json:"code"`
	FechaCarga           time.Time `json:"load_date"`
	Tasa                 float64   `json:"rate"`
	TasaDeFuncionamiento bool      `json:"is_operation_rate"`
}

func TdC(ruta *gin.Engine) {

	ruta.GET("/api/tasa", func(contexto *gin.Context) {

		url := "https://pruebas.api.sypago.net/api/v1/bank/bcv/rate?use_date_rate=true"
		respuesta, err := http.Get(url)
		if err != nil {
			contexto.JSON(http.StatusBadGateway, gin.H{
				"error": "Respuesta al comunicar con la API externa",
			})
			return
		}

		if respuesta.StatusCode != http.StatusOK {
			contexto.JSON(respuesta.StatusCode, gin.H{
				"error": "Respuesta no exitosa de la API externa",
			})
			return
		}

		var request []TasaCambio

		if err := json.NewDecoder(respuesta.Body).Decode(&request); err != nil {
			contexto.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error al decodificar la respuesta",
			})

			return
		}

		contexto.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"payload": request,
		})

	})

}

type CheckoutFrontRequest struct { //Datos que recibe el backend desde el front
	Monto           float64 `json:"monto"`
	Concepto        string  `json:"concepto"`
	NombreCliente   string  `json:"nombre_cliente"`
	TipoDocumento   string  `json:"tipo_documento"`
	NumeroDocumento string  `json:"numero_documento"`
}

type AccountInfo struct { //JSON que se le enviara a la API de Sypago ↓↓↓ ¦ ¦ ↓↓↓
	BankCode string `json:"bank_code"`
	Type     string `json:"type"`
	Number   string `json:"number"`
}

type AmountInfo struct {
	Type        string  `json:"type"`
	Amt         float64 `json:"amt"`
	Currency    string  `json:"currency"`
	MinAllowAmt float64 `json:"min_allow_amt"`
	MaxAllowAmt float64 `json:"max_allow_amt"`
	UseDayRate  bool    `json:"use_day_rate"`
}

type NotificationURLs struct {
	SuccessfulCallbackURL string `json:"sucessful_callback_url"`
	FailedCallbackURL     string `json:"failed_callback_url"`
	ReturnFrontEndURL     string `json:"return_front_end_url"`
	WebHookEndpoint       string `json:"web_hook_endpoint"`
}

type DocumentInfo struct {
	Type   string `json:"type"`
	Number string `json:"number"`
}

type ReceivingUser struct {
	Name         string       `json:"name"`
	DocumentInfo DocumentInfo `json:"document_info"`
	Account      AccountInfo  `json:"account"`
}

type SypagoPaylinkPayload struct {
	InternalID       string           `json:"internal_id"`
	GroupID          string           `json:"group_id"`
	Account          AccountInfo      `json:"account"`
	Amount           AmountInfo       `json:"amount"`
	Concept          string           `json:"concept"`
	NotificationURLs NotificationURLs `json:"notification_urls"`
	ReceivingUser    ReceivingUser    `json:"receiving_user"`
	Expiration       int              `json:"expiration"`
}

type SypagoPaylinkPayResponse struct { //Respuesta que entrega la API de SyPago
	TransactionID   string `json:"transaction_id"`
	PayLink         string `json:"pay_link"`
	OperationSecret string `json:"operation_secret"`
}

func CheckOut(ruta *gin.Engine) {

	ruta.POST("/api/checkout", func(contexto *gin.Context) {

		var solicitudFront CheckoutFrontRequest

		if err := contexto.ShouldBindJSON(&solicitudFront); err != nil {
			contexto.JSON(http.StatusBadRequest, gin.H{
				"error": "datos enviados por el usuario desde el front invalidos",
			})
			return
		}

		payloadSypago := SypagoPaylinkPayload{
			InternalID: fmt.Sprintf("%d", time.Now().UnixNano())[:12],
			GroupID:    "0326AB3008E8",
			Account: AccountInfo{
				BankCode: "0001",
				Type:     "CNTA",
				Number:   "00010174520100126130",
			},
			Amount: AmountInfo{
				Type:        "ALMM",
				Amt:         solicitudFront.Monto,
				Currency:    "VES",
				MinAllowAmt: 1,
				MaxAllowAmt: solicitudFront.Monto,
				UseDayRate:  false,
			},
			Concept: solicitudFront.Concepto,
			NotificationURLs: NotificationURLs{
				SuccessfulCallbackURL: "https://www.sypago.net/success",
				FailedCallbackURL:     "https://www.sypago.net/fail",
				ReturnFrontEndURL:     "https://www.sypago.net/return",
				WebHookEndpoint:       "https://www.sypago.net/notification",
			},
			ReceivingUser: ReceivingUser{
				Name: solicitudFront.NombreCliente,
				DocumentInfo: DocumentInfo{
					Type:   solicitudFront.TipoDocumento,
					Number: solicitudFront.NumeroDocumento,
				},
				Account: AccountInfo{
					BankCode: "0102",
					Type:     "CELE",
					Number:   "04140121877",
				},
			},
			Expiration: 300,
		}

		bodyBytes, err := json.Marshal(payloadSypago)

		if err != nil {
			contexto.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error al procesar la peticion interna",
			})
			return
		}

		url := "https://pruebas.api.sypago.net/api/v1/transaction/paylink"

	})

}
