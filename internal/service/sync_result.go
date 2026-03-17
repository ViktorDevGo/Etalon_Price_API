package service

import "time"

// SyncResult contains synchronization statistics
type SyncResult struct {
	Type            string        // "tyres" or "rims"
	NewCount        int           // Number of new records added
	SkippedCount    int           // Number of duplicates skipped
	TotalInDB       int           // Total records in database after sync
	FilteredCount   int           // Number of records filtered out by type
	FilteredOutType string        // Description of filtered types
	Duration        time.Duration // Total sync duration
	StartedAt       time.Time     // Sync start time
	CompletedAt     time.Time     // Sync completion time
	Error           error         // Error if sync failed
}
