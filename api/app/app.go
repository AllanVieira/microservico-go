package app

import (
	"bufio"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Nhanderu/brdoc"
	"github.com/allanvieira/microservico-go/api/db"
	"github.com/allanvieira/microservico-go/api/model"
)

// App : Struct contento propriedade globais da aplicacao
type Application struct {
	Database *sql.DB
}

// New : Funcao que retorna uma nova aplicacao com uma conexao com a base de dados
func New() (Application, error) {
	db, err := db.New()
	if err != nil {
		return Application{}, err
	}
	var app = Application{}
	app.Database = db
	return app, nil
}

//Escreve arquivo do Request em um arquivo na pasta files
func UploadFile(r *http.Request) error {
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	handler.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile("./files/file.text", data, 0666)
	if err != nil {
		return err
	}
	return nil
}

//ETL do Arquivo Texto para uma coleção de Struct Cliente
func ParseFile(app Application) error {
	fmt.Println(fmt.Sprintf("Inicio parse Arquivo: %v", time.Now()))
	file, err := os.Open("./files/file.text")
	if err != nil {
		return err
	}
	defer file.Close()

	var clientes []model.Cliente

	//Inicia a leitura de cada linha do arquivo
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		runes := []rune(scanner.Text())
		//Escapa da primeira linha do aquivo
		//Realiza a leitura através de um padrão de escrita estilo CNAB
		if string(runes[0:3]) != "CPF" {
			cliente := model.Cliente{}

			cliente.Cpf = string(runes[0:14])
			cliente.CpfValido = brdoc.IsCPF(cliente.Cpf)

			private, err := strconv.ParseInt(string(runes[19:20]), 10, 64)
			if err != nil {
				return err
			}
			cliente.Private = int(private)

			incompleto, err := strconv.ParseInt(string(runes[31:32]), 10, 64)
			if err != nil {
				return err
			}
			cliente.Incompleto = int(incompleto)

			if string(runes[43:47]) != "NULL" {
				dataUltimaCompra, err := time.Parse("2006-01-02", string(runes[43:53]))
				if err != nil {
					return err
				}
				cliente.DataUltimaCompra = dataUltimaCompra
			}

			if string(runes[65:69]) != "NULL" {
				ticketMedio, err := strconv.ParseFloat(strings.Replace(strings.Replace(string(runes[65:85]), " ", "", -1), ",", ".", -1), 64)
				if err != nil {
					return err
				}
				cliente.TicketMedio = ticketMedio
			}

			if string(runes[87:91]) != "NULL" {
				ticketUltimaCompra, err := strconv.ParseFloat(strings.Replace(strings.Replace(string(runes[87:110]), " ", "", -1), ",", ".", -1), 64)
				if err != nil {
					return err
				}
				cliente.TicketUltimaCompra = ticketUltimaCompra
			}

			if string(runes[111:115]) != "NULL" {
				cliente.LojaMaisFrequente = string(runes[111:129])
				cliente.LojaMaisFrequenteValido = brdoc.IsCNPJ(cliente.LojaMaisFrequente)
			} else {
				cliente.LojaMaisFrequenteValido = false
			}

			if string(runes[131:135]) != "NULL" {
				cliente.LojaUltimaCompra = string(runes[131:149])
				cliente.LojaUltimaCompraValido = brdoc.IsCNPJ(cliente.LojaUltimaCompra)
			} else {
				cliente.LojaUltimaCompraValido = false
			}
			clientes = append(clientes, cliente)
		}
	}

	errInsert := insertRows(app, clientes)
	if errInsert != nil {
		return errInsert
	}
	return nil
}

//Insere uma lista de struct Cientes na Base de Dados
func insertRows(app Application, clientes []model.Cliente) error {
	fmt.Println(fmt.Sprintf("Clientes Encontrados: %v", len(clientes)))
	fmt.Println(fmt.Sprintf("Inicio montagem query: %v", time.Now()))

	app.Database.Exec("DELETE FROM CLIENTES")
	query := "INSERT INTO CLIENTES (cpf, cpf_valido, private, incompleto, data_ultima_compra, ticket_medio, ticket_ultima_compra, loja_mais_frequente, loja_mais_frequente_valido, loja_ultima_compra, loja_ultima_compra_valido, created_at) VALUES "

	for idx, cliente := range clientes {
		row := fmt.Sprintf(" ('%v', %v, %v, %v, '%v', %v, %v, '%v', %v, '%v', %v, current_timestamp) ",
			cliente.Cpf, cliente.CpfValido, cliente.Private, cliente.Incompleto, cliente.DataUltimaCompra.Format("2006-01-02"), cliente.TicketMedio, cliente.TicketUltimaCompra, cliente.LojaMaisFrequente, cliente.LojaMaisFrequenteValido, cliente.LojaUltimaCompra, cliente.LojaUltimaCompraValido)
		query += row
		if idx < len(clientes)-1 {
			query += ","
		}
	}
	fmt.Println(fmt.Sprintf("Inicio insert: %v", time.Now()))
	result, err := app.Database.Exec(query)
	if err != nil {
		return err
	}
	result.LastInsertId()

	fmt.Println(fmt.Sprintf("Fim: %v", time.Now()))
	return nil
}
