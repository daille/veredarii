package pluginmanager

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	extism "github.com/extism/go-sdk"
)

var PM *PluginManager

// pluginPool mantiene un CompiledPlugin + un channel de instancias listas
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

	// Compilar una sola vez
	compiled, err := extism.NewCompiledPlugin(pm.ctx, manifest, config, pm.hostFunctions)
	if err != nil {
		log.Printf("❌ Error compilando plugin %s: %v", name, err)
		return
	}

	// Pre-crear `poolSize` instancias en el channel
	ch := make(chan *extism.Plugin, poolSize)
	for i := 0; i < poolSize; i++ {
		inst, err := compiled.Instance(pm.ctx, extism.PluginInstanceConfig{})
		if err != nil {
			log.Printf("❌ Error creando instancia %d de %s: %v", i, name, err)
			continue
		}
		ch <- inst
	}

	pm.mu.Lock()
	pm.pools[name] = &pluginPool{compiled: compiled, pool: ch}
	pm.mu.Unlock()

	log.Printf("✅ Plugin '%s' cargado con %d instancias", name, len(ch))
}

func (pm *PluginManager) Execute(ctx context.Context, pluginName string, data []byte) ([]byte, error) {
	pm.mu.RLock()
	p, ok := pm.pools[pluginName]
	pm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("plugin '%s' no encontrado", pluginName)
	}

	// Obtener instancia del pool (bloquea si todas están ocupadas)
	inst := <-p.pool
	defer func() { p.pool <- inst }() // Devolver al pool siempre

	_, out, err := inst.Call("handle_request", data)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando plugin '%s': %w", pluginName, err)
	}

	return out, nil
}

// Close libera todos los recursos
func (pm *PluginManager) Close() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for name, p := range pm.pools {
		close(p.pool)
		for inst := range p.pool {
			inst.Close(pm.ctx)
		}
		p.compiled.Close(pm.ctx)
		log.Printf("🔒 Plugin '%s' cerrado", name)
	}
}
