package service

import (
	"fmt"
	"log"
	"parkit-data-ETL/internal/models"
	"time"
)

type NYCClient interface {
	FetchParkingMeters(offset int) ([]models.ParkingMeter, error)
	GetTotalCount() (int, error)
}

type Database interface {
	UpsertParkingMeters(meters []models.ParkingMeter) error
}

type SyncService struct {
	client NYCClient
	db     Database
}

func NewSyncService(client NYCClient, db Database) *SyncService {
	return &SyncService{
		client: client,
		db:     db,
	}
}

func (s *SyncService) Run() error {
	// Get total count first
	totalCount, err := s.client.GetTotalCount()
	if err != nil {
		return fmt.Errorf("failed to get total count: %w", err)
	}

	offset := 0
	totalProcessed := 0
	startTime := time.Now()
	batchNum := 1

	log.Printf("Starting sync of %d parking meters...", totalCount)

	for {
		log.Printf("Fetching batch %d (offset: %d, progress: %.1f%%)...",
			batchNum, offset, float64(totalProcessed)/float64(totalCount)*100)

		meters, err := s.client.FetchParkingMeters(offset)
		if err != nil {
			return fmt.Errorf("failed to fetch batch %d: %w", batchNum, err)
		}

		if len(meters) == 0 {
			break
		}

		if err := s.db.UpsertParkingMeters(meters); err != nil {
			return fmt.Errorf("failed to upsert batch %d: %w", batchNum, err)
		}

		totalProcessed += len(meters)
		offset += len(meters)
		batchNum++

		log.Printf("Processed batch %d: %d meters (total: %d, %.1f%% complete)",
			batchNum-1, len(meters), totalProcessed,
			float64(totalProcessed)/float64(totalCount)*100)

		// Break if we've processed all records
		if totalProcessed >= totalCount {
			break
		}

		// Rate limiting to be nice to the API
		time.Sleep(100 * time.Millisecond)
	}

	duration := time.Since(startTime)
	log.Printf("Sync completed. Processed %d meters in %v", totalProcessed, duration)
	return nil
}
