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
	"Veredarii/global"
	"context"
	"fmt"
	"strings"

	"github.com/ipfs/go-datastore/query"
	dspebble "github.com/ipfs/go-ds-pebble"
	"github.com/spf13/cobra"
)

var utilCmd = &cobra.Command{
	Use:   "util",
	Short: "Operaciones de util",
}

var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Operaciones de base de datos",
	Run: func(cmd *cobra.Command, args []string) {
		inspectLocalDB2("store/network_veredarii_test.db")
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Operaciones de test",
	Run: func(cmd *cobra.Command, args []string) {
		holaSocket()
	},
}

func init() {
	utilCmd.AddCommand(databaseCmd)
	utilCmd.AddCommand(testCmd)
	rootCmd.AddCommand(utilCmd)
}

func GetNetwork() (string, bool) {
	respuesta := socketClient(Mensaje{Entrada: []string{"get"}})
	if respuesta.Salida != "" {
		return respuesta.Salida, true
	} else if network != "" {
		return network, true
	} else {
		return "", false
	}
}

func holaSocket() {
	msg := Mensaje{
		Entrada: []string{"hola", "veredarii", "test"},
	}
	respuesta := socketClient(msg)
	fmt.Println(respuesta)
}

func inspectLocalDB2(dbPath string) error {
	bucket := "peers"
	ds, err := dspebble.NewDatastore(dbPath, nil)
	if err != nil {
		return fmt.Errorf("error abriendo db: %w", err)
	}
	defer ds.Close()

	results, err := ds.Query(context.Background(), query.Query{
		Prefix: "/crdt-data/s/k/" + bucket,
	})
	if err != nil {
		return fmt.Errorf("error en query: %w", err)
	}
	defer results.Close()

	count := 0
	for r := range results.Next() {
		if r.Error != nil {
			fmt.Printf("Error: %v\n", r.Error)
			continue
		}
		if !strings.HasSuffix(r.Key, "/v") {
			continue
		}
		cleanKey := strings.TrimPrefix(r.Key, "/crdt-data/s/k/"+bucket)
		cleanKey = strings.TrimSuffix(cleanKey, "/v")
		fmt.Printf("[%d] KEY: %s | VALUE: %s\n", count, cleanKey, string(r.Entry.Value))
		count++
	}
	resp := []global.KVType{}

	results, err = ds.Query(context.Background(), query.Query{
		Prefix: "/crdt-data/s/k/",
	})
	if err != nil {
		fmt.Println("error en query: %w", err)
	}
	defer results.Close()

	for r := range results.Next() {
		if r.Error != nil {
			fmt.Println("error en query: %w", r.Error)
			continue
		}
		if !strings.HasSuffix(r.Key, "/v") {
			continue
		}
		cleanKey := strings.TrimPrefix(r.Key, "/crdt-data/s/k/")
		cleanKey = strings.TrimSuffix(cleanKey, "/v")
		resp = append(resp, global.KVType{Key: cleanKey, Name: string(r.Entry.Value)})
	}
	fmt.Println(resp)

	return nil
}
