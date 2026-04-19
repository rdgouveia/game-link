# game-link

API para uma plataforma social de gamers, com busca integrada à [IGDB](https://www.igdb.com/). Permite criar perfis, montar uma biblioteca de jogos e se conectar com outros jogadores através de links.

## Pré-requisitos

- [Go 1.21+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/)
- [Redis](https://redis.io/)
- [goose](https://github.com/pressly/goose) — para rodar as migrations
- [sqlc](https://sqlc.dev/) — apenas se for modificar queries SQL
- Conta de desenvolvedor na [Twitch](https://dev.twitch.tv/console) para obter as credenciais da IGDB
- Bucket no [Cloudflare R2](https://developers.cloudflare.com/r2) para armazenamento de fotos de perfil

---

## Configuração

Copie o arquivo de exemplo e preencha as variáveis:

```bash
cp .env.example .env
```

| Variável | Descrição |
|---|---|
| `DATABASE_URL` | String de conexão do PostgreSQL. Ex: `postgres://user:pass@localhost:5432/gamelink` |
| `JWT_SECRET` | Segredo usado para assinar os tokens JWT. Use uma string longa e aleatória. |
| `IGDB_CLIENT_ID` | Client ID obtido no [Twitch Developer Console](https://dev.twitch.tv/console) |
| `IGDB_CLIENT_SECRET` | Client Secret obtido no Twitch Developer Console |
| `REDIS_URL` | URL de conexão do Redis. Ex: `redis://localhost:6379` |
| `ALLOWED_ORIGINS` | Origins permitidas pelo CORS, separadas por vírgula. Deixe vazio para permitir `*` |
| `R2_ACCOUNT_ID` | ID da conta Cloudflare (visível no dashboard) |
| `R2_ACCESS_KEY_ID` | Access Key ID do token de API do R2 |
| `R2_SECRET_KEY` | Secret Key do token de API do R2 |
| `R2_BUCKET` | Nome do bucket R2 |
| `R2_AVATAR_DIR` | Diretório dentro do bucket para fotos de perfil. Ex: `avatars` |
| `R2_PUBLIC_URL` | URL pública do bucket. Ex: `https://seu-dominio.r2.dev` |

### Obtendo credenciais da IGDB

1. Acesse [dev.twitch.tv/console](https://dev.twitch.tv/console) e faça login com sua conta Twitch
2. Crie uma nova aplicação (categoria: *Application Integration*)
3. Copie o **Client ID** e gere um **Client Secret**

### Configurando o Cloudflare R2

1. Crie um bucket no dashboard do R2
2. Em **Settings → Public Access**, habilite o acesso público e copie o domínio (`R2_PUBLIC_URL`)
3. Em **Manage API Tokens**, crie um token com permissão de leitura/escrita no bucket

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

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| POST | `/register` | Não | Cria um novo usuário |
| POST | `/login` | Não | Autentica e retorna JWT |
| GET | `/check/email?email=` | Não | Verifica se o e-mail já está cadastrado |
| GET | `/check/username?username=` | Não | Verifica se o username já está em uso |

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

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| GET | `/user/me` | Sim | Retorna os dados do usuário logado e seus jogos |
| PATCH | `/user/me` | Sim | Atualiza nome e biografia |
| POST | `/user/me/avatar` | Sim | Faz upload da foto de perfil |
| GET | `/user/search?name=` | Sim | Busca usuários por nome |
| GET | `/user/:username` | Sim | Busca um usuário pelo username |

**PATCH /user/me**
```json
{
  "fullname": "Rafael Gouveia",
  "biography": "Gamer desde os 8 anos."
}
```

**POST /user/me/avatar**
```
Content-Type: multipart/form-data
campo: avatar → arquivo de imagem (jpeg, png ou webp, máx. 5 MB)
```

---

### Jogos

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| GET | `/game/search` | Sim | Busca jogos na IGDB por nome |
| GET | `/game/:igdbId` | Sim | Retorna detalhes de um jogo (local ou IGDB) |
| POST | `/game/add` | Sim | Adiciona um jogo à biblioteca do usuário |
| DELETE | `/game/:id` | Sim | Remove um jogo da biblioteca do usuário |

**GET /game/search?name=zelda**

**POST /game/add**
```json
{
  "igdbId": 1020
}
```

---

### Links

Sistema de conexão entre usuários. Um link precisa ser solicitado e aceito pelo outro lado para ser efetivado.

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| POST | `/link` | Sim | Envia um pedido de link para outro usuário |
| POST | `/link/:userId/accept` | Sim | Aceita um pedido de link recebido |
| DELETE | `/link/:userId` | Sim | Remove um link ou cancela um pedido |
| GET | `/link` | Sim | Lista os links aceitos do usuário logado |
| GET | `/link/pending` | Sim | Lista os pedidos de link recebidos e pendentes |

**POST /link**
```json
{
  "userId": "uuid-do-outro-usuario"
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
