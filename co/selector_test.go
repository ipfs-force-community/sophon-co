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

	weight := sel.ListPriority()
	assert.Equal(t, 4, len(weight))
	assert.Equal(t, DefaultWeight, weight["a"])
	assert.Equal(t, DefaultWeight, weight["b"])
	weight_d := sel.getPriority("d")
	assert.Equal(t, DefaultWeight, weight_d)
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
	assert.Equal(t, 3, len(sel.ListPriority()))
	assert.Equal(t, DefaultWeight, sel.ListPriority()["a"])

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

func Test_Selector_SetPriority(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeProvider := NewMockINodeProvider(ctrl)

	// init selector
	sel, _ := NewSelector()

	nodeProvider.EXPECT().AddHook(gomock.Any())
	nodeProvider.EXPECT().GetHosts().Return([]string{"a", "b", "c"})
	sel.SetNodeProvider(nodeProvider)

	// test setPriority
	sel.setPriority("a", 4)
	pa := sel.getPriority("a")
	assert.Equal(t, 4, pa)

	records := make([]string, 0)
	for i := 0; i < 4; i++ {
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
		assert.NoError(t, err)
	}
	dict := countSlice(records)
	assert.Equal(t, 2, dict["a"])
	assert.Equal(t, 1, dict["c"])
	assert.Equal(t, 1, dict["b"])
}

func Test_Selector_SetPriority_ErrWeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeProvider := NewMockINodeProvider(ctrl)

	// init selector
	sel, _ := NewSelector()

	nodeProvider.EXPECT().AddHook(gomock.Any())
	nodeProvider.EXPECT().GetHosts().Return([]string{"a", "b", "c"})
	sel.SetNodeProvider(nodeProvider)

	// test setPriority errweight
	sel.setPriority("a", ErrWeight)
	sel.setPriority("c", ErrWeight)
	records := make([]string, 0)
	for i := 0; i < 4; i++ {
		nodeProvider.EXPECT().GetNode(gomock.Any()).Do(func(k string) {
			records = append(records, k)
		}).Return(nil)
		_, err := sel.Select()
		assert.NoError(t, err)
	}
	dict := countSlice(records)
	assert.Equal(t, 4, dict["b"])

	sel, _ = NewSelector()
	nodeProvider.EXPECT().AddHook(gomock.Any())
	nodeProvider.EXPECT().GetHosts().Return([]string{"a", "d", "c"})
	sel.SetNodeProvider(nodeProvider)

	sel.setPriority("a", ErrWeight)
	sel.setPriority("c", ErrWeight)
	sel.setPriority("b", ErrWeight)
	for i := 0; i < 3; i++ {
		nodeProvider.EXPECT().GetNode(gomock.Any()).Return(nil)
		_, err := sel.Select()
		assert.NoError(t, err)
	}
}

func Test_Selector_SetPriority_BlockWeight(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	nodeProvider := NewMockINodeProvider(ctrl)

	// init selector
	sel, _ := NewSelector()

	nodeProvider.EXPECT().AddHook(gomock.Any())
	nodeProvider.EXPECT().GetHosts().Return([]string{"a", "b", "c"})
	sel.SetNodeProvider(nodeProvider)

	// test setPriority blockWeight
	sel.setPriority("a", BlockWeight)
	sel.setPriority("c", BlockWeight)
	sel.setPriority("b", ErrWeight)
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

	sel.setPriority("a", BlockWeight)
	sel.setPriority("c", BlockWeight)
	sel.setPriority("b", BlockWeight)
	_, err := sel.Select()
	assert.Error(t, err)
	assert.Equal(t, ErrNoNodeAvailable, err)
}

func Test_SWRRA(t *testing.T) {
	cases := []map[string]int{
		{"a": 1, "b": 2},
		{"a": 1, "b": 2, "c": 3},
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
