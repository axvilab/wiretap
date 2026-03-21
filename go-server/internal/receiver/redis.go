package receiver

import (
	"compress/gzip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"time"

	"github.com/axvi/wiretap/internal/storage"
	"github.com/redis/go-redis/v9"
)

type RedisReceiver struct {
	client   *redis.Client
	queue    string
	ingestCh chan<- *storage.TracePayload
}

func NewRedis(addr, password string, db int, queue string, ch chan<- *storage.TracePayload) *RedisReceiver {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisReceiver{
		client:   client,
		queue:    queue,
		ingestCh: ch,
	}
}

func (r *RedisReceiver) Run(ctx context.Context) {
	log.Printf("Redis receiver listening on queue: %s", r.queue)

	for {
		select {
		case <-ctx.Done():
			r.client.Close()
			return
		default:
		}

		// BRPOP with 1 second timeout — allows checking ctx.Done()
		result, err := r.client.BRPop(ctx, 1*time.Second, r.queue).Result()
		if err != nil {
			if err == redis.Nil || err == context.Canceled || err == context.DeadlineExceeded {
				continue
			}
			log.Printf("[redis] error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if len(result) < 2 {
			continue
		}

		data := []byte(result[1])

		// Try gzip decompress
		body, err := tryDecompress(data)
		if err != nil {
			body = data // not compressed, use raw
		}

		var payload storage.TracePayload
		if err := json.Unmarshal(body, &payload); err != nil {
			log.Printf("[redis] invalid json: %v", err)
			continue
		}

		if payload.RID == "" {
			continue
		}

		log.Printf("[ingest:redis] %s %s %s (%d events)", payload.Type, payload.RID, redisMetaURI(&payload), len(payload.Events))
		r.ingestCh <- &payload
	}
}

func tryDecompress(data []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gr.Close()
	return io.ReadAll(gr)
}

func redisMetaURI(p *storage.TracePayload) string {
	var meta struct {
		Method string `json:"method"`
		URI    string `json:"uri"`
	}
	json.Unmarshal(p.Meta, &meta)
	if meta.Method != "" {
		return meta.Method + " " + meta.URI
	}
	return ""
}