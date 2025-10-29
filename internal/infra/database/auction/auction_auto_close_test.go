package auction_test

import (
	"context"
	"fullcycle-auction_go/configuration/database/mongodb"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/infra/database/auction"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo"
)

func setupTestDatabase() (*mongo.Database, error) {
	// Set environment variables directly for testing
	os.Setenv("MONGODB_URL", "mongodb://admin:admin@localhost:27017/auctions?authSource=admin")
	os.Setenv("MONGODB_DB", "auctions_test")

	ctx := context.Background()
	return mongodb.NewMongoDBConnection(ctx)
}

func TestAuctionAutoClose(t *testing.T) {
	// Set a very short auction interval for testing
	os.Setenv("AUCTION_INTERVAL", "2s")

	database, err := setupTestDatabase()
	assert.Nil(t, err)
	assert.NotNil(t, database)

	auctionRepository := auction.NewAuctionRepository(database)

	// Create a test auction
	testAuction, err := auction_entity.CreateAuction(
		"Test Product",
		"Electronics",
		"Test Description for automated closing",
		auction_entity.New)
	assert.Nil(t, err)
	assert.NotNil(t, testAuction)

	// Create the auction in the database
	createErr := auctionRepository.CreateAuction(context.Background(), testAuction)
	assert.Nil(t, createErr)

	// Verify auction is initially active
	foundAuction, findErr := auctionRepository.FindAuctionById(context.Background(), testAuction.Id)
	assert.Nil(t, findErr)
	assert.Equal(t, auction_entity.Active, foundAuction.Status)

	// Wait for auto-close (auction interval + buffer time)
	time.Sleep(4 * time.Second)

	// Verify auction has been automatically closed
	updatedAuction, findErr2 := auctionRepository.FindAuctionById(context.Background(), testAuction.Id)
	assert.Nil(t, findErr2)
	assert.Equal(t, auction_entity.Completed, updatedAuction.Status)

	// Cleanup
	auctionRepository.Close()

	// Clean up test data
	_, deleteErr := database.Collection("auctions").DeleteOne(
		context.Background(),
		map[string]interface{}{"_id": testAuction.Id},
	)
	assert.NoError(t, deleteErr)
}

func TestAuctionNotClosedBeforeInterval(t *testing.T) {
	// Set a longer auction interval for testing
	os.Setenv("AUCTION_INTERVAL", "10s")

	database, err := setupTestDatabase()
	assert.Nil(t, err)
	assert.NotNil(t, database)

	auctionRepository := auction.NewAuctionRepository(database)

	// Create a test auction
	testAuction, err := auction_entity.CreateAuction(
		"Test Product 2",
		"Electronics",
		"Test Description for non-automated closing",
		auction_entity.New)
	assert.Nil(t, err)
	assert.NotNil(t, testAuction)

	// Create the auction in the database
	createErr := auctionRepository.CreateAuction(context.Background(), testAuction)
	assert.Nil(t, createErr)

	// Wait a short time (less than interval)
	time.Sleep(2 * time.Second)

	// Verify auction is still active
	foundAuction, findErr := auctionRepository.FindAuctionById(context.Background(), testAuction.Id)
	assert.Nil(t, findErr)
	assert.Equal(t, auction_entity.Active, foundAuction.Status)

	// Cleanup
	auctionRepository.Close()

	// Clean up test data
	_, deleteErr := database.Collection("auctions").DeleteOne(
		context.Background(),
		map[string]interface{}{"_id": testAuction.Id},
	)
	assert.NoError(t, deleteErr)
}
