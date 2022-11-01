package co

import (
	"testing"

	"github.com/stretchr/testify/assert"
)


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
