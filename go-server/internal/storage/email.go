package storage

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (s *Store) SaveEmail(e *Email) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	if e.CreatedAt == 0 {
		e.CreatedAt = float64(time.Now().UnixMilli()) / 1000
	}

	toJSON, _ := json.Marshal(e.To)
	ccJSON, _ := json.Marshal(e.Cc)
	attachJSON, _ := json.Marshal(e.Attachments)

	_, err := s.db.Exec(`
		INSERT INTO emails (id, from_addr, to_addrs, cc_addrs, subject, html, text_body, attachments_json, trace_rid, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		e.ID, e.From, string(toJSON), string(ccJSON), e.Subject,
		e.HTML, e.Text, string(attachJSON), e.TraceRID, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert email: %w", err)
	}
	return nil
}

func (s *Store) ListEmails(page, perPage int) ([]Email, int, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}

	var total int
	s.db.QueryRow("SELECT COUNT(*) FROM emails").Scan(&total)

	offset := (page - 1) * perPage
	rows, err := s.db.Query(
		`SELECT id, from_addr, to_addrs, cc_addrs, subject, html, text_body, attachments_json, trace_rid, created_at
		FROM emails ORDER BY created_at DESC LIMIT ? OFFSET ?`, perPage, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var emails []Email
	for rows.Next() {
		var e Email
		var toJSON, ccJSON, attachJSON string
		err := rows.Scan(&e.ID, &e.From, &toJSON, &ccJSON, &e.Subject, &e.HTML, &e.Text, &attachJSON, &e.TraceRID, &e.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		json.Unmarshal([]byte(toJSON), &e.To)
		json.Unmarshal([]byte(ccJSON), &e.Cc)
		json.Unmarshal([]byte(attachJSON), &e.Attachments)
		emails = append(emails, e)
	}
	if emails == nil {
		emails = []Email{}
	}
	return emails, total, nil
}

func (s *Store) GetEmail(id string) (*Email, error) {
	var e Email
	var toJSON, ccJSON, attachJSON string
	err := s.db.QueryRow(
		`SELECT id, from_addr, to_addrs, cc_addrs, subject, html, text_body, attachments_json, trace_rid, created_at
		FROM emails WHERE id = ?`, id,
	).Scan(&e.ID, &e.From, &toJSON, &ccJSON, &e.Subject, &e.HTML, &e.Text, &attachJSON, &e.TraceRID, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(toJSON), &e.To)
	json.Unmarshal([]byte(ccJSON), &e.Cc)
	json.Unmarshal([]byte(attachJSON), &e.Attachments)
	return &e, nil
}

func (s *Store) GetEmailByTrace(traceRID string) (*Email, error) {
	var e Email
	var toJSON, ccJSON, attachJSON string
	err := s.db.QueryRow(
		`SELECT id, from_addr, to_addrs, cc_addrs, subject, html, text_body, attachments_json, trace_rid, created_at
		FROM emails WHERE trace_rid = ? LIMIT 1`, traceRID,
	).Scan(&e.ID, &e.From, &toJSON, &ccJSON, &e.Subject, &e.HTML, &e.Text, &attachJSON, &e.TraceRID, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal([]byte(toJSON), &e.To)
	json.Unmarshal([]byte(ccJSON), &e.Cc)
	json.Unmarshal([]byte(attachJSON), &e.Attachments)
	return &e, nil
}

func (s *Store) GetEmailAttachment(emailID string, index int) (*EmailAttachment, error) {
	var attachJSON string
	err := s.db.QueryRow("SELECT attachments_json FROM emails WHERE id = ?", emailID).Scan(&attachJSON)
	if err != nil {
		return nil, err
	}

	var attachments []EmailAttachment
	json.Unmarshal([]byte(attachJSON), &attachments)

	if index >= len(attachments) {
		return nil, fmt.Errorf("attachment index out of range")
	}

	att := &attachments[index]
	if att.ContentBase64 != "" {
		decoded, err := base64Decode(att.ContentBase64)
		if err == nil {
			att.Content = decoded
		}
	}
	return att, nil
}

func base64Decode(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}
