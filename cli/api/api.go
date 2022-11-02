package api

import "context"

type LocalAPI interface {
	SetWeight(ctx context.Context, addr string, weight int) error //perm:admin
	ListWeight(ctx context.Context) (map[string]int, error)       //perm:read
	ListPriority(ctx context.Context) (map[string]int, error)     //perm:read
}
