package co

import (
	"fmt"
	"math/rand"
	"sync"
)

const (
	// MaxWeight is the default max weight of a node
	MaxValidWeight = 5

	// DefaultWeight is the default weight of a node
	DefaultWeight = 2

	// ErrWeight means the node respond with error
	// it will never be selected unless it's the only nodes that available
	ErrWeight = 1

	// BlockWeight means the node is blocked manually
	// it will never be selected unless it's recover manually
	BlockWeight = 0
)

// NewSelector constructs a Selector instance
func NewSelector() (*Selector, error) {
	sel := &Selector{}
	sel.weight = make(map[string]int)
	sel.selectALG = SWRRA()
	return sel, nil
}

// Selector is used to select a best chain node to route the requests to
type Selector struct {
	lk     sync.RWMutex
	weight map[string]int

	selectALG func(map[string]int) (string, error)

	nodeProvider INodeProvider
}

func (s *Selector) SetNodeProvider(provider INodeProvider) {
	s.nodeProvider = provider
	provider.AddHook(func(add map[string]bool) {
		s.lk.Lock()
		defer s.lk.Unlock()
		for addr, alter := range add {
			if alter == ADD {
				if _, ok := s.weight[addr]; !ok {
					s.weight[addr] = DefaultWeight
				}
			} else {
				delete(s.weight, addr)
			}
		}
	})

	initNodes := provider.GetHosts()
	s.lk.Lock()
	defer s.lk.Unlock()
	for _, node := range initNodes {
		s.weight[node] = DefaultWeight
	}
}

func (s *Selector) getPriority(adds string) int {
	s.lk.RLock()
	defer s.lk.RUnlock()
	return s.weight[adds]
}

func (s *Selector) setPriority(addr string, priority int) {
	s.lk.Lock()
	defer s.lk.Unlock()

	if priority < BlockWeight {
		priority = BlockWeight
	} else if priority > MaxValidWeight {
		priority = MaxValidWeight
	}

	w := s.weight[addr]
	before := w
	w = priority

	s.weight[addr] = w
	log.Warnf("change priority of %s from %d to %d", addr, before, w)
}

func (s *Selector) ListPriority() map[string]int {
	s.lk.RLock()
	defer s.lk.RUnlock()
	newWeight := make(map[string]int, len(s.weight))
	for addr, w := range s.weight {
		newWeight[addr] = w
	}
	return newWeight
}

// Select tries to choose a node from the candidates
func (s *Selector) Select() (*Node, error) {
	s.lk.RLock()
	defer s.lk.RUnlock()

	blockQue := make([]string, 0)
	errQue := make([]string, 0)
	normalQue := make(map[string]int)

	for addr, w := range s.weight {
		if w <= BlockWeight {
			blockQue = append(blockQue, addr)
		} else if w == ErrWeight {
			errQue = append(errQue, addr)
		} else {
			normalQue[addr] = w
		}
	}

	var addr string
	if len(normalQue) > 0 {
		addr, _ = s.selectALG(normalQue)
	} else if len(errQue) > 0 {
		addr = errQue[rand.Intn(len(errQue))]
	} else {
		return nil, ErrNoNodeAvailable
	}

	return s.nodeProvider.GetNode(addr), nil
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// Smooth Weight Round Robin Algorithm
// weight should be positive and len of weight shuold greater than 0
func SWRRA() func(map[string]int) (string, error) {
	state := make(map[string]int)
	lk := sync.Mutex{}
	return func(weight map[string]int) (string, error) {
		lk.Lock()
		defer lk.Unlock()
		// 0. check len of weight
		if len(weight) == 0 {
			return "", fmt.Errorf("no key available")
		}
		for {
			// 1. calc state
			for k, w := range weight {
				s, exist := state[k]
				if !exist {
					state[k] = 0
				}
				state[k] = s + int(w)
			}

			// 2. select biggest state of weight
			maxState := 0
			maxKey := ""
			for k, s := range state {
				if s > maxState {
					maxState = s
					maxKey = k
				}
			}

			// 3. deduct total weight
			totalWeight := 0
			for _, w := range weight {
				totalWeight += int(w)
			}
			state[maxKey] -= int(totalWeight)

			// 4. check if maxKey is available
			if _, exist := weight[maxKey]; exist {
				return maxKey, nil
			}
		}
	}
}

const (
	ADD    = true
	REMOVE = false
)
