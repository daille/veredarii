package pluginmanager

/*
MIT License

Copyright (c) 2026 Juan Carlos Daille

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
import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

const (
	MinInstances = 2                // Mínimo de instancias siempre vivas
	MaxInstances = 50               // Máximo de instancias permitidas por carga
	IdleTimeout  = 1 * time.Minute  // Tiempo antes de cerrar una instancia ociosa
	CleanupTick  = 30 * time.Second // Frecuencia de revisión de Scale Down
)

type WasmInstance struct {
	Module   api.Module
	LastUsed time.Time
}

type ElasticPool struct {
	name         string
	compiled     wazero.CompiledModule
	instances    chan *WasmInstance
	currentCount int32
	mu           sync.Mutex
	runtime      wazero.Runtime
}

type PluginManager struct {
	runtime wazero.Runtime
	ctx     context.Context
	pools   map[string]*ElasticPool
	mu      sync.RWMutex
}

func (p *ElasticPool) GetInstance(ctx context.Context) (*WasmInstance, error) {
	select {
	case inst := <-p.instances:
		inst.LastUsed = time.Now()
		return inst, nil
	default:
		// SCALE UP
		if atomic.LoadInt32(&p.currentCount) < MaxInstances {
			p.mu.Lock()
			if atomic.LoadInt32(&p.currentCount) < MaxInstances {
				newID := atomic.AddInt32(&p.currentCount, 1)
				conf := wazero.NewModuleConfig().
					WithName(fmt.Sprintf("%s-%d-%d", p.name, newID, time.Now().UnixNano())).
					WithStdout(os.Stdout).WithStderr(os.Stderr)

				mod, err := p.runtime.InstantiateModule(ctx, p.compiled, conf)
				p.mu.Unlock()
				if err != nil {
					atomic.AddInt32(&p.currentCount, -1)
					return nil, err
				}
				log.Printf("📈 Scale Up [%s]: Instancia #%d creada", p.name, newID)
				return &WasmInstance{Module: mod, LastUsed: time.Now()}, nil
			}
			p.mu.Unlock()
		}
		select {
		case inst := <-p.instances:
			inst.LastUsed = time.Now()
			return inst, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func (pm *PluginManager) instantiateHostModule(ctx context.Context) {
	builder := pm.runtime.NewHostModuleBuilder("env")

	// función log
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module, ptr, size uint32) {
			buf, _ := m.Memory().Read(ptr, size)
			fmt.Printf("📢 LOG DESDE WASM: %s\n", string(buf))
		}).
		Export("log_huesped")

	// función transformar
	builder.NewFunctionBuilder().
		WithFunc(func(ctx context.Context, m api.Module, ptr, size uint32) uint64 {

			buf, _ := m.Memory().Read(ptr, size)
			textoModificado := strings.ToUpper(string(buf))

			m.Memory().Write(ptr, []byte(textoModificado))

			return uint64(len(textoModificado))
		}).
		Export("transformar_texto")

	_, err := builder.Instantiate(ctx)

	if err != nil {
		log.Fatalf("Error instanciando Host Module: %v", err)
	}
}

func (p *ElasticPool) ReleaseInstance(inst *WasmInstance) {
	p.instances <- inst
}

// SCALE DOWN
func (p *ElasticPool) Cleanup(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	numInPool := len(p.instances)
	for i := 0; i < numInPool; i++ {
		if atomic.LoadInt32(&p.currentCount) <= MinInstances {
			break
		}

		select {
		case inst := <-p.instances:
			if time.Since(inst.LastUsed) > IdleTimeout {
				inst.Module.Close(ctx)
				atomic.AddInt32(&p.currentCount, -1)
				log.Printf("📉 Scale Down [%s]: Instancia ociosa cerrada. Quedan: %d", p.name, atomic.LoadInt32(&p.currentCount))
			} else {
				p.instances <- inst
			}
		default:
			return
		}
	}
}

func NewPluginManager(ctx context.Context) (*PluginManager, error) {
	r := wazero.NewRuntime(ctx)
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	pm := &PluginManager{
		runtime: r,
		ctx:     ctx,
		pools:   make(map[string]*ElasticPool),
	}

	pm.instantiateHostModule(ctx)

	go pm.startCleanupLoop()
	return pm, nil
}

func (pm *PluginManager) startCleanupLoop() {
	ticker := time.NewTicker(CleanupTick)
	for {
		select {
		case <-ticker.C:
			pm.mu.RLock()
			for _, pool := range pm.pools {
				pool.Cleanup(pm.ctx)
			}
			pm.mu.RUnlock()
		case <-pm.ctx.Done():
			return
		}
	}
}

func (pm *PluginManager) loadPlugin(path string) {
	name := strings.TrimSuffix(filepath.Base(path), ".wasm")
	wasmBytes, _ := os.ReadFile(path)
	compiled, err := pm.runtime.CompileModule(pm.ctx, wasmBytes)
	if err != nil {
		log.Printf("❌ Error compilando %s: %v", name, err)
		return
	}

	pool := &ElasticPool{
		name:      name,
		compiled:  compiled,
		instances: make(chan *WasmInstance, MaxInstances),
		runtime:   pm.runtime,
	}

	for i := 0; i < MinInstances; i++ {
		atomic.AddInt32(&pool.currentCount, 1)
		conf := wazero.NewModuleConfig().WithName(fmt.Sprintf("%s-init-%d", name, i))
		mod, _ := pm.runtime.InstantiateModule(pm.ctx, compiled, conf)
		pool.instances <- &WasmInstance{Module: mod, LastUsed: time.Now()}
	}

	pm.mu.Lock()
	pm.pools[name] = pool
	pm.mu.Unlock()
	log.Printf("✅ Pool elástico listo para: %s (Min: %d, Max: %d)", name, MinInstances, MaxInstances)
}

func (pm *PluginManager) StartWatcher(dir string) {
	watcher, _ := fsnotify.NewWatcher()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if strings.HasSuffix(event.Name, ".wasm") && (event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create) {
					pm.loadPlugin(event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("Watcher error:", err)
			}
		}
	}()
	watcher.Add(dir)
	files, _ := filepath.Glob(filepath.Join(dir, "*.wasm"))
	for _, f := range files {
		pm.loadPlugin(f)
	}
}

func (pm *PluginManager) Execute(ctx context.Context, pluginName string, req *WasmRequest) (*WasmRequest, error) {
	pm.mu.RLock()
	pool, ok := pm.pools[pluginName]
	pm.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("plugin no encontrado")
	}

	inst, err := pool.GetInstance(ctx)
	if err != nil {
		return nil, err
	}
	defer pool.ReleaseInstance(inst)
	return CallWasmModule(ctx, inst.Module, req)
}

func CallWasmModule(ctx context.Context, mod api.Module, req *WasmRequest) (*WasmRequest, error) {
	data, _ := req.MarshalVT()
	alloc := mod.ExportedFunction("allocate")
	handle := mod.ExportedFunction("handle")

	res, err := alloc.Call(ctx, uint64(len(data)))
	if err != nil {
		return nil, err
	}
	ptr := uint32(res[0])

	mod.Memory().Write(ptr, data)
	out, err := handle.Call(ctx, uint64(ptr), uint64(len(data)))
	if err != nil {
		return nil, err
	}
	packed := out[0]
	rPtr, rSize := uint32(packed>>32), uint32(packed)
	resBytes, _ := mod.Memory().Read(rPtr, rSize)

	resp := &WasmRequest{}
	resp.UnmarshalVT(resBytes)
	return resp, nil
}
