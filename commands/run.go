package cmd

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
	"Veredarii/configuration"
	"Veredarii/connection"
	"Veredarii/global"
	"Veredarii/localdatabase"
	"Veredarii/localinterface"
	pluginmanager "Veredarii/pluginManager"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	extism "github.com/extism/go-sdk"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Ejecuta la aplicación Veredarii",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:
Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		IniciaVeredarii()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func IniciaVeredarii() {
	slog := slog.With(slog.String("comando", "run"))

	fmt.Println("\n\n╭────────────────────────────────────────────────────────────────────────╮")
	fmt.Printf("│%s%-29s│\n", "                                Veredarii  ", "")
	fmt.Println("│                                                                        │")
	fmt.Printf("│ Versión: %-62s│\n", global.Version)
	fmt.Print("╰────────────────────────────────────────────────────────────────────────╯\n\n")

	_, err := localdatabase.NewDatabase("./veredarii_data")
	if err != nil {
		slog.Error("Error al crear la base de datos", "error", err.Error())
	}
	defer localdatabase.DB.Close()

	slog.Debug("Cargando configuracion...")
	configuration.CM = configuration.NewConfigurationManager()
	err = configuration.CM.LoadConfig()
	if err != nil {
		slog.Error("Error cargando configuracion:", "error", err.Error())
		os.Exit(1)
		return
	}

	connection.NM = connection.NewNetworkManager()
	for _, network := range configuration.CM.GetConfig().Networks {
		slog.Info("Agregando red", "red", network.Name)
		connection.NM.AddNetwork(network)
	}
	go connection.NM.StartProcess()

	ctx := context.Background()

	// 1. Registrar funciones que los plugins pueden llamar
	dbLookup := pluginmanager.RegisterHostFunc("db_lookup", func(input []byte) ([]byte, error) {
		var req struct{ Key string }
		json.Unmarshal(input, &req)
		value := "valor"
		return json.Marshal(map[string]string{"value": value})
	})

	logEvent := pluginmanager.RegisterHostFunc("log_event", func(input []byte) ([]byte, error) {
		slog.Info("plugin event", "data", string(input))
		return nil, nil
	})
	pluginmanager.PM = pluginmanager.NewPluginManager(ctx, []extism.HostFunction{dbLookup, logEvent})
	defer pluginmanager.PM.Close()
	entries, err := os.ReadDir("./plugins")
	if err != nil {
		slog.Error("Error al leer el directorio de plugins", "error", err.Error())
	} else {
		for _, entry := range entries {
			if !entry.IsDir() && filepath.Ext(entry.Name()) == ".wasm" {
				pluginmanager.PM.LoadPlugin("plugins/"+entry.Name(), 5)
			}
		}
	}
	connection.NM.ChannelNetworks <- "init"

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	time.Sleep(3 * time.Second)
	go localinterface.Start()

	fmt.Println("Veredarii en ejecución. Presiona Ctrl+C para detener.")
	slog.Info("Veredarii en ejecución.")

	<-sigCh
}
