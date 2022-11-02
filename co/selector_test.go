package co

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_Selector_Init(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeProvider := NewMockINodeProvider(ctrl)

	// init selector
	sel, _ := NewSelector()
	nodeProvider.EXPECT().AddHook(gomock.Any())
	nodeProvider.EXPECT().GetHosts().Return([]string{"a", "b", "c", "d"})
	sel.SetNodeProvider(nodeProvider)

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

func Test_Selector_UpdateNodes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeProvider := NewMockINodeProvider(ctrl)

	// init selector
	sel, _ := NewSelector()
	var hook func(map[string]bool)
	nodeProvider.EXPECT().AddHook(gomock.Any()).Do(func(h func(map[string]bool)) {
		hook = h
	})
	nodeProvider.EXPECT().GetHosts().Return([]string{"a", "b", "c", "d"})
	sel.SetNodeProvider(nodeProvider)

	// update nodes
	hook(map[string]bool{"a": ADD, "b": REMOVE})
	assert.Equal(t, 3, len(sel.ListWeight()))
	assert.Equal(t, 3, len(sel.ListPriority()))

	records := make([]string, 0)
	for i := 0; i < 3; i++ {
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
		assert.NoError(t, err)
	}
	dict := countSlice(records)
	assert.Equal(t, 1, dict["a"])
	assert.Equal(t, 1, dict["c"])
	assert.Equal(t, 1, dict["d"])
}

func Test_Selector_SetWeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeProvider := NewMockINodeProvider(ctrl)

	// init selector
	sel, _ := NewSelector()

	nodeProvider.EXPECT().AddHook(gomock.Any())
	nodeProvider.EXPECT().GetHosts().Return([]string{"a", "b", "c"})
	sel.SetNodeProvider(nodeProvider)

	// test setPriority
	sel.SetWeight("a", 3)
	weight := sel.ListWeight()
	assert.Equal(t, 3, weight["a"])
	t.Log(weight)

	records := make([]string, 0)
	for i := 0; i < 5; i++ {
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
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

	nodeProvider := NewMockINodeProvider(ctrl)

	// init selector
	sel, _ := NewSelector()

	nodeProvider.EXPECT().AddHook(gomock.Any())
	nodeProvider.EXPECT().GetHosts().Return([]string{"a", "b", "c"})
	sel.SetNodeProvider(nodeProvider)

	// test setPriority blockWeight
	sel.SetWeight("a", BlockWeight)
	sel.SetWeight("c", BlockWeight)
	sel.SetWeight("b", DefaultWeight)
	weights := sel.ListWeight()
	assert.Equal(t, BlockWeight, weights["a"])
	assert.Equal(t, BlockWeight, weights["c"])
	assert.Equal(t, DefaultWeight, weights["b"])
	records := make([]string, 0)
	for i := 0; i < 3; i++ {
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
		assert.NoError(t, err)
	}
	dict := countSlice(records)
	assert.Equal(t, 0, dict["a"])
	assert.Equal(t, 0, dict["c"])
	assert.Equal(t, 3, dict["b"])

	sel.SetWeight("a", BlockWeight)
	sel.SetWeight("c", BlockWeight)
	sel.SetWeight("b", BlockWeight)
	_, err := sel.Select()
	assert.Error(t, err)
	assert.Equal(t, ErrNoNodeAvailable, err)
}

func Test_Selector_SetPriority(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeProvider := NewMockINodeProvider(ctrl)

	sel, _ := NewSelector()
	nodeProvider.EXPECT().AddHook(gomock.Any())
	nodeProvider.EXPECT().GetHosts().Return([]string{"a", "b", "c"})
	sel.SetNodeProvider(nodeProvider)

	// a 2 b 1 c 1
	sel.setPriority(CatchUpPriority, "a")
	records := make([]string, 0)
	for i := 0; i < 4; i++ {
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
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
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
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
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
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
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
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
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
		assert.NoError(t, err)
	}
	dict = countSlice(records)
	assert.Equal(t, 3, dict["a"])
}

func Test_Selector_SetPriority_SetWeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeProvider := NewMockINodeProvider(ctrl)

	sel, _ := NewSelector()
	nodeProvider.EXPECT().AddHook(gomock.Any())
	nodeProvider.EXPECT().GetHosts().Return([]string{"a", "b", "c"})
	sel.SetNodeProvider(nodeProvider)

	// a 2 b 1 c 0
	sel.setPriority(CatchUpPriority, "a")
	sel.setPriority(DelayPriority, "b")
	sel.setPriority(ErrPriority, "c")
	records := make([]string, 0)
	for i := 0; i < 4; i++ {
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
		assert.NoError(t, err)
	}
	dict := countSlice(records)
	assert.Equal(t, 4, dict["a"])

	// wa 0
	sel.SetWeight("a", BlockWeight)
	records = make([]string, 0)
	for i := 0; i < 4; i++ {
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
		assert.NoError(t, err)
	}
	dict = countSlice(records)
	assert.Equal(t, 4, dict["b"])

	// wa 0 wb 0
	sel.SetWeight("a", BlockWeight)
	sel.SetWeight("b", BlockWeight)
	records = make([]string, 0)
	for i := 0; i < 4; i++ {
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
		assert.NoError(t, err)
	}
	dict = countSlice(records)
	assert.Equal(t, 4, dict["c"])

	// wa 0 wb 0 wc 0
	sel.SetWeight("a", BlockWeight)
	sel.SetWeight("b", BlockWeight)
	sel.SetWeight("c", BlockWeight)
	_, err := sel.Select()
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
