package pluginmanager

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"sync"

	extism "github.com/extism/go-sdk"
)

var PM *PluginManager

type PluginMessageType struct {
	Action     string
	Parameters map[string]string
	Status     string
	Payload    []byte
	Error      string
}

type pluginPool struct {
	compiled *extism.CompiledPlugin
	pool     chan *extism.Plugin
}

type PluginManager struct {
	ctx           context.Context
	pools         map[string]*pluginPool
	mu            sync.RWMutex
	hostFunctions []extism.HostFunction
}

func NewPluginManager(ctx context.Context, hostFunctions []extism.HostFunction) *PluginManager {
	return &PluginManager{
		ctx:           ctx,
		pools:         make(map[string]*pluginPool),
		hostFunctions: hostFunctions,
	}
}

func (pm *PluginManager) LoadPlugin(path string, poolSize int) {
	name := strings.TrimSuffix(filepath.Base(path), ".wasm")

	manifest := extism.Manifest{
		Wasm: []extism.Wasm{extism.WasmFile{Path: path}},
	}
	config := extism.PluginConfig{EnableWasi: true}

	compiled, err := extism.NewCompiledPlugin(pm.ctx, manifest, config, pm.hostFunctions)
	if err != nil {
		slog.Error("Error compilando plugin", "name", name, "error", err)
		return
	}

	ch := make(chan *extism.Plugin, poolSize)
	for i := 0; i < poolSize; i++ {
		inst, err := compiled.Instance(pm.ctx, extism.PluginInstanceConfig{})
		if err != nil {
			slog.Error("Error creando instancia", "index", i, "name", name, "error", err)
			continue
		}

		if i == 0 {
			_, out, err := inst.Call("metadata", nil)
			if err == nil {
				var meta map[string]any
				json.Unmarshal(out, &meta)
			}
		}
		ch <- inst
	}

	pm.mu.Lock()
	pm.pools[name] = &pluginPool{compiled: compiled, pool: ch}
	pm.mu.Unlock()

	slog.Info("Plugin cargado", "name", name, "instances", len(ch))
}

func (pm *PluginManager) Execute(ctx context.Context, pluginName string, data []byte) ([]byte, error) {
	pm.mu.RLock()
	p, ok := pm.pools[pluginName]
	pm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("plugin '%s' no encontrado", pluginName)
	}

	inst := <-p.pool
	defer func() { p.pool <- inst }()

	_, out, err := inst.Call("handle_request", data)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando plugin '%s': %w", pluginName, err)
	}

	return out, nil
}

func (pm *PluginManager) Close() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for name, p := range pm.pools {
		close(p.pool)
		for inst := range p.pool {
			inst.Close(pm.ctx)
		}
		p.compiled.Close(pm.ctx)
		slog.Info("Plugin cerrado", "name", name)
	}
}
