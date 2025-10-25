# Auction System Copilot Instructions

## Architecture Overview

This is a Go-based auction system using **Clean Architecture** with distinct layers:

- **Entities** (`internal/entity/`) - Core domain models with validation
- **Use Cases** (`internal/usecase/`) - Business logic layer with DTOs
- **Infrastructure** (`internal/infra/`) - Controllers, repositories, and external services
- **Configuration** (`configuration/`) - Database, logging, and error handling setup

Key domain entities: `Auction`, `Bid`, `User` - each with their own complete vertical slice.

## Critical Patterns

### Error Handling

Use the **dual error system**:

- `internal_error.InternalError` for internal domain/business errors
- `rest_err.RestErr` for HTTP API responses via `rest_err.ConvertError()`

```go
// Domain layer
return internal_error.NewBadRequestError("invalid auction object")

// Controller layer
restErr := rest_err.ConvertError(err)
c.JSON(restErr.Code, restErr)
```

### Entity Creation & Validation

All entities use **factory functions** with built-in validation:

```go
auction, err := auction_entity.CreateAuction(productName, category, description, condition)
if err != nil {
    return err // Returns *InternalError
}
```

### Repository Pattern

Repositories use **MongoDB-specific structs** for persistence with Unix timestamps:

```go
type AuctionEntityMongo struct {
    Id        string `bson:"_id"`
    Timestamp int64  `bson:"timestamp"` // Unix timestamp
}
```

### Dependency Injection

All dependencies are manually wired in `main.go` `initDependencies()` function:

```go
auctionRepository := auction.NewAuctionRepository(database)
bidRepository := bid.NewBidRepository(database, auctionRepository) // Note: bid depends on auction
```

### Controller Structure

Controllers follow consistent patterns:

1. Bind JSON to DTO
2. Validate using `validation.ValidateErr()`
3. Call use case
4. Convert errors with `rest_err.ConvertError()`
5. Return appropriate HTTP status

## Development Workflows

### Running Locally

```bash
# Start dependencies
docker-compose up mongodb

# Set environment variables in cmd/auction/.env:
# MONGODB_URL=mongodb://localhost:27017
# MONGODB_DB=auction_db

# Run application
go run cmd/auction/main.go
```

### Docker Development

```bash
docker-compose up --build
```

### API Endpoints

- `POST /auction` - Create auction
- `GET /auction` - List auctions (with filters)
- `GET /auction/:auctionId` - Get specific auction
- `GET /auction/winner/:auctionId` - Get winning bid
- `POST /bid` - Place bid
- `GET /bid/:auctionId` - Get bids for auction
- `GET /user/:userId` - Get user details

## Project Conventions

### Package Organization

- Each entity has its own subfolder with complete vertical slice
- Controllers are grouped by domain (`auction_controller/`, `bid_controller/`)
- Database layer mirrors entity structure
- Use cases contain both interfaces and DTOs

### Naming Patterns

- Interfaces end with `Interface` (e.g., `AuctionUseCaseInterface`)
- DTOs are `InputDTO`/`OutputDTO` in use case packages
- MongoDB structs end with `Mongo` (e.g., `AuctionEntityMongo`)
- Repository constructors: `NewXxxRepository(database)`

### Validation Tags

Use Gin binding tags consistently:

```go
type AuctionInputDTO struct {
    ProductName string `json:"product_name" binding:"required,min=1"`
    Category    string `json:"category" binding:"required,min=2"`
    Description string `json:"description" binding:"required,min=10,max=200"`
    Condition   ProductCondition `json:"condition" binding:"oneof=0 1 2"`
}
```

### Critical Dependencies

- **Gin** for HTTP routing
- **MongoDB Driver** for data persistence
- **Zap** for structured logging via `configuration/logger`
- **UUID** for entity IDs (Google UUID library)
- **Validator v10** with custom error translation

When extending the system, maintain the clean architecture boundaries and follow the established error handling patterns.
