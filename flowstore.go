package cofunc

import (
	"errors"
	"sync"

	"github.com/cofunclabs/cofunc/internal/flow"
)

// todo
type PersistentStore interface {
}

type FlowStore struct {
	sync.RWMutex
	entity  map[string]*flow.Flow
	backend PersistentStore
}

// Store store a kv into flowstore
func (s *FlowStore) Store(k string, f *flow.Flow) (err error) {
	s.Lock()
	defer s.Unlock()

	_, ok := s.entity[k]
	if ok {
		err = errors.New("repeat to store flow: " + k)
		return
	}
	s.entity[k] = f
	return nil
}

func (s *FlowStore) Update(k string, f *flow.Flow) (err error) {
	s.Lock()
	defer s.Unlock()

	_, ok := s.entity[k]
	if !ok {
		return errors.New("can't update, not found flow: " + k)
	}
	s.entity[k] = f
	return nil
}

func (s *FlowStore) Get(k string) (*flow.Flow, error) {
	s.RLock()
	defer s.RUnlock()
	v, ok := s.entity[k]
	if !ok {
		return nil, errors.New("not found flow: " + k)
	}
	return v, nil
}

// If 'do' return a error, will stop the 'Foreach'
func (s *FlowStore) Foreach(do func(string, *flow.Flow) error) {
	s.RLock()
	defer s.RUnlock()
	for k, v := range s.entity {
		if err := do(k, v); err != nil {
			return
		}
	}
}
