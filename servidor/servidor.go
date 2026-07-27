package servidor

// re-hacer este archivo!!!!!

import (
	"embed" // 1. Importas este paquete
	"fmt"
	"net/http"
	"text/template"
)

// 2. Esta directiva le dice a Go que guarde el contenido de la carpeta en la variable
//
//go:embed templates/*
var archivosPlantillas embed.FS

func Index(rw http.ResponseWriter, r *http.Request) {
	// 3. Usas ParseFS en lugar de ParseFiles
	tmpl, err := template.ParseFS(archivosPlantillas, "templates/index.html")
	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(rw, nil)
}

func Main() {
	http.HandleFunc("/", Index)
	fmt.Println("Servidor corriendo en el puerto 3000...")
	fmt.Println("Run server: http://localhost:3000")
	http.ListenAndServe(":3000", nil)
}
