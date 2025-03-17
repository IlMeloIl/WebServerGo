# WebServerGo (Chirpy)

## O que este projeto faz

WebServerGo é uma implementação de um servidor web em Go que fornece uma API RESTful para um serviço de microblogging chamado "Chirpy". Ele permite que os usuários se registrem, façam login, publiquem mensagens curtas ("chirps"), e gerenciem seu perfil. O projeto inclui:

- Sistema de autenticação com JWT
- Armazenamento local persistente em PostgreSQL
- Endpoints RESTful para gerenciamento de usuários e conteúdo
- Simulação de webhooks para integração com serviços externos

## Por que você deveria se interessar

Este projeto demonstra:

- Como criar uma API RESTful usando Go e bibliotecas padrão
- Implementação de autenticação segura com JWT e tokens de atualização
- Interação com bancos de dados SQL usando a biblioteca sqlc
- Hashing seguro de senhas e gerenciamento de autenticação

## Como instalar e executar o projeto

### Pré-requisitos

- Go 1.24.0 ou superior
- PostgreSQL
- Git

### Configuração

1. Clone o repositório:
```bash
git clone https://github.com/seu-usuario/WebServerGo.git
cd WebServerGo
```

2. Instale as dependências:
```bash
go mod download
```

3. Configure as variáveis de ambiente criando um arquivo `.env` na raiz do projeto:
```
DB_URL=postgres://usuario:senha@localhost:5432/nome_do_banco
PLATFORM=dev
SECRET_KEY=seu_segredo_para_jwt
POLKA_KEY=chave_para_webhooks
```

4. Configure o banco de dados PostgreSQL:
   - Crie um banco de dados para o projeto
   - Execute os scripts de migração da pasta `sql/schema` na ordem numérica

5. Compile e execute o servidor:
```bash
go build -o chirpy
./chirpy
```

O servidor estará disponível em `http://localhost:8080`

## Documentação da API

### Endpoints de Autenticação

#### Criar Usuário
```
POST /api/users
```
Corpo da requisição:
```json
{
  "email": "usuario@exemplo.com",
  "password": "senha123"
}
```

#### Login
```
POST /api/login
```
Corpo da requisição:
```json
{
  "email": "usuario@exemplo.com",
  "password": "senha123"
}
```
Resposta:
```json
{
  "id": "uuid-do-usuario",
  "email": "usuario@exemplo.com",
  "created_at": "2023-01-01T00:00:00Z",
  "updated_at": "2023-01-01T00:00:00Z",
  "token": "jwt-token",
  "refresh_token": "refresh-token",
  "is_chirpy_red": false
}
```

#### Atualizar Token
```
POST /api/refresh
```
Cabeçalho:
```
Authorization: Bearer refresh-token
```

#### Revogar Token
```
POST /api/revoke
```
Cabeçalho:
```
Authorization: Bearer refresh-token
```

### Endpoints de Usuário

#### Atualizar Usuário
```
PUT /api/users
```
Cabeçalho:
```
Authorization: Bearer jwt-token
```
Corpo da requisição:
```json
{
  "email": "novo@exemplo.com",
  "password": "nova-senha"
}
```

### Endpoints de Chirps

#### Criar Chirp
```
POST /api/chirps
```
Cabeçalho:
```
Authorization: Bearer jwt-token
```
Corpo da requisição:
```json
{
  "body": "Meu primeiro chirp!"
}
```

#### Listar Chirps
```
GET /api/chirps
```
Parâmetros opcionais:
- `author_id` - Filtrar por ID do autor
- `sort` - Ordenar por data (`asc` ou `desc`)

#### Obter Chirp por ID
```
GET /api/chirps/{chirpID}
```

#### Excluir Chirp
```
DELETE /api/chirps/{chirpID}
```
Cabeçalho:
```
Authorization: Bearer jwt-token
```

### Endpoints de Administração

#### Métricas
```
GET /admin/metrics
```

#### Resetar (Somente para ambiente de desenvolvimento)
```
POST /admin/reset
```

### Webhooks

#### Webhook Polka (para upgrade de usuários)
```
POST /api/polka/webhooks
```
Cabeçalho:
```
Authorization: Bearer polka-api-key
```
Corpo da requisição:
```json
{
  "event": "user.upgraded",
  "data": {
    "user_id": "uuid-do-usuario"
  }
}
```

### Verificação de Saúde

```
GET /api/healthz
```

## Notas de Implementação

- O sistema limita chirps a 140 caracteres
- O acesso ao Chirpy Red é gerenciado através de webhooks simulados
- Os tokens JWT expiram após 1 hora
- Os tokens de atualização são válidos por 60 dias
- Senhas são armazenadas com hash seguro usando bcrypt