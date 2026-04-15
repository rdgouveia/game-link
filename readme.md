# game-link

API para gerenciar uma biblioteca pessoal de jogos, com busca integrada à [IGDB](https://www.igdb.com/).

## Pré-requisitos

- [Go 1.21+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)
- [goose](https://github.com/pressly/goose) — para rodar as migrations
- [sqlc](https://sqlc.dev/) — apenas se for modificar queries SQL
- Conta de desenvolvedor na [Twitch](https://dev.twitch.tv/console) para obter as credenciais da IGDB

---

## Configuração

Copie o arquivo de exemplo e preencha as variáveis:

```bash
cp .env.example .env
```

| Variável           | Descrição                                                  |
|--------------------|------------------------------------------------------------|
| `DATABASE_URL`     | String de conexão do PostgreSQL. Ex: `postgres://user:pass@localhost:5432/gamelink` |
| `JWT_SECRET`       | Segredo usado para assinar os tokens JWT. Use uma string longa e aleatória. |
| `IGDB_CLIENT_ID`   | Client ID obtido no [Twitch Developer Console](https://dev.twitch.tv/console) |
| `IGDB_CLIENT_SECRET` | Client Secret obtido no Twitch Developer Console         |
| `REDIS_URL`        | URL de conexão do Redis. Ex: `redis://localhost:6379`      |

### Obtendo credenciais da IGDB

1. Acesse [dev.twitch.tv/console](https://dev.twitch.tv/console) e faça login com sua conta Twitch
2. Crie uma nova aplicação (categoria: *Application Integration*)
3. Copie o **Client ID** e gere um **Client Secret**

---

## Instalando dependências

```bash
go mod download
```

---

## Banco de dados

Crie o banco no PostgreSQL:

```sql
CREATE DATABASE gamelink;
```

Rode as migrations:

```bash
make migrate-up
```

Para desfazer a última migration:

```bash
make migrate-down
```

Para verificar o status das migrations:

```bash
make migrate-status
```

---

## Rodando o projeto

```bash
go run .
```

O servidor sobe na porta `8080`.

---

## Endpoints

Todos os endpoints protegidos exigem o header:
```
Authorization: Bearer <token>
```

### Auth

| Método | Rota        | Auth | Descrição              |
|--------|-------------|------|------------------------|
| POST   | `/register` | Não  | Cria um novo usuário   |
| POST   | `/login`    | Não  | Autentica e retorna JWT |

**POST /register**
```json
{
  "username": "rafael",
  "email": "rafael@email.com",
  "password": "senha123",
  "fullname": "Rafael Gouveia"
}
```

**POST /login**
```json
{
  "email": "rafael@email.com",
  "password": "senha123"
}
```

---

### Usuário

| Método | Rota       | Auth | Descrição                        |
|--------|------------|------|----------------------------------|
| GET    | `/user/me` | Sim  | Retorna os dados do usuário logado |

---

### Jogos

| Método | Rota               | Auth | Descrição                                         |
|--------|--------------------|------|---------------------------------------------------|
| GET    | `/game/search`     | Sim  | Busca jogos na IGDB por nome                      |
| GET    | `/game/:igdbId`    | Sim  | Retorna detalhes de um jogo (local ou IGDB)       |
| POST   | `/game/add`        | Sim  | Adiciona um jogo à biblioteca do usuário          |

**GET /game/search?name=zelda**

**POST /game/add**
```json
{
  "igdbId": 1020
}
```

---

## Comandos úteis

```bash
# Rodar migrations
make migrate-up

# Desfazer última migration
make migrate-down

# Criar nova migration
make migrate-create NAME=nome_da_migration

# Regenerar código a partir das queries SQL
make generate
```
