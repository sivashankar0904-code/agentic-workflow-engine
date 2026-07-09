package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/twmb/franz-go/pkg/kgo"

	"orchestrator/internals/config"
	"orchestrator/internals/router"
)

// Broker resolves the Kafka broker address from KAFKA_BROKER, defaulting to
// localhost:9092.
func Broker() string {
	if b := os.Getenv("KAFKA_BROKER"); b != "" {
		return b
	}
	return "localhost:9092"
}

// Orchestrator consumes from the DAG source topic, routes each message, and
// produces it to the matched target topic.
type Orchestrator struct {
	store    *config.Store
	consumer *kgo.Client
	producer *kgo.Client
}

// New builds the consumer/producer clients for the given store's source topic.
func New(store *config.Store) (*Orchestrator, error) {
	source := store.Get().Routing.Source

	consumer, err := kgo.NewClient(
		kgo.SeedBrokers(Broker()),
		kgo.ConsumeTopics(source),
		kgo.ConsumerGroup("orchestrator-group"),
	)
	if err != nil {
		return nil, err
	}

	producer, err := kgo.NewClient(kgo.SeedBrokers(Broker()))
	if err != nil {
		consumer.Close()
		return nil, err
	}

	return &Orchestrator{store: store, consumer: consumer, producer: producer}, nil
}

// Close releases the underlying Kafka clients.
func (o *Orchestrator) Close() {
	o.consumer.Close()
	o.producer.Close()
}

// Run blocks, polling the source topic and routing messages until ctx is done.
func (o *Orchestrator) Run(ctx context.Context) {
	log.Printf("Consuming from %s...", o.store.Get().Routing.Source)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		fetches := o.consumer.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				log.Printf("fetch error: %v", e)
			}
			continue
		}
		fetches.EachRecord(func(r *kgo.Record) {
			var msg map[string]any
			if err := json.Unmarshal(r.Value, &msg); err != nil {
				log.Printf("parse error: %v", err)
				return
			}
			target := router.Match(o.store.Get(), msg)
			if target == "" {
				log.Printf("no rule matched for message: %v", msg)
				return
			}
			o.producer.Produce(ctx, &kgo.Record{
				Topic: target,
				Value: r.Value,
			}, func(_ *kgo.Record, err error) {
				if err != nil {
					log.Printf("produce error -> %s: %v", target, err)
				} else {
					log.Printf("routed -> %s: %v", target, msg)
				}
			})
		})
	}
}
