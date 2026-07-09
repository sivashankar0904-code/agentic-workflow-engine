package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/twmb/franz-go/pkg/kgo"
	"gopkg.in/yaml.v3"
)

const dagFile = "dag.yaml"

func broker() string {
	if b := os.Getenv("KAFKA_BROKER"); b != "" {
		return b
	}
	return "localhost:9092"
}

// ── YAML schema ───────────────────────────────────────────────────────────────

type Condition struct {
	Field    string `yaml:"field"`
	Contains string `yaml:"contains"`
}

type Rule struct {
	Condition Condition `yaml:"condition"`
	Target    string    `yaml:"target"`
}

type Routing struct {
	Source string `yaml:"source"`
	Rules  []Rule `yaml:"rules"`
}

type Service struct {
	Name  string `yaml:"name"`
	Host  string `yaml:"host"`
	Port  int    `yaml:"port"`
	Topic string `yaml:"topic"`
}

type DAG struct {
	Services []Service `yaml:"services"`
	Routing  Routing   `yaml:"routing"`
}

// ── Live config (hot-reloadable) ──────────────────────────────────────────────

var (
	dagMu  sync.RWMutex
	dag    DAG
)

func loadDAG(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var d DAG
	if err := yaml.Unmarshal(data, &d); err != nil {
		return err
	}
	dagMu.Lock()
	dag = d
	dagMu.Unlock()
	log.Printf("DAG loaded: source=%s rules=%d", d.Routing.Source, len(d.Routing.Rules))
	return nil
}

func route(msg map[string]any) string {
	dagMu.RLock()
	defer dagMu.RUnlock()
	for _, rule := range dag.Routing.Rules {
		val, ok := msg[rule.Condition.Field]
		if !ok {
			continue
		}
		if strings.Contains(strings.ToLower(fmt.Sprintf("%v", val)), strings.ToLower(rule.Condition.Contains)) {
			return rule.Target
		}
	}
	return ""
}

func main() {
	if err := loadDAG(dagFile); err != nil {
		log.Fatalf("failed to load DAG: %v", err)
	}

	consumer, err := kgo.NewClient(
		kgo.SeedBrokers(broker()),
		kgo.ConsumeTopics(dag.Routing.Source),
		kgo.ConsumerGroup("orchestrator-group"),
	)
	if err != nil {
		log.Fatalf("failed to create consumer: %v", err)
	}
	defer consumer.Close()

	producer, err := kgo.NewClient(kgo.SeedBrokers(broker()))
	if err != nil {
		log.Fatalf("failed to create producer: %v", err)
	}
	defer producer.Close()

	go func() {
		ctx := context.Background()
		log.Printf("Consuming from %s...", dag.Routing.Source)
		for {
			fetches := consumer.PollFetches(ctx)
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
				target := route(msg)
				if target == "" {
					log.Printf("no rule matched for message: %v", msg)
					return
				}
				producer.Produce(ctx, &kgo.Record{
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
	}()

	// ── HTTP endpoints ────────────────────────────────────────────────────────

	cors := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next(w, r)
		}
	}

	http.HandleFunc("/health", cors(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"orchestrator"}`))
	}))

	// GET /config — return current DAG as JSON
	http.HandleFunc("/config", cors(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			return
		}
		if r.Method == http.MethodGet {
			dagMu.RLock()
			defer dagMu.RUnlock()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(dag)
			return
		}
		// POST /config — upload new YAML, hot-reload
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read body", http.StatusBadRequest)
				return
			}
			var d DAG
			if err := yaml.Unmarshal(body, &d); err != nil {
				http.Error(w, "invalid YAML: "+err.Error(), http.StatusBadRequest)
				return
			}
			if err := os.WriteFile(dagFile, body, 0644); err != nil {
				http.Error(w, "failed to save config", http.StatusInternalServerError)
				return
			}
			dagMu.Lock()
			dag = d
			dagMu.Unlock()
			log.Printf("DAG hot-reloaded: source=%s rules=%d", d.Routing.Source, len(d.Routing.Rules))
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status":"reloaded"}`))
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}))

	log.Println("Orchestrator running on :8000")
	log.Fatal(http.ListenAndServe(":8000", nil))
}
