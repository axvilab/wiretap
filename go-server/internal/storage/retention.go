package storage

import (
	"log"
	"time"
)

func (s *Store) StartRetention(days int) {
	if days <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		// Run once on startup
		s.cleanup(days)

		for range ticker.C {
			s.cleanup(days)
		}
	}()
	log.Printf("Retention policy: %d days, cleanup every hour", days)
}

func (s *Store) cleanup(days int) {
	cutoff := float64(time.Now().AddDate(0, 0, -days).Unix())

	res, err := s.db.Exec("DELETE FROM traces WHERE ts_start < ?", cutoff)
	if err != nil {
		log.Printf("[retention] error cleaning traces: %v", err)
		return
	}
	tracesDeleted, _ := res.RowsAffected()

	res, err = s.db.Exec("DELETE FROM emails WHERE created_at < ?", cutoff)
	if err != nil {
		log.Printf("[retention] error cleaning emails: %v", err)
		return
	}
	emailsDeleted, _ := res.RowsAffected()

	res, err = s.db.Exec("DELETE FROM job_links WHERE dispatched_at < ?", cutoff)
	if err != nil {
		log.Printf("[retention] error cleaning job_links: %v", err)
		return
	}
	jobsDeleted, _ := res.RowsAffected()

	if tracesDeleted > 0 || emailsDeleted > 0 || jobsDeleted > 0 {
		log.Printf("[retention] cleaned: %d traces, %d emails, %d job_links", tracesDeleted, emailsDeleted, jobsDeleted)
	}
}
