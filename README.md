# Neoway Teste

Microseriço de importação de arquivo texto para base de dados postgresql, ecrito em GO e Docker Compose



```
git clone https://github.com/AllanVieira/microservico-go
```

### Instalando

Rodando com Docker composer

```
docker-compose up
```

Preciso verificar porque não cria base de dados antes de inicar a API em GO
Se der conexão com a base de dados recusada, roda de novo o comando :
```
docker-compose up
```

A aplicação vai criar automaticamente a tabela de dados:


```
CREATE TABLE IF NOT EXISTS CLIENTES ( 
  id serial, 
  cpf text, 
  cpf_valido bool, 
  private int, 
  incompleto int,
  data_ultima_compra date, 
  ticket_medio numeric(18,2), 
  ticket_ultima_compra numeric(18,2), 
  loja_mais_frequente text, 
  loja_mais_frequente_valido bool, 
  loja_ultima_compra text, 
  loja_ultima_compra_valido bool, 
  created_at timestamp
 )
```

### Testando

Acessando http://localhost vai encontrar um formulario para envio do arquivo texto
Depois de enviado, os dados vão estar sendo importados para dentro da base de dados de nome Files
No console da Aplicação exibir os resumos da importação, após a mensagem de fim pode conferir os registros na base de dados.

Na Base de dados :
  HOST : localhost:5432
  USER: postgres
  PASSWORD: postgres
  DB: Files
  TABLE: Clientes
  
Selecionar clientes com cpf inválidos :
```
select * from clientes where not cpf_valido
```

### TODO:
    - Refatorar código, implementar novos pacotes para amazenar as funções do main
