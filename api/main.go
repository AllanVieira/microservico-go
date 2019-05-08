package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/allanvieira/microservico-go/api/app"
)

func main() {

	application, err := app.New()
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	http.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "x-requested-with")
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		//TODO
		//Lista os Arquivos que já foram enviados e seus status atraves do metodo GET
		case http.MethodGet:
			json.NewEncoder(w).Encode("GET")
		//Upload de Arquivo através do metodo POST
		case http.MethodPost:
			app.UploadFile(r)
			//Inicia o parse do arquivo em um nova thread
			go app.ParseFile(application)
			json.NewEncoder(w).Encode("File uploaded successfully!.")
		default:
			fmt.Fprintf(w, "Algo deu errado", r.URL.Path)
		}
	})

	http.ListenAndServe(":8080", nil)
}
