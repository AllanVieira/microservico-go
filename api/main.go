package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

//Configuração Base de Dados
const (
	DB_HOST = "db" //trocar para local host quando rodar local
	//DB_HOST     = "localhost"
	DB_USER     = "postgres"
	DB_PASSWORD = "postgres"
	DB_NAME     = "files"
)

type Cliente struct {
	cpf                     string    `json:"cpf"`
	cpfValido               bool      `json:"cpf_valido"`
	private                 int       `json:"private"`
	incompleto              int       `json:"incompleto"`
	dataUltimaCompra        time.Time `json:"data_ultima_compra"`
	ticketMedio             float64   `json:"ticket_medio"`
	ticketUltimaCompra      float64   `json:"ticket_ultima_compra"`
	lojaMaisFrequente       string    `json:"loja_mais_frequente"`
	lojaMaisFrequenteValido bool      `json:"loja_mais_frequente_valido"`
	lojaUltimaCompra        string    `json:"loja_ultima_compra"`
	lojaUltimaCompraValido  bool      `json:"loja_ultima_compra_valido"`
}

func main() {

	// Cria conexão com o banco
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	//Cria tabelas
	res, err := db.Exec(" CREATE TABLE IF NOT EXISTS CLIENTES ( id serial, cpf text, cpf_valido bool, private int, incompleto int, data_ultima_compra date, ticket_medio numeric(18,2), ticket_ultima_compra numeric(18,2), loja_mais_frequente text, loja_mais_frequente_valido bool, loja_ultima_compra text, loja_ultima_compra_valido bool, created_at timestamp)")
	checkErr(err)
	fmt.Println(res)

	http.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "x-requested-with")
		switch r.Method {
		//TODO
		//Lista os Arquivos que já foram enviados e seus status atraves do metodo GET
		case http.MethodGet:
			responseJSON(w, "")
		//Upload de Arquivo através do metodo POST
		case http.MethodPost:
			uploadFile(r)
			//Inicia o parse do arquivo em um nova tread
			go parseFile()
			responseJSON(w, "File uploaded successfully!.")
		default:
			fmt.Fprintf(w, "Algo deu errado", r.URL.Path)
		}
	})

	http.ListenAndServe(":8080", nil)
}

func checkErr(err error) {
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
}

func responseJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

//Escreve arquivo do Request em um arquivo na pasta files
func uploadFile(r *http.Request) {
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	handler.Open()
	checkErr(err)
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	checkErr(err)

	err = ioutil.WriteFile("./files/file.text", data, 0666)
	checkErr(err)
}

//ETL do Arquivo Texto para uma coleção de Struct Cliente
func parseFile() {
	fmt.Println(fmt.Sprintf("Inicio parse Arquivo: %v", time.Now()))
	file, err := os.Open("./files/file.text")
	checkErr(err)
	defer file.Close()

	var clientes []Cliente

	//Inicia a leitura de cada linha do arquivo
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		runes := []rune(scanner.Text())
		//Escapa da primeira linha do aquivo
		//Realiza a leitura através de um padrão de escrita estilo CNAB
		if string(runes[0:3]) != "CPF" {
			cliente := Cliente{}

			cliente.cpf = string(runes[0:14])
			if ValidaCPF(cliente.cpf) != nil {
				cliente.cpfValido = false
			} else {
				cliente.cpfValido = true
			}

			private, err := strconv.ParseInt(string(runes[19:20]), 10, 64)
			checkErr(err)
			cliente.private = int(private)

			incompleto, err := strconv.ParseInt(string(runes[31:32]), 10, 64)
			checkErr(err)
			cliente.incompleto = int(incompleto)

			if string(runes[43:47]) != "NULL" {
				dataUltimaCompra, err := time.Parse("2006-01-02", string(runes[43:53]))
				checkErr(err)
				cliente.dataUltimaCompra = dataUltimaCompra
			}

			if string(runes[65:69]) != "NULL" {
				ticketMedio, err := strconv.ParseFloat(strings.Replace(strings.Replace(string(runes[65:85]), " ", "", -1), ",", ".", -1), 64)
				checkErr(err)
				cliente.ticketMedio = ticketMedio
			}

			if string(runes[87:91]) != "NULL" {
				ticketUltimaCompra, err := strconv.ParseFloat(strings.Replace(strings.Replace(string(runes[87:110]), " ", "", -1), ",", ".", -1), 64)
				checkErr(err)
				cliente.ticketUltimaCompra = ticketUltimaCompra
			}

			if string(runes[111:115]) != "NULL" {
				cliente.lojaMaisFrequente = string(runes[111:129])
				if validaCNPJ(cliente.lojaMaisFrequente) != nil {
					cliente.lojaMaisFrequenteValido = false
				} else {
					cliente.lojaMaisFrequenteValido = true
				}
			} else {
				cliente.lojaMaisFrequenteValido = false
			}

			if string(runes[131:135]) != "NULL" {
				cliente.lojaUltimaCompra = string(runes[131:149])
				if validaCNPJ(cliente.lojaUltimaCompra) != nil {
					cliente.lojaUltimaCompraValido = false
				} else {
					cliente.lojaUltimaCompraValido = true
				}
			} else {
				cliente.lojaUltimaCompraValido = false
			}
			clientes = append(clientes, cliente)
		}
	}

	insertRows(clientes)
}

//Insere uma lista de struct Cientes na Base de Dados
func insertRows(clientes []Cliente) {
	fmt.Println(fmt.Sprintf("Clientes Encontrados: %v", len(clientes)))
	fmt.Println(fmt.Sprintf("Inicio montagem query: %v", time.Now()))
	// Cria conexão com o banco
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	checkErr(err)
	defer db.Close()

	db.Exec("DELETE FROM CLIENTES")
	query := "INSERT INTO CLIENTES (cpf, cpf_valido, private, incompleto, data_ultima_compra, ticket_medio, ticket_ultima_compra, loja_mais_frequente, loja_mais_frequente_valido, loja_ultima_compra, loja_ultima_compra_valido, created_at) VALUES "

	for idx, cliente := range clientes {
		row := fmt.Sprintf(" ('%v', %v, %v, %v, '%v', %v, %v, '%v', %v, '%v', %v, current_timestamp) ",
			cliente.cpf, cliente.cpfValido, cliente.private, cliente.incompleto, cliente.dataUltimaCompra.Format("2006-01-02"), cliente.ticketMedio, cliente.ticketUltimaCompra, cliente.lojaMaisFrequente, cliente.lojaMaisFrequenteValido, cliente.lojaUltimaCompra, cliente.lojaUltimaCompraValido)
		query += row
		if idx < len(clientes)-1 {
			query += ","
		}
	}
	fmt.Println(fmt.Sprintf("Inicio insert: %v", time.Now()))
	result, err := db.Exec(query)
	checkErr(err)
	result.LastInsertId()

	fmt.Println(fmt.Sprintf("Fim: %v", time.Now()))

}

func ValidaCPF(cpf string) error {
	cpf = strings.Replace(cpf, ".", "", -1)
	cpf = strings.Replace(cpf, "-", "", -1)
	if len(cpf) != 11 {
		return errors.New("CPF inválido")
	}
	var eq bool
	var dig string
	for _, val := range cpf {
		if len(dig) == 0 {
			dig = string(val)
		}
		if string(val) == dig {
			eq = true
			continue
		}
		eq = false
		break
	}
	if eq {
		return errors.New("CPF inválido")
	}
	i := 10
	sum := 0
	for index := 0; index < len(cpf)-2; index++ {
		pos, _ := strconv.Atoi(string(cpf[index]))
		sum += pos * i
		i--
	}
	prod := sum * 10
	mod := prod % 11
	if mod == 10 {
		mod = 0
	}
	digit1, _ := strconv.Atoi(string(cpf[9]))
	if mod != digit1 {
		return errors.New("CPF inválido")
	}
	i = 11
	sum = 0
	for index := 0; index < len(cpf)-1; index++ {
		pos, _ := strconv.Atoi(string(cpf[index]))
		sum += pos * i
		i--
	}
	prod = sum * 10
	mod = prod % 11
	if mod == 10 {
		mod = 0
	}
	digit2, _ := strconv.Atoi(string(cpf[10]))
	if mod != digit2 {
		return errors.New("CPF inválido")
	}
	return nil
}

func validaCNPJ(cnpj string) error {
	cnpj = strings.Replace(cnpj, ".", "", -1)
	cnpj = strings.Replace(cnpj, "-", "", -1)
	cnpj = strings.Replace(cnpj, "/", "", -1)
	if len(cnpj) != 14 {
		return errors.New("CNPJ inválido")
	}
	algs := []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}
	var algProdCpfDig1 = make([]int, 12, 12)
	for key, val := range algs {
		intParsed, _ := strconv.Atoi(string(cnpj[key]))
		sumTmp := val * intParsed
		algProdCpfDig1[key] = sumTmp
	}
	sum := 0
	for _, val := range algProdCpfDig1 {
		sum += val
	}
	digit1 := sum % 11
	if digit1 < 2 {
		digit1 = 0
	} else {
		digit1 = 11 - digit1
	}
	char12, _ := strconv.Atoi(string(cnpj[12]))
	if char12 != digit1 {
		return errors.New("CNPJ inválido")
	}
	algs = append([]int{6}, algs...)
	var algProdCpfDig2 = make([]int, 13, 13)
	for key, val := range algs {
		intParsed, _ := strconv.Atoi(string(cnpj[key]))
		sumTmp := val * intParsed
		algProdCpfDig2[key] = sumTmp
	}
	sum = 0
	for _, val := range algProdCpfDig2 {
		sum += val
	}
	digit2 := sum % 11
	if digit2 < 2 {
		digit2 = 0
	} else {
		digit2 = 11 - digit2
	}
	char13, _ := strconv.Atoi(string(cnpj[13]))
	if char13 != digit2 {
		return errors.New("CNPJ inválido")
	}
	return nil
}
