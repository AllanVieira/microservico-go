package db

import (
	"database/sql"
	"fmt"

	//Import da biblioteca do Postgresq
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

//New : Função que inicia uma nova conexão com a base de dado e retorna a sessão conectada
func New() (*sql.DB, error) {
	// Cria conexão com o banco
	dbinfo := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		DB_HOST, DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, err
	}
	errTables := initTables(db)
	if errTables != nil {
		return nil, errTables
	}
	return db, nil
}

func initTables(db *sql.DB) error {
	//Cria tabelas
	_, err := db.Exec(" CREATE TABLE IF NOT EXISTS CLIENTES ( id serial, cpf text, cpf_valido bool, private int, incompleto int, data_ultima_compra date, ticket_medio numeric(18,2), ticket_ultima_compra numeric(18,2), loja_mais_frequente text, loja_mais_frequente_valido bool, loja_ultima_compra text, loja_ultima_compra_valido bool, created_at timestamp)")
	if err != nil {
		return err
	}
	return nil
}
