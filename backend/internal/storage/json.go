package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"simarena/internal/models"
)

// JSONStore is a file-based JSON storage for simulations.
type JSONStore struct {
	mu       sync.Mutex
	filePath string
}

// NewJSONStore creates a new JSON file store at the given directory.
func NewJSONStore(dataDir string) (*JSONStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	return &JSONStore{
		filePath: filepath.Join(dataDir, "simulations.json"),
	}, nil
}

func (s *JSONStore) readAll() ([]models.Simulation, error) {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Simulation{}, nil
		}
		return nil, fmt.Errorf("read file: %w", err)
	}
	if len(data) == 0 {
		return []models.Simulation{}, nil
	}
	var sims []models.Simulation
	if err := json.Unmarshal(data, &sims); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}
	return sims, nil
}

func (s *JSONStore) writeAll(sims []models.Simulation) error {
	data, err := json.MarshalIndent(sims, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	if err := os.WriteFile(s.filePath, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// List returns all simulations.
func (s *JSONStore) List() ([]models.Simulation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.readAll()
}

// Get returns a single simulation by ID.
func (s *JSONStore) Get(id string) (*models.Simulation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sims, err := s.readAll()
	if err != nil {
		return nil, err
	}
	for i := range sims {
		if sims[i].ID == id {
			return &sims[i], nil
		}
	}
	return nil, fmt.Errorf("simulation %s not found", id)
}

// Create adds a new simulation to the store.
func (s *JSONStore) Create(sim models.Simulation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sims, err := s.readAll()
	if err != nil {
		return err
	}
	sims = append(sims, sim)
	return s.writeAll(sims)
}

// Update replaces a simulation in the store by ID.
func (s *JSONStore) Update(sim models.Simulation) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sims, err := s.readAll()
	if err != nil {
		return err
	}
	for i := range sims {
		if sims[i].ID == sim.ID {
			sims[i] = sim
			return s.writeAll(sims)
		}
	}
	return fmt.Errorf("simulation %s not found", sim.ID)
}

// Delete removes a simulation by ID.
func (s *JSONStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	sims, err := s.readAll()
	if err != nil {
		return err
	}
	filtered := make([]models.Simulation, 0, len(sims))
	found := false
	for _, sim := range sims {
		if sim.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, sim)
	}
	if !found {
		return fmt.Errorf("simulation %s not found", id)
	}
	return s.writeAll(filtered)
}
