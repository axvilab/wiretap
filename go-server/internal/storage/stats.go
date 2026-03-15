package storage

import "os"

type StorageStatsResult struct {
	DBSizeBytes int64   `json:"db_size_bytes"`
	DBSizeMB    float64 `json:"db_size_mb"`
	TotalTraces int     `json:"total_traces"`
	TotalEmails int     `json:"total_emails"`
	TotalJobs   int     `json:"total_jobs"`
}

func (s *Store) StorageStats() (*StorageStatsResult, error) {
	stats := &StorageStatsResult{}

	// DB file size (main + WAL)
	var totalSize int64
	if info, err := os.Stat(s.path); err == nil {
		totalSize += info.Size()
	}
	if info, err := os.Stat(s.path + ".wal"); err == nil {
		totalSize += info.Size()
	}
	stats.DBSizeBytes = totalSize
	stats.DBSizeMB = float64(totalSize) / 1024 / 1024

	s.db.QueryRow("SELECT COUNT(*) FROM traces").Scan(&stats.TotalTraces)
	s.db.QueryRow("SELECT COUNT(*) FROM emails").Scan(&stats.TotalEmails)
	s.db.QueryRow("SELECT COUNT(*) FROM job_links").Scan(&stats.TotalJobs)

	return stats, nil
}

func (s *Store) ClearData(olderThanDays int) error {
	if olderThanDays == 0 {
		// Clear everything
		s.db.Exec("DELETE FROM traces")
		s.db.Exec("DELETE FROM emails")
		s.db.Exec("DELETE FROM job_links")
		return nil
	}

	// Clear older than N days
	s.cleanup(olderThanDays)
	return nil
}
