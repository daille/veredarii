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
	"Veredarii/localdatabase"
	"Veredarii/localinterface"
	pluginmanager "Veredarii/pluginManager"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
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

const Version = "0.1.0"

func init() {
	rootCmd.AddCommand(startCmd)
}

func IniciaVeredarii() {

	fmt.Println("\n\n╭────────────────────────────────────────────────────────────────────────╮")
	fmt.Printf("│%s%-29s│\n", "                                Veredarii  ", "")
	fmt.Println("│                                                                        │")
	fmt.Printf("│ Versión: %-62s│\n", Version)
	fmt.Print("╰────────────────────────────────────────────────────────────────────────╯\n\n")

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(handler)
	slog.SetDefault(logger)
	slog := slog.With(
		slog.String("comando", "start"),
	)

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
	}
	connection.NM = connection.NewNetworkManager()
	for _, network := range configuration.CM.GetConfig().Networks {
		slog.Info("Agregando red", "red", network.Name)
		connection.NM.AddNetwork(network)
	}
	go connection.NM.StartProcess()

	ctx := context.Background()
	plugManager, _ := pluginmanager.NewPluginManager(ctx)
	plugManager.StartWatcher("./plugins")

	connection.NM.ChannelNetworks <- "init"

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	time.Sleep(3 * time.Second)
	go localinterface.Start()

	fmt.Println("Veredarii en ejecución. Presiona Ctrl+C para detener.")
	slog.Info("Veredarii en ejecución.")

	<-sigCh
}
