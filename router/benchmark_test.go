package router

import (
	"github.com/DarthPestilane/easytcp/packet"
	"testing"
)

//go:generate go test -bench=^BenchmarkRouter$ -run=none -benchmem -memprofile=bench_profiles/BenchmarkRouter.mem.out -cpuprofile=bench_profiles/BenchmarkRouter.profile.out
func BenchmarkRouter(b *testing.B) {
	rt := NewRouter()
	rt.Register(1, nilHandler)
	msg := &packet.MessageEntry{ID: 1}
	for i := 0; i < b.N; i++ {
		if err := rt.handleReq(nil, msg); err != nil {
			panic(err)
		}
	}
}
