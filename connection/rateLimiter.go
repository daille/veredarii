package connection

import (
	"sync"

	"golang.org/x/time/rate"
)

/*
MIT License

# Copyright (c) 2026 Juan Carlos Daille

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
type RateManager struct {
	entidadLimits map[string]map[string]*rate.Limiter
	mu            sync.Mutex
}

func (rm *RateManager) GetLimiter(entidad string, servicio string, tps int64) *rate.Limiter {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.entidadLimits[entidad] == nil {
		rm.entidadLimits[entidad] = make(map[string]*rate.Limiter)
	}

	l, exists := rm.entidadLimits[entidad][servicio]
	if !exists {
		l = rate.NewLimiter(rate.Limit(tps), int(tps)+1)
		rm.entidadLimits[entidad][servicio] = l
	}
	return l
}
