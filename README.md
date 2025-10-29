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

## 🧪 Validação do Fechamento Automático

### ✅ Teste Principal de Validação

**Arquivo**: `internal/infra/database/auction/automated_closing_test.go`

O sistema inclui um **teste específico** que valida o fechamento automatizado:

```go
func TestAutomaticAuctionClosingBehavior(t *testing.T)
```

**Este teste:**

- Cria múltiplos leilões com intervalo de 3 segundos
- Verifica que todos estão ativos inicialmente
- Aguarda o tempo de expiração + buffer
- **VALIDA que todos foram fechados automaticamente**
- Fornece logs detalhados do processo

### Execução do Teste de Validação

```bash
# Iniciar MongoDB para testes
docker-compose up mongodb -d

# Executar o teste específico de fechamento automático
go test ./internal/infra/database/auction -v -run TestAutomaticAuctionClosingBehavior

# Executar todos os testes de fechamento automático
go test ./internal/infra/database/auction -v -run "TestAutomatic|TestAuction.*Close"
```

### Saída Esperada do Teste

```
=== RUN   TestAutomaticAuctionClosingBehavior
    automated_closing_test.go:35: 🚀 Iniciando teste de fechamento automático...
    automated_closing_test.go:49: ✅ Leilão 1 criado: Produto Teste 1 (ID: abc-123)
    automated_closing_test.go:49: ✅ Leilão 2 criado: Produto Teste 2 (ID: def-456)
    automated_closing_test.go:49: ✅ Leilão 3 criado: Produto Teste 3 (ID: ghi-789)
    automated_closing_test.go:58: 📋 Verificando estado inicial dos leilões...
    automated_closing_test.go:64: ✅ Todos os leilões estão ativos inicialmente
    automated_closing_test.go:70: ⏰ Aguardando 6s para fechamento automático...
    automated_closing_test.go:80: 🔍 Verificando fechamento automático...
    automated_closing_test.go:86: ✅ Leilão 1: FECHADO AUTOMATICAMENTE
    automated_closing_test.go:86: ✅ Leilão 2: FECHADO AUTOMATICAMENTE
    automated_closing_test.go:86: ✅ Leilão 3: FECHADO AUTOMATICAMENTE
    automated_closing_test.go:98: 🎉 SUCESSO: Todos os 3 leilões foram fechados automaticamente!
--- PASS: TestAutomaticAuctionClosingBehavior (6.05s)
```

### 1. Teste Manual via API

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

## 🧪 Arquivos de Teste Criados

### Teste Principal (Solicitado na Avaliação)

- **`internal/infra/database/auction/automated_closing_test.go`**
  - `TestAutomaticAuctionClosingBehavior()` - **TESTE PRINCIPAL**
  - `TestAuctionStaysActiveBeforeInterval()` - Validação de controle

### Testes Complementares

- `internal/infra/database/auction/auction_auto_close_test.go` - Testes básicos
- `internal/infra/database/auction/auction_unit_test.go` - Testes unitários

### Como Executar o Teste Principal

```bash
# Apenas o teste solicitado na correção
go test ./internal/infra/database/auction -v -run TestAutomaticAuctionClosingBehavior

# Com logs detalhados
go test ./internal/infra/database/auction -v -run TestAutomaticAuctionClosingBehavior -test.v
```

## Contribuições

Este projeto foi desenvolvido como atividade avaliativa do curso Go Expert da Full Cycle.

**Autor**: Thyrso Mancini Neto  
**Repositório**: https://github.com/thyrso/PosGoExpert-Labs-02

### ✅ Atualização da Avaliação

**Ponto Solicitado**: "Adicione um teste para validar se o fechamento está acontecendo de forma automatizada"

**✅ Implementado**: Teste `TestAutomaticAuctionClosingBehavior` que demonstra e valida o fechamento automático de múltiplos leilões com logs detalhados do processo.
