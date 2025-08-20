package commands

/**
 *    Veredarii, software for interoperability.
 *    This file is part of Veredarii.
 *
 *    @author jcDaille
 *
 *
 *    MIT License
 *
 * Copyright (c) 2025 JC Daille
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 */

import (
	"Veredarii/components"
	"Veredarii/general"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the execution of Interop",
	Long:  `...`,
	Run: func(cmd *cobra.Command, args []string) {
		Start()
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}

func Start() {
	execution := true
	fmt.Println("Starting Interop...")
	general.StartingChannels()

	log.SetFormatter(&prefixed.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceFormatting: true,
		DisableColors:   true,
	})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)

	InteropController := components.NewInteropController()
	InteropController.Start()
	signal.Notify(InteropController.ChSigs, syscall.SIGINT, syscall.SIGTERM)
	/*
		go func() {
			time.Sleep(10 * time.Second)
			log.Debug(util.Red("channel -> terminar cosas"))
			general.Chan.StopLocalServer <- true
			general.Chan.StopNetwork <- true
		}()

		go func() {
			time.Sleep(20 * time.Second)
			log.Debug(util.Red("channel -> reiniciar cosas"))
			general.Chan.StartLocalServer <- true
			general.Chan.StartNetwork <- true
		}()*/

	for execution {
		select {
		case <-InteropController.ChInit:
		case <-InteropController.ChSigs:
			log.Info("Exit")
			execution = false
		}
	}
}
