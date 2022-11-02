package co

import (
	"fmt"
	"sync"
)

type Priority int

const (
	// MaxWeight is the default max weight of a node
	MaxValidWeight = 10

	// DefaultWeight is the default weight of a node
	DefaultWeight = 1

	// BlockWeight means the node is blocked manually
	// it will never be selected unless it's recover manually
	BlockWeight = 0
)
const (
	// ErrPriority means the node once respond with error and will be selected with lowest priority
	ErrPriority = iota
	// DelayPriority means the node is behind the latest head
	DelayPriority
	// CatchUpPriority means the node have catch up the latest head, they will be selected with highest priority
	CatchUpPriority
)

// NewSelector constructs a Selector instance
func NewSelector() (*Selector, error) {
	sel := &Selector{}
	sel.weight = make(map[string]int)
	sel.priority = make(map[string]int)
	sel.selectALG = SWRRA()
	return sel, nil
}

// Selector is used to select a best chain node to route the requests to
type Selector struct {
	lk        sync.RWMutex
	weight    map[string]int
	priority  map[string]int
	selectALG func(map[string]int) (string, error)

	nodeProvider INodeProvider
}

func (s *Selector) SetNodeProvider(provider INodeProvider) {
	s.nodeProvider = provider
	provider.AddHook(func(add map[string]bool) {
		s.lk.Lock()
		defer s.lk.Unlock()
		for addr, alter := range add {
			if alter == ADD { // nolint:gosimple
				if _, ok := s.weight[addr]; !ok {
					s.weight[addr] = DefaultWeight
					s.priority[addr] = DelayPriority
				}
			} else {
				delete(s.weight, addr)
				delete(s.priority, addr)
			}
		}
	})

	initNodes := provider.GetHosts()
	s.lk.Lock()
	defer s.lk.Unlock()
	for _, node := range initNodes {
		s.weight[node] = DefaultWeight
		s.priority[node] = DelayPriority
	}
}

func (s *Selector) getAddrOfPriority(priority int) []string {
	s.lk.RLock()
	defer s.lk.RUnlock()
	ret := make([]string, 0)
	for addr, p := range s.priority {
		if p == priority {
			ret = append(ret, addr)
		}
	}
	return ret
}

func (s *Selector) setPriority(priority int, addrs ...string) {
	s.lk.Lock()
	defer s.lk.Unlock()

	if priority < ErrPriority {
		priority = ErrPriority
	} else if priority > CatchUpPriority {
		priority = CatchUpPriority
	}

	for _, addr := range addrs {
		w := s.priority[addr]
		before := w
		w = priority

		s.priority[addr] = w
		log.Warnf("change priority of %s from %d to %d", addr, before, w)
	}
}

func (s *Selector) ListPriority() map[string]int {
	s.lk.RLock()
	defer s.lk.RUnlock()
	ret := make(map[string]int)
	for addr, p := range s.priority {
		ret[addr] = p
	}
	return ret
}

func (s *Selector) SetWeight(addr string, weight int) error {
	s.lk.Lock()
	defer s.lk.Unlock()

	if weight < BlockWeight {
		return fmt.Errorf("priority must be greater than %d", BlockWeight)
	} else if weight > MaxValidWeight {
		return fmt.Errorf("priority must be less than %d", MaxValidWeight)
	}

	w, ok := s.weight[addr]
	if !ok {
		return fmt.Errorf("node %s not found", addr)
	}
	before := w
	w = weight

	s.weight[addr] = w
	log.Warnf("change priority of %s from %d to %d", addr, before, w)
	return nil
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
	s.lk.RLock()
	defer s.lk.RUnlock()

	errQue := make(map[string]int)
	delayQue := make(map[string]int)
	catchUpQue := make(map[string]int)

	for addr, p := range s.priority {
		if p == CatchUpPriority {
			catchUpQue[addr] = s.weight[addr]
		} else if p == DelayPriority {
			delayQue[addr] = s.weight[addr]
		} else {
			errQue[addr] = s.weight[addr]
		}
	}

	var addr string = ""
	if len(catchUpQue) > 0 {
		addr, _ = s.selectALG(catchUpQue)
	}
	if addr == "" && len(delayQue) > 0 {
		addr, _ = s.selectALG(delayQue)
	}
	if addr == "" && len(errQue) > 0 {
		addr, _ = s.selectALG(errQue)
	}

	if addr == "" {
		return nil, ErrNoNodeAvailable
	}

	return s.nodeProvider.GetNode(addr), nil
}

// Smooth Weight Round Robin Algorithm
// weight should be positive and len of weight shuold greater than 0
func SWRRA() func(map[string]int) (string, error) {
	state := make(map[string]int)
	lk := sync.Mutex{}
	return func(weight map[string]int) (string, error) {
		lk.Lock()
		defer lk.Unlock()
		// check len of weight
		if len(weight) == 0 {
			return "", fmt.Errorf("weight is empty")
		}

		// 0. exclude the nodes with weight 0
		selectSet := make(map[string]int, len(weight))
		for addr, w := range weight {
			if w > 0 {
				selectSet[addr] = w
			}
		}
		if len(selectSet) == 0 {
			return "", fmt.Errorf("no node available")
		}

		for {
			// 1. calc state
			for k, w := range selectSet {
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
			for _, w := range selectSet {
				totalWeight += int(w)
			}
			state[maxKey] -= int(totalWeight)

			// 4. check if maxKey is available
			if _, exist := selectSet[maxKey]; exist {
				return maxKey, nil
			}
		}
	}
}
