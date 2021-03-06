package state

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	"github.com/docker/libkv"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/boltdb"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/pborman/uuid"
	"github.com/sonm-io/core/insonmnia/hardware"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

// todo: candidate to move into the `structs` package
type askPlan struct {
	ID    string
	Order *structs.Order
}

type stateJSON struct {
	AskPlans   map[string]*askPlan `json:"ask_plans"`
	Benchmarks map[uint64]bool     `json:"benchmarks"`
	Hardware   *hardware.Hardware  `json:"hardware"`
	HwHash     string              `json:"hw_hash"`
}

type Storage struct {
	mu  sync.Mutex
	ctx context.Context

	store store.Store
	data  *stateJSON
}

type StorageConfig struct {
	Endpoint string `yaml:"endpoint" required:"true" default:"/var/lib/sonm/worker.boltdb"`
	Bucket   string `yaml:"bucket" required:"true" default:"sonm"`
}

func makeStore(ctx context.Context, cfg *StorageConfig) (store.Store, error) {
	boltdb.Register()
	log.G(ctx).Info("creating store", zap.Any("store", cfg))

	config := store.Config{
		Bucket: cfg.Bucket,
	}

	return libkv.NewStore(store.BOLTDB, []string{cfg.Endpoint}, &config)
}

func newEmptyState() *stateJSON {
	return &stateJSON{
		AskPlans:   make(map[string]*askPlan, 0),
		Benchmarks: make(map[uint64]bool),
		Hardware:   new(hardware.Hardware),
	}
}

func NewState(ctx context.Context, cfg *StorageConfig) (*Storage, error) {
	clusterStore, err := makeStore(ctx, cfg)
	if err != nil {
		return nil, err
	}

	out := &Storage{
		ctx:   ctx,
		store: clusterStore,
		data:  newEmptyState(),
	}

	if err := out.loadInitial(); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Storage) dump() error {
	bin, err := json.Marshal(s.data)
	if err != nil {
		return err
	}

	return s.store.Put("state", bin, &store.WriteOptions{})
}

func (s *Storage) loadInitial() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	kv, err := s.store.Get("state")
	if err != nil && err != store.ErrKeyNotFound {
		return err
	}

	if kv != nil {
		// unmarshal exiting state
		if err := json.Unmarshal(kv.Value, &s.data); err != nil {
			return err
		}
	}

	return s.dump()
}

func (s *Storage) AskPlans() map[string]*pb.Slot {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]*pb.Slot)
	for id, plan := range s.data.AskPlans {
		result[id] = plan.Order.Slot
	}

	return result
}

func (s *Storage) CreateAskPlan(order *structs.Order) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	plan := askPlan{
		ID:    uuid.New(),
		Order: order,
	}

	s.data.AskPlans[plan.ID] = &plan
	if err := s.dump(); err != nil {
		return "", err
	}

	return plan.ID, nil
}

func (s *Storage) RemoveAskPlan(planID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.data.AskPlans[planID]
	if !ok {
		return errors.New("specified ask-plan does not exist")
	}

	delete(s.data.AskPlans, planID)
	return s.dump()
}

func (s *Storage) PassedBenchmarks() map[uint64]bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data.Benchmarks
}

func (s *Storage) SetPassedBenchmarks(v map[uint64]bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Benchmarks = v
	return s.dump()
}

func (s *Storage) HardwareHash() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data.HwHash
}

func (s *Storage) SetHardwareHash(v string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.HwHash = v
	return s.dump()
}

func (s *Storage) HardwareWithBenchmarks() *hardware.Hardware {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.data.Hardware
}

func (s *Storage) SetHardwareWithBenchmarks(hw *hardware.Hardware) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data.Hardware = hw
	return s.dump()
}
