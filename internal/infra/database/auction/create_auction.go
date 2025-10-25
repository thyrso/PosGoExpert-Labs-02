package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection      *mongo.Collection
	auctionInterval time.Duration
	closeOnce       sync.Once
	ctx             context.Context
	cancel          context.CancelFunc
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	ctx, cancel := context.WithCancel(context.Background())

	repo := &AuctionRepository{
		Collection:      database.Collection("auctions"),
		auctionInterval: getAuctionInterval(),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Start the auction closer goroutine only once
	repo.closeOnce.Do(func() {
		go repo.auctionCloser()
	})

	return repo
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	return nil
}

// getAuctionInterval gets auction interval from environment variables
func getAuctionInterval() time.Duration {
	auctionInterval := os.Getenv("AUCTION_INTERVAL")
	duration, err := time.ParseDuration(auctionInterval)
	if err != nil {
		return time.Minute * 5 // default 5 minutes
	}
	return duration
}

// auctionCloser runs in background to close expired auctions
func (ar *AuctionRepository) auctionCloser() {
	ticker := time.NewTicker(time.Second * 10) // Check every 10 seconds
	defer ticker.Stop()

	logger.Info("Auction closer goroutine started")

	for {
		select {
		case <-ar.ctx.Done():
			logger.Info("Auction closer goroutine stopped")
			return
		case <-ticker.C:
			ar.closeExpiredAuctions()
		}
	}
}

// closeExpiredAuctions finds and closes auctions that have exceeded their time limit
func (ar *AuctionRepository) closeExpiredAuctions() {
	ctx := context.Background()

	// Calculate the cutoff timestamp (current time - auction interval)
	cutoffTime := time.Now().Add(-ar.auctionInterval)
	cutoffTimestamp := cutoffTime.Unix()

	// Find active auctions older than the cutoff time
	expiredAuctions, err := ar.FindActiveAuctionsOlderThan(ctx, cutoffTimestamp)
	if err != nil {
		logger.Error("Error finding expired auctions", err)
		return
	}

	// Close each expired auction
	for _, auction := range expiredAuctions {
		if updateErr := ar.UpdateAuctionStatus(ctx, auction.Id, auction_entity.Completed); updateErr != nil {
			logger.Error("Error closing expired auction", updateErr)
		} else {
			logger.Info("Auction automatically closed: " + auction.Id + " - " + auction.ProductName)
		}
	}
}

// Close stops the auction closer goroutine
func (ar *AuctionRepository) Close() {
	ar.cancel()
}
