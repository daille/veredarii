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
	"Veredarii/localinterface"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	log "github.com/sirupsen/logrus"
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

	log.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	log.Debug("Cargando configuracion...")
	configuration.CM = configuration.NewConfigurationManager()
	err := configuration.CM.LoadConfig()
	if err != nil {
		log.Fatal("Error cargando configuracion:", err)
	}
	connection.NM = connection.NewNetworkManager()
	for _, network := range configuration.CM.GetConfig().Networks {
		log.Debug("Agregando red: ", network.Name)
		connection.NM.AddNetwork(network)
	}
	go connection.NM.StartProcess()
	connection.NM.ChannelNetworks <- "init"

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	time.Sleep(3 * time.Second)
	go localinterface.Start()

	fmt.Println("Nodo corriendo. Presiona Ctrl+C para detener.")

	<-sigCh
}
