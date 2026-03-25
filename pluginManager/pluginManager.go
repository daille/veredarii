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

// PM es la instancia global del PluginManager.
var PM *PluginManager

// PluginMessageType es el formato estándar de mensajes entre app y plugins.
type PluginMessageType struct {
	Action     string
	Parameters map[string]string
	Status     string
	Payload    []byte
	Error      string
}

// ExportDef describe una función exportada por un plugin que la app puede invocar.
type ExportDef struct {
	FuncName string // nombre exacto de la función en el .wasm
}

// HostFuncHandler es la firma de las funciones de la app que los plugins pueden llamar.
// input: bytes enviados por el plugin. output: bytes de respuesta.
type HostFuncHandler func(input []byte) ([]byte, error)

// pluginPool agrupa la instancia compilada y el pool de instancias de un plugin.
type pluginPool struct {
	compiled *extism.CompiledPlugin
	pool     chan *extism.Plugin
	exports  map[string]ExportDef // exports invocables declarados por la app
}

// PluginManager gestiona el ciclo de vida de los plugins WASM.
type PluginManager struct {
	ctx           context.Context
	pools         map[string]*pluginPool
	mu            sync.RWMutex
	hostFunctions []extism.HostFunction
}

// NewPluginManager crea un nuevo PluginManager con las host functions dadas.
// Las host functions son funciones de la app que los plugins pueden llamar.
func NewPluginManager(ctx context.Context, hostFunctions []extism.HostFunction) *PluginManager {
	return &PluginManager{
		ctx:           ctx,
		pools:         make(map[string]*pluginPool),
		hostFunctions: hostFunctions,
	}
}

// RegisterHostFunc crea un extism.HostFunction que expone handler al plugin.
// El plugin la invoca desde su namespace "extism:host/user" con el nombre dado.
//
// Protocolo de memoria: el plugin pasa (ptr, len) de su linear memory,
// la función escribe la respuesta en esa misma memoria y devuelve (ptr, len).
func RegisterHostFunc(name string, handler HostFuncHandler) extism.HostFunction {
	return extism.NewHostFunctionWithStack(
		name,
		func(ctx context.Context, p *extism.CurrentPlugin, stack []uint64) {
			// El i64 codifica ptr en los 32 bits altos y len en los 32 bajos
			packed := stack[0]
			offset := packed >> 32
			//length := packed & 0xFFFFFFFF

			input, err := p.ReadBytes(offset)
			if err != nil {
				errOut, _ := json.Marshal(map[string]string{"error": err.Error()})
				out, _ := p.WriteBytes(errOut)
				// Empaquetar ptr+len en un solo i64
				stack[0] = (out << 32) | uint64(len(errOut))
				return
			}

			output, err := handler(input)
			if err != nil {
				output, _ = json.Marshal(map[string]string{"error": err.Error()})
			}
			if output == nil {
				output = []byte("{}")
			}

			outOffset, _ := p.WriteBytes(output)
			stack[0] = (outOffset << 32) | uint64(len(output))
		},
		[]extism.ValueType{extism.ValueTypeI64}, // ← un solo i64
		[]extism.ValueType{extism.ValueTypeI64}, // ← un solo i64
	)
}

// LoadPlugin compila y crea un pool de instancias para el plugin en path.
// El nombre del plugin se deriva del nombre del archivo sin la extensión .wasm.
func (pm *PluginManager) LoadPlugin(path string, poolSize int) {
	name := strings.TrimSuffix(filepath.Base(path), ".wasm")

	absolutePath, _ := filepath.Abs("./dataspace")

	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmFile{Path: path},
		},
		AllowedPaths: map[string]string{
			absolutePath: "/data", // El plugin buscará en "/data"
		},
		Memory: &extism.ManifestMemory{
			MaxPages: 2048, // 128MB para estar sobrados
		},
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
			if _, out, err := inst.Call("metadata", nil); err == nil {
				var meta map[string]any
				if jsonErr := json.Unmarshal(out, &meta); jsonErr == nil {
					slog.Info("Plugin metadata", "name", name, "meta", meta)
				}
			}
		}

		ch <- inst
	}

	pm.mu.Lock()
	pm.pools[name] = &pluginPool{
		compiled: compiled,
		pool:     ch,
		exports:  make(map[string]ExportDef),
	}
	pm.mu.Unlock()

	slog.Info("Plugin cargado", "name", name, "instances", len(ch))
}

// Execute llama a "handle_request" en el plugin con el payload dado.
// Es el punto de entrada genérico para mensajes de la app hacia el plugin.
func (pm *PluginManager) Execute(ctx context.Context, pluginName string, method string, data []byte) ([]byte, error) {
	pm.mu.RLock()
	p, ok := pm.pools[pluginName]
	pm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("plugin '%s' no encontrado", pluginName)
	}

	inst := <-p.pool
	defer func() { p.pool <- inst }()

	e, out, err := inst.Call(method, data)
	if err != nil {
		return nil, fmt.Errorf("error ejecutando plugin %d '%s': %w", e, pluginName, err)
	}
	return out, nil
}

// RegisterExport declara qué funciones exportadas de un plugin son invocables por la app.
// Debe llamarse después de LoadPlugin. Puede llamarse múltiples veces; cada llamada
// agrega exports sin borrar los anteriores.
func (pm *PluginManager) RegisterExport(pluginName string, exports ...ExportDef) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pool, ok := pm.pools[pluginName]
	if !ok {
		return fmt.Errorf("plugin '%s' no cargado", pluginName)
	}
	for _, e := range exports {
		pool.exports[e.FuncName] = e
	}
	return nil
}

// CallExport invoca una función exportada de un plugin con input arbitrario.
// La función debe haber sido previamente registrada con RegisterExport.
// Utiliza el pool de instancias igual que Execute.
func (pm *PluginManager) CallExport(ctx context.Context, pluginName, funcName string, input []byte) ([]byte, error) {
	pm.mu.RLock()
	pool, ok := pm.pools[pluginName]
	pm.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("plugin '%s' no encontrado", pluginName)
	}
	if _, defined := pool.exports[funcName]; !defined {
		return nil, fmt.Errorf("export '%s' no registrado en plugin '%s'", funcName, pluginName)
	}

	inst := <-pool.pool
	defer func() { pool.pool <- inst }()

	_, out, err := inst.Call(funcName, input)
	if err != nil {
		return nil, fmt.Errorf("error llamando '%s.%s': %w", pluginName, funcName, err)
	}
	return out, nil
}

// CallExportJSON es un helper genérico tipado sobre CallExport.
// Serializa in como JSON, llama al export y deserializa la respuesta en Out.
//
// Ejemplo:
//
//	result, err := CallExportJSON[ScoreRequest, ScoreResult](
//	    ctx, pm, "analytics", "compute_score", ScoreRequest{UserID: "u-1"},
//	)
func CallExportJSON[In any, Out any](
	ctx context.Context,
	pm *PluginManager,
	pluginName, funcName string,
	in In,
) (Out, error) {
	var zero Out

	data, err := json.Marshal(in)
	if err != nil {
		return zero, fmt.Errorf("CallExportJSON marshal: %w", err)
	}

	raw, err := pm.CallExport(ctx, pluginName, funcName, data)
	if err != nil {
		return zero, err
	}

	var out Out
	if err := json.Unmarshal(raw, &out); err != nil {
		return zero, fmt.Errorf("CallExportJSON unmarshal: %w", err)
	}
	return out, nil
}

// Close libera todas las instancias y plugins compilados del manager.
// Debe llamarse al apagar la aplicación.
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
