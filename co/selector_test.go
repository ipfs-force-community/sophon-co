package co

import (
	"testing"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/ipfs/go-cid"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_Selector_Init(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeStore := NewMockINodeStore(ctrl)

	// init selector
	sel, _ := NewSelector(nodeStore)
	nodeStore.EXPECT().AddNodes(gomock.Any())
	sel.AddNodes(
		&Node{Addr: "a"},
		&Node{Addr: "b"},
		&Node{Addr: "c"},
		&Node{Addr: "d"},
	)

	weight := sel.ListWeight()
	assert.Equal(t, 4, len(weight))
	assert.Equal(t, DefaultWeight, weight["a"])
	assert.Equal(t, DefaultWeight, weight["b"])
	assert.Equal(t, DefaultWeight, weight["d"])

	priorities := sel.ListPriority()
	assert.Equal(t, 4, len(priorities))
	assert.Equal(t, DelayPriority, priorities["a"])
	assert.Equal(t, DelayPriority, priorities["b"])
	assert.Equal(t, DelayPriority, priorities["d"])
}

func genBlockHeader(t *testing.T) *types.BlockHeader {
	addr, err := address.NewIDAddress(12512063)
	assert.NoError(t, err)

	c, err := cid.Decode("bafyreicmaj5hhoy5mgqvamfhgexxyergw7hdeshizghodwkjg6qmpoco7i")
	assert.NoError(t, err)

	return &types.BlockHeader{
		Miner: addr,
		Ticket: &types.Ticket{
			VRFProof: []byte("vrf proof0000000vrf proof0000000"),
		},
		ElectionProof: &types.ElectionProof{
			VRFProof: []byte("vrf proof0000000vrf proof0000000"),
		},
		Parents:               []cid.Cid{c, c},
		ParentMessageReceipts: c,
		BLSAggregate:          &crypto.Signature{Type: crypto.SigTypeBLS, Data: []byte("boo! im a signature")},
		ParentWeight:          types.NewInt(123125126212),
		Messages:              c,
		Height:                85919298723,
		ParentStateRoot:       c,
		BlockSig:              &crypto.Signature{Type: crypto.SigTypeBLS, Data: []byte("boo! im a signature")},
		ParentBaseFee:         types.NewInt(3432432843291),
	}
}

func TestSelect(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeStore := NewMockINodeStore(ctrl)

	// init selector
	sel, _ := NewSelector(nodeStore)
	nodeStore.EXPECT().AddNodes(gomock.Any())

	var nodes []*Node
	nodeMap := make(map[string]*Node, 3)
	for _, addr := range []string{"a", "b", "c"} {
		blkCache, err := newBlockHeaderCache(20)
		assert.NoError(t, err)
		node := &Node{
			Addr:     addr,
			blkCache: blkCache,
		}
		nodes = append(nodes, node)
		nodeMap[addr] = node

		sel.setPriority(DelayPriority, addr)
	}

	ts, err := types.NewTipSet([]*types.BlockHeader{genBlockHeader(t)})
	assert.NoError(t, err)
	nodeMap["c"].blkCache.add([]*api.HeadChange{
		{Val: ts},
	})

	sel.AddNodes(nodes...)

	nodeStore.EXPECT().GetNode(gomock.Any()).DoAndReturn(
		func(arg string) interface{} {
			node, ok := nodeMap[arg]
			if !ok {
				panic("not found node: " + arg)
			}

			return node
		},
	).AnyTimes()

	addrs := []string{"a", "b"}
	for i := 0; i < 100000; i++ {
		addr := addrs[i%2]
		sel.setPriority(CatchUpPriority, addr)

		node, err := sel.Select(types.EmptyTSK)
		assert.NoError(t, err)
		assert.Equal(t, addr, node.Addr)

		node, err = sel.Select(ts.Key())
		assert.NoError(t, err)
		assert.Equal(t, "c", node.Addr)

		sel.setPriority(DelayPriority, addr)
	}
}

func Test_Selector_UpdateNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeStore := NewMockINodeStore(ctrl)

	// init selector
	sel, _ := NewSelector(nodeStore)
	nodeStore.EXPECT().AddNodes(gomock.Any())
	nodes := []*Node{{Addr: "a"},
		{Addr: "b"},
		{Addr: "c"}}
	sel.AddNodes(nodes...)

	nodeStore.EXPECT().AddNodes(gomock.Any())
	sel.AddNodes(
		&Node{Addr: "a"},
	)

	assert.Equal(t, 3, len(sel.ListWeight()))
	assert.Equal(t, 3, len(sel.ListPriority()))

	records := make([]string, 0)
	for i := 0; i < 3; i++ {
		nodeStore.EXPECT().GetNode(gomock.Any()).AnyTimes().DoAndReturn(func(arg0 string) *Node {
			for _, node := range nodes {
				if node.Addr == arg0 {
					return node
				}
			}
			return nil
		})
		node, err := sel.Select(types.EmptyTSK)
		records = append(records, node.Addr)
		assert.NoError(t, err)
	}
	dict := countSlice(records)
	assert.Equal(t, 1, dict["a"])
	assert.Equal(t, 1, dict["c"])
}

func Test_Selector_SetWeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeStore := NewMockINodeStore(ctrl)

	// init selector
	sel, _ := NewSelector(nodeStore)

	nodeStore.EXPECT().AddNodes(gomock.Any())
	nodes := []*Node{{Addr: "a"},
		{Addr: "b"},
		{Addr: "c"}}
	sel.AddNodes(nodes...)

	// test setPriority
	sel.SetWeight("a", 3) // nolint:errcheck
	weight := sel.ListWeight()
	assert.Equal(t, 3, weight["a"])
	t.Log(weight)

	records := make([]string, 0)
	for i := 0; i < 5; i++ {
		nodeStore.EXPECT().GetNode(gomock.Any()).AnyTimes().DoAndReturn(func(arg0 string) *Node {
			for _, node := range nodes {
				if node.Addr == arg0 {
					return node
				}
			}
			return nil
		})
		node, err := sel.Select(types.EmptyTSK)
		records = append(records, node.Addr)
		assert.NoError(t, err)
	}
	dict := countSlice(records)
	t.Log(records)
	assert.Equal(t, 3, dict["a"])
	assert.Equal(t, 1, dict["c"])
	assert.Equal(t, 1, dict["b"])
}

func Test_Selector_SetWeight_BlockWeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeStore := NewMockINodeStore(ctrl)

	// init selector
	sel, _ := NewSelector(nodeStore)

	nodeStore.EXPECT().AddNodes(gomock.Any())
	nodes := []*Node{{Addr: "a"},
		{Addr: "b"},
		{Addr: "c"}}
	sel.AddNodes(nodes...)

	// test setPriority blockWeight
	sel.SetWeight("a", BlockWeight)   // nolint:errcheck
	sel.SetWeight("c", BlockWeight)   // nolint:errcheck
	sel.SetWeight("b", DefaultWeight) // nolint:errcheck
	weights := sel.ListWeight()
	assert.Equal(t, BlockWeight, weights["a"])
	assert.Equal(t, BlockWeight, weights["c"])
	assert.Equal(t, DefaultWeight, weights["b"])
	records := make([]string, 0)
	for i := 0; i < 3; i++ {
		nodeStore.EXPECT().GetNode(gomock.Any()).AnyTimes().DoAndReturn(func(arg0 string) *Node {
			for _, node := range nodes {
				if node.Addr == arg0 {
					return node
				}
			}
			return nil
		})
		node, err := sel.Select(types.EmptyTSK)
		records = append(records, node.Addr)
		assert.NoError(t, err)
	}
	dict := countSlice(records)
	assert.Equal(t, 0, dict["a"])
	assert.Equal(t, 0, dict["c"])
	assert.Equal(t, 3, dict["b"])

	sel.SetWeight("a", BlockWeight) // nolint:errcheck
	sel.SetWeight("c", BlockWeight) // nolint:errcheck
	sel.SetWeight("b", BlockWeight) // nolint:errcheck
	_, err := sel.Select(types.EmptyTSK)
	assert.Error(t, err)
	assert.Equal(t, ErrNoNodeAvailable, err)
}

func Test_Selector_SetPriority(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeStore := NewMockINodeStore(ctrl)

	sel, _ := NewSelector(nodeStore)
	nodeStore.EXPECT().AddNodes(gomock.Any())
	nodes := []*Node{{Addr: "a"},
		{Addr: "b"},
		{Addr: "c"}}
	sel.AddNodes(nodes...)
	nodeStore.EXPECT().GetNode(gomock.Any()).AnyTimes().DoAndReturn(func(arg0 string) *Node {
		for _, node := range nodes {
			if node.Addr == arg0 {
				return node
			}
		}
		return nil
	})

	// a 2 b 1 c 1
	sel.setPriority(CatchUpPriority, "a")
	records := make([]string, 0)
	for i := 0; i < 4; i++ {
		node, err := sel.Select(types.EmptyTSK)
		records = append(records, node.Addr)
		assert.NoError(t, err)
	}
	dict := countSlice(records)
	assert.Equal(t, 4, dict["a"])

	// a 2 b 2 c 1
	sel.setPriority(CatchUpPriority, "a")
	sel.setPriority(CatchUpPriority, "b")
	sel.setPriority(DelayPriority, "c")
	records = make([]string, 0)
	for i := 0; i < 4; i++ {
		node, err := sel.Select(types.EmptyTSK)
		records = append(records, node.Addr)
		assert.NoError(t, err)
	}
	dict = countSlice(records)
	assert.Equal(t, 2, dict["a"])
	assert.Equal(t, 2, dict["b"])

	// a 2 b 2 c 2
	sel.setPriority(CatchUpPriority, "a")
	sel.setPriority(CatchUpPriority, "b")
	sel.setPriority(CatchUpPriority, "c")
	records = make([]string, 0)
	for i := 0; i < 3; i++ {
		node, err := sel.Select(types.EmptyTSK)
		records = append(records, node.Addr)
		assert.NoError(t, err)
	}
	dict = countSlice(records)
	assert.Equal(t, 1, dict["a"])
	assert.Equal(t, 1, dict["b"])

	// a 0 b 0 c 0
	sel.setPriority(ErrPriority, "a")
	sel.setPriority(ErrPriority, "b")
	sel.setPriority(ErrPriority, "c")
	records = make([]string, 0)
	for i := 0; i < 3; i++ {
		node, err := sel.Select(types.EmptyTSK)
		records = append(records, node.Addr)
		assert.NoError(t, err)
	}
	dict = countSlice(records)
	assert.Equal(t, 1, dict["a"])
	assert.Equal(t, 1, dict["b"])

	// a 1 b 0 c 0
	sel.setPriority(DelayPriority, "a")
	sel.setPriority(ErrPriority, "b")
	sel.setPriority(ErrPriority, "c")
	records = make([]string, 0)
	for i := 0; i < 3; i++ {
		node, err := sel.Select(types.EmptyTSK)
		records = append(records, node.Addr)
		assert.NoError(t, err)
	}
	dict = countSlice(records)
	assert.Equal(t, 3, dict["a"])
}

func Test_Selector_SetPriority_SetWeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeStore := NewMockINodeStore(ctrl)

	sel, _ := NewSelector(nodeStore)
	nodeStore.EXPECT().AddNodes(gomock.Any())
	nodes := []*Node{{Addr: "a"},
		{Addr: "b"},
		{Addr: "c"}}
	sel.AddNodes(nodes...)
	nodeStore.EXPECT().GetNode(gomock.Any()).AnyTimes().DoAndReturn(func(arg0 string) *Node {
		for _, node := range nodes {
			if node.Addr == arg0 {
				return node
			}
		}
		return nil
	})

	// a 2 b 1 c 0
	sel.setPriority(CatchUpPriority, "a")
	sel.setPriority(DelayPriority, "b")
	sel.setPriority(ErrPriority, "c")
	records := make([]string, 0)
	for i := 0; i < 4; i++ {
		node, err := sel.Select(types.EmptyTSK)
		records = append(records, node.Addr)
		assert.NoError(t, err)
	}
	dict := countSlice(records)
	assert.Equal(t, 4, dict["a"])

	// wa 0
	sel.SetWeight("a", BlockWeight) // nolint:errcheck
	records = make([]string, 0)
	for i := 0; i < 4; i++ {
		node, err := sel.Select(types.EmptyTSK)
		records = append(records, node.Addr)
		assert.NoError(t, err)
	}
	dict = countSlice(records)
	assert.Equal(t, 4, dict["b"])

	// wa 0 wb 0
	sel.SetWeight("a", BlockWeight) // nolint:errcheck
	sel.SetWeight("b", BlockWeight) // nolint:errcheck
	records = make([]string, 0)
	for i := 0; i < 4; i++ {
		node, err := sel.Select(types.EmptyTSK)
		records = append(records, node.Addr)
		assert.NoError(t, err)
	}
	dict = countSlice(records)
	assert.Equal(t, 4, dict["c"])

	// wa 0 wb 0 wc 0
	sel.SetWeight("a", BlockWeight) // nolint:errcheck
	sel.SetWeight("b", BlockWeight) // nolint:errcheck
	sel.SetWeight("c", BlockWeight) // nolint:errcheck
	_, err := sel.Select(types.EmptyTSK)
	assert.Equal(t, ErrNoNodeAvailable, err)
}

func Test_SWRRA(t *testing.T) {
	cases := []map[string]int{
		{"a": 1, "b": 2},
		{"a": 1, "b": 2, "c": 3},
		{"a": 1, "b": 1, "c": 3},
		{"a": 2, "b": 2, "c": 6},
		{"a": 1, "b": 2, "c": 3, "d": 4},
		{"a": 1, "b": 1, "c": 1, "d": 1},
		{"a": 0, "b": 1, "c": 5, "d": 1},
		{"a": 0, "b": 10},
	}

	testCase := func(weight map[string]int) {
		alg := SWRRA()

		count := sumWeight(weight)
		res := make([]string, 0, count)
		for i := 0; i < count; i++ {
			k, _ := alg(weight)
			res = append(res, k)
		}
		t.Log(res)
		resMap := countSlice(res)
		for k, v := range weight {
			assert.Equal(t, v, resMap[k])
		}
	}

	for _, c := range cases {
		testCase(c)
	}
}

func Test_SWRRA_Err(t *testing.T) {
	cases := []map[string]int{
		{"a": 0, "b": 0, "c": 0},
		{},
	}

	testCase := func(weight map[string]int) {
		alg := SWRRA()
		k, err := alg(weight)
		assert.Error(t, err)
		assert.Equal(t, "", k)
	}

	for _, c := range cases {
		testCase(c)
	}
}

func countSlice(slice []string) map[string]int {
	res := make(map[string]int)
	for _, v := range slice {
		res[v]++
	}
	return res
}

func sumWeight(weight map[string]int) int {
	res := 0
	for _, v := range weight {
		res += v
	}
	return res
}
