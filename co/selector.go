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
	ErrWeight = 0

	// BlockWeight means the node is blocked manually
	// it will never be selected unless it's recover manually
	BlockWeight = -1
)

// NewSelector constructs a Selector instance
func NewSelector() (*Selector, error) {
	sel := &Selector{}
	sel.weight = make(map[string]int)
	sel.all.addrs = make([]string, 0, 64)
	sel.all.nodes = map[string]*Node{}

	sel.selectALG = SWRRA()

	return sel, nil
}

// Selector is used to select a best chain node to route the requests to
type Selector struct {
	lk     sync.RWMutex
	weight map[string]int

	selectALG func(map[string]int) (string, error)

	all struct {
		sync.RWMutex
		addrs []string
		nodes map[string]*Node
	}
}

// ReplaceNodes adds and removes nodes
func (s *Selector) ReplaceNodes(add []*Node, removes map[string]bool, removesAll bool) {
	s.all.Lock()

	// reset
	if removesAll || len(removes) > 0 {
		s.all.addrs = s.all.addrs[:0]
		for host := range s.all.nodes {
			if removesAll || removes[host] {
				s.all.nodes[host].Stop() // nolint:errcheck
				delete(s.all.nodes, host)
				continue
			}

			s.all.addrs = append(s.all.addrs, host)
		}
	}

	for i := range add {
		current := add[i]
		if prev, has := s.all.nodes[current.info.Addr]; has {
			prev.Stop() // nolint:errcheck
		} else {
			s.all.addrs = append(s.all.addrs, current.info.Addr)
		}

		s.all.nodes[current.info.Addr] = current
		go current.Start()
	}
	s.all.Unlock()

	s.lk.Lock()
	newWeight := make(map[string]int, len(s.weight))
	for addr, _ := range s.all.nodes {
		if w, ok := s.weight[addr]; ok {
			newWeight[addr] = w
		} else {
			newWeight[addr] = DefaultWeight
		}
	}
	s.weight = newWeight
	s.lk.Unlock()
}

func (s *Selector) getPriority(adds string) int {
	s.lk.RLock()
	defer s.lk.RUnlock()
	return s.weight[adds]
}

func (s *Selector) setPriority(addr string, priority int) {
	s.lk.Lock()
	defer s.lk.Unlock()

	if priority < -1 {
		priority = -1
	} else if priority > MaxValidWeight {
		priority = MaxValidWeight
	}

	s.lk.Lock()
	w := s.weight[addr]
	before := w
	w = priority

	s.weight[addr] = w
	log.Warnf("change priority of %s from %d to %d", addr, before, w)
	s.lk.Unlock()
}

func (s *Selector) ListWeight() map[string]int {
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
	blockQue := make([]string, 0)
	errQue := make([]string, 0)
	normalQue := make(map[string]int)

	s.lk.Lock()
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

	return s.all.nodes[addr], nil
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
	return func(weight map[string]int) (string, error) {
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
