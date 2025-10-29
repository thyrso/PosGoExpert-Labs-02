package auction_test

import (
	"context"
	"fmt"
	"fullcycle-auction_go/configuration/database/mongodb"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/infra/database/auction"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func setupTestDatabaseForAutoClose() (*mongo.Database, error) {
	// Set environment variables directly for testing
	os.Setenv("MONGODB_URL", "mongodb://admin:admin@localhost:27017/auctions?authSource=admin")
	os.Setenv("MONGODB_DB", "auctions_test_autoclose")

	ctx := context.Background()
	return mongodb.NewMongoDBConnection(ctx)
}

// TestAutomaticAuctionClosingBehavior - Teste específico solicitado para validar o fechamento automatizado
func TestAutomaticAuctionClosingBehavior(t *testing.T) {
	// Configure uma duração muito curta para teste rápido
	os.Setenv("AUCTION_INTERVAL", "3s")

	database, err := setupTestDatabaseForAutoClose()
	assert.Nil(t, err)
	assert.NotNil(t, database)

	auctionRepository := auction.NewAuctionRepository(database)
	defer auctionRepository.Close()

	// Criar múltiplos leilões para testar o comportamento em lote
	auctions := make([]*auction_entity.Auction, 0, 3)

	t.Log("🚀 Iniciando teste de fechamento automático de leilões...")

	for i := 0; i < 3; i++ {
		testAuction, createErr := auction_entity.CreateAuction(
			fmt.Sprintf("Produto Teste %d", i+1),
			"Electronics",
			"Descrição de teste para validação de fechamento automático",
			auction_entity.New)
		assert.Nil(t, createErr)
		assert.NotNil(t, testAuction)

		// Inserir no banco
		insertErr := auctionRepository.CreateAuction(context.Background(), testAuction)
		assert.Nil(t, insertErr)

		auctions = append(auctions, testAuction)
		t.Logf("✅ Leilão %d criado: %s (ID: %s)", i+1, testAuction.ProductName, testAuction.Id)
	}

	// Verificar que todos estão ativos inicialmente
	t.Log("\n📋 Verificando estado inicial dos leilões...")
	for i, auction := range auctions {
		foundAuction, findErr := auctionRepository.FindAuctionById(context.Background(), auction.Id)
		assert.Nil(t, findErr)
		assert.Equal(t, auction_entity.Active, foundAuction.Status,
			"Leilão %d deveria estar ativo inicialmente", i+1)
		t.Logf("   Leilão %d (%s): ATIVO ✅", i+1, auction.ProductName)
	}

	// Aguardar o tempo de fechamento automático + buffer
	waitTime := 6 * time.Second // 3s interval + 3s buffer para garantir execução
	t.Logf("\n⏰ Aguardando %v para o fechamento automático acontecer...", waitTime)

	// Mostrar progresso do tempo
	for i := 0; i < int(waitTime.Seconds()); i++ {
		time.Sleep(1 * time.Second)
		t.Logf("   Aguardando... %d/%d segundos", i+1, int(waitTime.Seconds()))
	}

	// Verificar que todos foram fechados automaticamente
	t.Log("\n🔍 Verificando se os leilões foram fechados automaticamente...")
	allClosed := true
	closedCount := 0

	for i, auction := range auctions {
		updatedAuction, findErr := auctionRepository.FindAuctionById(context.Background(), auction.Id)
		assert.Nil(t, findErr)

		if updatedAuction.Status == auction_entity.Completed {
			t.Logf("   ✅ Leilão %d (%s): FECHADO AUTOMATICAMENTE", i+1, auction.ProductName)
			closedCount++
		} else {
			t.Errorf("   ❌ Leilão %d (%s): NÃO foi fechado (status: %d)", i+1, auction.ProductName, updatedAuction.Status)
			allClosed = false
		}

		// Esta é a asserção principal do teste
		assert.Equal(t, auction_entity.Completed, updatedAuction.Status,
			"Leilão %d (%s) deveria ter sido fechado automaticamente após %v",
			i+1, auction.ProductName, waitTime)
	}

	// Resumo final
	if allClosed {
		t.Logf("\n🎉 SUCESSO: Todos os %d leilões foram fechados automaticamente!", closedCount)
		t.Log("✅ O sistema de fechamento automático está funcionando corretamente!")
	} else {
		t.Errorf("\n❌ FALHA: Apenas %d de %d leilões foram fechados automaticamente", closedCount, len(auctions))
	}

	// Limpeza
	t.Log("\n🧹 Limpando dados de teste...")
	for i, auction := range auctions {
		_, deleteErr := database.Collection("auctions").DeleteOne(
			context.Background(),
			map[string]interface{}{"_id": auction.Id},
		)
		assert.Nil(t, deleteErr)
		t.Logf("   Leilão %d removido ✅", i+1)
	}

	t.Log("\n✅ Teste de fechamento automático concluído!")
}

// TestAuctionStaysActiveBeforeInterval - Validar que leilões NÃO são fechados antes do tempo
func TestAuctionStaysActiveBeforeInterval(t *testing.T) {
	// Configure um intervalo longo
	os.Setenv("AUCTION_INTERVAL", "30s")

	database, err := setupTestDatabaseForAutoClose()
	assert.Nil(t, err)
	assert.NotNil(t, database)

	auctionRepository := auction.NewAuctionRepository(database)
	defer auctionRepository.Close()

	t.Log("🚀 Testando que leilões permanecem ativos antes do intervalo...")

	// Criar um leilão de teste
	testAuction, createErr := auction_entity.CreateAuction(
		"Produto Controle",
		"Electronics",
		"Teste para validar que não fecha antes do tempo",
		auction_entity.New)
	assert.Nil(t, createErr)
	assert.NotNil(t, testAuction)

	// Inserir no banco
	insertErr := auctionRepository.CreateAuction(context.Background(), testAuction)
	assert.Nil(t, insertErr)
	t.Logf("✅ Leilão controle criado: %s", testAuction.ProductName)

	// Verificar que está ativo inicialmente
	foundAuction, findErr := auctionRepository.FindAuctionById(context.Background(), testAuction.Id)
	assert.Nil(t, findErr)
	assert.Equal(t, auction_entity.Active, foundAuction.Status)
	t.Log("✅ Leilão está ativo inicialmente")

	// Aguardar um tempo menor que o intervalo configurado
	waitTime := 5 * time.Second
	t.Logf("⏰ Aguardando %v (menos que os 30s configurados)...", waitTime)
	time.Sleep(waitTime)

	// Verificar que ainda está ativo
	updatedAuction, findErr2 := auctionRepository.FindAuctionById(context.Background(), testAuction.Id)
	assert.Nil(t, findErr2)
	assert.Equal(t, auction_entity.Active, updatedAuction.Status,
		"Leilão deveria ainda estar ativo após %v (intervalo configurado: 30s)", waitTime)

	t.Log("✅ SUCESSO: Leilão permanece ativo antes do tempo limite!")

	// Limpeza
	_, deleteErr := database.Collection("auctions").DeleteOne(
		context.Background(),
		map[string]interface{}{"_id": testAuction.Id},
	)
	assert.Nil(t, deleteErr)
	t.Log("🧹 Leilão de controle removido")
}
