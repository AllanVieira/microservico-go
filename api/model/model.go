package model

import "time"

type Cliente struct {
	Cpf                     string    `json:"cpf"`
	CpfValido               bool      `json:"cpf_valido"`
	Private                 int       `json:"private"`
	Incompleto              int       `json:"incompleto"`
	DataUltimaCompra        time.Time `json:"data_ultima_compra"`
	TicketMedio             float64   `json:"ticket_medio"`
	TicketUltimaCompra      float64   `json:"ticket_ultima_compra"`
	LojaMaisFrequente       string    `json:"loja_mais_frequente"`
	LojaMaisFrequenteValido bool      `json:"loja_mais_frequente_valido"`
	LojaUltimaCompra        string    `json:"loja_ultima_compra"`
	LojaUltimaCompraValido  bool      `json:"loja_ultima_compra_valido"`
}
