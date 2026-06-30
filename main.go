package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
)

const broker = "localhost:9092"

var topicRoutes = map[string]string{
	"csv":   "mock-service-2",
	"excel": "mock-service-3",
	"pdf":   "mock-service-4",
}

type Message struct {
	Message string `json:"message"`
}

func main() {
	consumer, err := kgo.NewClient(
		kgo.SeedBrokers(broker),
		kgo.ConsumeTopics("mock-service-1"),
		kgo.ConsumerGroup("orchestrator-group"),
	)
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}
	defer consumer.Close()

	producer, err := kgo.NewClient(kgo.SeedBrokers(broker))
	if err != nil {
		log.Fatalf("failed to create producer: %v", err)
	}
	defer producer.Close()

	go func() {
		ctx := context.Background()
		log.Println("Consuming from mock-service-1...")
		for {
			fetches := consumer.PollFetches(ctx)
			if errs := fetches.Errors(); len(errs) > 0 {
				for _, e := range errs {
					log.Printf("fetch error: %v", e)
				}
				continue
			}

			fetches.EachRecord(func(r *kgo.Record) {
				var msg Message
				if err := json.Unmarshal(r.Value, &msg); err != nil {
					log.Printf("failed to parse message: %v", err)
					return
				}

				fileType := extractFileType(msg.Message)
				target, ok := topicRoutes[fileType]
				if !ok {
					log.Printf("unknown file type %q, dropping message", fileType)
					return
				}

				producer.Produce(ctx, &kgo.Record{
					Topic: target,
					Value: r.Value,
				}, func(_ *kgo.Record, err error) {
					if err != nil {
						log.Printf("failed to publish to %s: %v", target, err)
					} else {
						log.Printf("routed [%s] -> %s: %s", fileType, target, msg.Message)
					}
				})
			})
		}
	}()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"orchestrator"}`))
	})

	log.Println("Orchestrator running on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}

// extractFileType parses the "[CSV] ..." prefix set by the Streamlit UI.
func extractFileType(message string) string {
	if !strings.HasPrefix(message, "[") {
		return ""
	}
	end := strings.Index(message, "]")
	if end == -1 {
		return ""
	}
	return strings.ToLower(message[1:end])
}
