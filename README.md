# Auction System with Auto-Close Feature

## Overview

Este é um sistema de leilões desenvolvido em Go que implementa fechamento automático de leilões baseado em tempo configurável via variáveis de ambiente.

## Nova Funcionalidade - Fechamento Automático

### Implementações Realizadas

1. **Goroutine para Fechamento Automático**: Uma goroutine que executa em background verificando leilões expirados a cada 10 segundos
2. **Cálculo de Tempo de Leilão**: Baseado na variável de ambiente `AUCTION_INTERVAL`
3. **Concorrência Segura**: Implementação com mutexes e context para controle de concorrência
4. **Testes Automatizados**: Validação do comportamento de fechamento automático

### Arquivos Modificados

- `internal/entity/auction_entity/auction_entity.go` - Adicionados novos métodos na interface
- `internal/infra/database/auction/create_auction.go` - Implementação principal da funcionalidade
- `internal/infra/database/auction/find_auction.go` - Métodos de busca e atualização
- `internal/infra/database/auction/auction_auto_close_test.go` - Testes automatizados

## Configuração

### Variáveis de Ambiente

Configure as seguintes variáveis no arquivo `cmd/auction/.env`:

```env
# Intervalo para fechamento automático do leilão
AUCTION_INTERVAL=20s

# Configurações do MongoDB
MONGODB_URL=mongodb://admin:admin@mongodb:27017/auctions?authSource=admin
MONGODB_DB=auctions
MONGO_INITDB_ROOT_USERNAME=admin
MONGO_INITDB_ROOT_PASSWORD=admin

# Configurações de lote para bids
BATCH_INSERT_INTERVAL=20s
MAX_BATCH_SIZE=4
```

## Como Executar

### Pré-requisitos

- Docker e Docker Compose
- Go 1.20+ (para desenvolvimento local)

### Execução com Docker Compose (Recomendado)

1. Clone o repositório:

```bash
git clone https://github.com/thyrso/PosGoExpert-Labs-02.git
cd PosGoExpert-Labs-02
```

2. Execute o sistema completo:

```bash
docker-compose up --build
```

O sistema estará disponível em `http://localhost:8080`

### Execução em Desenvolvimento Local

1. Inicie apenas o MongoDB:

```bash
docker-compose up mongodb
```

2. Execute a aplicação Go:

```bash
go run cmd/auction/main.go
```

## API Endpoints

### Leilões

- `POST /auction` - Criar leilão
- `GET /auction` - Listar leilões
- `GET /auction/:auctionId` - Buscar leilão por ID
- `GET /auction/winner/:auctionId` - Buscar lance vencedor

### Lances

- `POST /bid` - Criar lance
- `GET /bid/:auctionId` - Buscar lances do leilão

### Usuários

- `GET /user/:userId` - Buscar usuário por ID

## Testando a Funcionalidade de Fechamento Automático

### 1. Teste Manual

```bash
# 1. Configure um intervalo curto (ex: 30s) no .env
AUCTION_INTERVAL=30s

# 2. Crie um leilão
curl -X POST http://localhost:8080/auction \
  -H "Content-Type: application/json" \
  -d '{
    "product_name": "Smartphone",
    "category": "Electronics",
    "description": "iPhone 14 Pro Max in excellent condition",
    "condition": 1
  }'

# 3. Aguarde o tempo configurado + buffer e verifique se foi fechado
curl http://localhost:8080/auction
```

### 2. Testes Automatizados

Execute os testes para validar o fechamento automático:

```bash
# Executar testes específicos de fechamento automático
go test ./internal/infra/database/auction -v -run TestAuctionAutoClose

# Executar todos os testes
go test ./... -v
```

### Exemplo de Teste

O teste automatizado:

1. Configura `AUCTION_INTERVAL=2s`
2. Cria um leilão
3. Verifica que está ativo inicialmente
4. Aguarda 4 segundos
5. Verifica que foi automaticamente fechado

## Funcionalidades Técnicas Implementadas

### Goroutine de Fechamento Automático

- **Localização**: `internal/infra/database/auction/create_auction.go`
- **Funcionamento**:
  - Executa a cada 10 segundos
  - Busca leilões ativos criados antes do tempo limite
  - Atualiza o status para "Completed"
  - Registra logs das operações

### Gerenciamento de Concorrência

- **Context Control**: Para parada graceful da goroutine
- **Singleton Pattern**: Uma única goroutine por repositório
- **Thread Safety**: Operações seguras no MongoDB

### Cálculo de Tempo

```go
func getAuctionInterval() time.Duration {
    auctionInterval := os.Getenv("AUCTION_INTERVAL")
    duration, err := time.ParseDuration(auctionInterval)
    if err != nil {
        return time.Minute * 5 // default 5 minutos
    }
    return duration
}
```

## Logs

O sistema registra automaticamente:

- Início e parada da goroutine de fechamento
- Leilões fechados automaticamente com ID e produto
- Erros durante o processo de fechamento

Exemplo de log:

```
INFO: Auction closer goroutine started
INFO: Auction automatically closed: abc123 - Smartphone iPhone
```

## Estrutura do Projeto

```
internal/
├── entity/auction_entity/          # Entidades e interfaces de domínio
├── infra/database/auction/         # Implementação de persistência
├── usecase/auction_usecase/        # Casos de uso e DTOs
└── infra/api/web/controller/       # Controllers HTTP
```

## Troubleshooting

### Problemas Comuns

1. **MongoDB não conecta**: Verifique se o container está rodando
2. **Leilões não fecham**: Verifique a variável `AUCTION_INTERVAL`
3. **Testes falham**: Certifique-se que o MongoDB está disponível

### Debug

Para debug adicional, ajuste o nível de log ou adicione logs personalizados no método `closeExpiredAuctions()`.

## Contribuições

Este projeto foi desenvolvido como atividade avaliativa do curso Go Expert da Full Cycle.

**Autor**: Thyrso Mancini Neto
**Repositório**: https://github.com/thyrso/PosGoExpert-Labs-02
