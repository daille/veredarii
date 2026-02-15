package connection

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
	global "Veredarii/global"
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	_ "github.com/marcboeker/go-duckdb"
	log "github.com/sirupsen/logrus"
	"github.com/xitongsys/parquet-go/source"
	"google.golang.org/protobuf/proto"
)

type QueryType struct {
	Query     string `json:"query"`
	Format    string `json:"format"`
	FileName  string `json:"file_name"`
	BlockSize int    `json:"block_size"`
}

func StringToQueryType(jsonStr string) (*QueryType, error) {
	var req QueryType

	fmt.Println("jsonStr -> StringToQueryType -> ", jsonStr)
	err := json.Unmarshal([]byte(jsonStr), &req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func QueryTypeToString(req *QueryType) (string, error) {
	bytes, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func (n *Network) HandleSearch(s network.Stream) {
	defer s.Close()

	log.Debug("HandleSearch")
	db, err := sql.Open("duckdb", "")
	if err != nil {
		log.Error("Error al abrir la base de datos: ", err)
		return
	}
	defer db.Close()

	msg := &global.Envelop{}
	data, err := readDelimited(s)
	if err != nil {
		if err != io.EOF {
			log.Printf("Error leyendo stream: %v", err)
		}
		return
	}
	if err := proto.Unmarshal(data, msg); err != nil {
		log.Printf("Error unmarshal protobuf: %v", err)
		return
	}
	log.Debug("msg: ", msg)

	queryType, err := StringToQueryType(string(msg.Payload))
	if err != nil {
		log.Error("Error al convertir el payload a QueryType: ", err)
		return
	}

	for _, ds := range n.Resources.DATASOURCE {
		if ds.Name == msg.Service {
			queryType.FileName = ds.ResourcePath

			query := strings.ReplaceAll(queryType.Query, "{{ORIGIN}}", ds.ResourcePath)
			log.Debug("Query: ", query)
			if queryType.Format == "parquet" {
				file, err := os.CreateTemp("", "parquet-export-*.parquet")
				if err != nil {
					log.Errorf("Error creando archivo temporal: %v", err)
					return
				}
				fileName := file.Name()
				file.Close()
				defer os.Remove(fileName)

				fmt.Println("query: ", query)
				fmt.Println("format: ", queryType.Format)
				fmt.Println("fileName: ", fileName)
				exportToParquet(db, query, fileName)
				TransferFile(s, fileName)
				return
			}

			rows, err := db.Query(query)
			if err != nil {
				log.Error("Error ejecutando consulta: %v", err)
				return
			}
			defer rows.Close()

			cols, _ := rows.Columns()

			var batch []map[string]interface{}
			var batchCsv [][]string

			if queryType.Format == "csv" {
				sendCsvBatch(s, [][]string{cols}, false)
			}

			columns := make([]interface{}, len(cols))
			columnPointers := make([]interface{}, len(cols))
			for i := range columns {
				columnPointers[i] = &columns[i]
			}

			count := 0
			for rows.Next() {
				count++
				if err := rows.Scan(columnPointers...); err != nil {
					log.Error("Error escaneando fila: %v", err)
					break
				}

				if queryType.Format == "json" || queryType.Format == "parquet" {
					m := make(map[string]interface{})
					for i, colName := range cols {
						val := columns[i]
						if b, ok := val.([]byte); ok {
							m[colName] = string(b)
						} else {
							m[colName] = val
						}
					}
					batch = append(batch, m)
				} else if queryType.Format == "csv" {
					fila := make([]string, len(cols))
					for i, val := range columns {
						if val == nil {
							fila[i] = ""
						} else {
							fila[i] = fmt.Sprint(val)
						}
					}
					batchCsv = append(batchCsv, fila)
				}

				if count >= queryType.BlockSize {
					if queryType.Format == "csv" {
						log.Debug("filas: ", batchCsv)
						sendCsvBatch(s, batchCsv, true)
						batchCsv = nil
					} else if queryType.Format == "json" {
						sendJsonBatch(s, batch, false)
						batch = nil
					} else {
						log.Error("Formato no soportado: ", queryType.Format)
					}
					count = 0
				}
			}

			if queryType.Format == "csv" {
				sendCsvBatch(s, batchCsv, true)
			} else if queryType.Format == "json" {
				sendJsonBatch(s, batch, true)
			}
			break
		}
	}
}

func exportToParquet(db *sql.DB, query string, outputFile string) error {
	exportQuery := fmt.Sprintf("COPY (%s) TO '%s' (FORMAT PARQUET)", query, outputFile)

	_, err := db.Exec(exportQuery)
	if err != nil {
		return fmt.Errorf("error exportando a parquet: %w", err)
	}

	return nil
}

func scanRowToMap(rows *sql.Rows, cols []string) map[string]interface{} {
	numCols := len(cols)
	if numCols == 0 {
		return nil
	}

	values := make([]interface{}, numCols)
	valuePtrs := make([]interface{}, numCols)
	for i := 0; i < numCols; i++ {
		valuePtrs[i] = &values[i]
	}

	err := rows.Scan(valuePtrs...)
	if err != nil {
		log.Errorf("Error en Scan: %v", err)
		return nil
	}

	rowMap := make(map[string]interface{})
	for i, colName := range cols {
		if i >= len(values) {
			break
		}

		val := values[i]
		if val == nil {
			rowMap[colName] = ""
		} else {
			switch v := val.(type) {
			case []byte:
				rowMap[colName] = string(v)
			default:
				rowMap[colName] = fmt.Sprintf("%v", v)
			}
		}
	}
	return rowMap
}

func sendJsonBatch(s network.Stream, batch []map[string]interface{}, isLast bool) {
	if len(batch) == 0 && !isLast {
		return
	}
	jsonData, err := json.Marshal(batch)
	if err != nil {
		log.Error("Error al codificar JSON: %v", err)
		return
	}
	resp := &global.Envelop{
		Payload: jsonData,
	}
	out, _ := proto.Marshal(resp)
	writeDelimited(s, out)
}

func sendCsvBatch(s network.Stream, batch [][]string, isLast bool) {
	if len(batch) == 0 && !isLast {
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.WriteAll(batch); err != nil {
		log.Error("Error escribiendo CSV: %v", err)
		return
	}
	writer.Flush()
	resp := &global.Envelop{
		Payload: buf.Bytes(),
	}
	out, _ := proto.Marshal(resp)
	writeDelimited(s, out)
}

func sendBatch(s network.Stream, data []string, final bool) {
	log.Debug("mandando batch ", len(data))
	resp := &global.Envelop{Payload: []byte(strings.Join(data, "\n"))}
	out, _ := proto.Marshal(resp)
	fmt.Println("out: ", string(out))
	writeDelimited(s, out)
}

func (n *Network) Query(targetID peer.ID, query QueryType, service string) []byte {

	s, err := n.Host.NewStream(context.Background(), targetID, global.ProtocolQuery)
	if err != nil {
		log.Printf("Error abriendo stream: %v", err)
		return nil
	}
	defer s.Close()

	qt, err := json.Marshal(query)
	if err != nil {
		log.Error("Error al codificar JSON: %v", err)
		return nil
	}

	msg := &global.Envelop{
		Id:      uuid.New().String(),
		Service: service,
		Payload: qt,
	}
	data, _ := proto.Marshal(msg)
	writeDelimited(s, data)

	f, err := os.Create(query.FileName)
	if err != nil {
		log.Error("Error creando archivo: ", err)
		return nil
	}
	defer f.Close()

	br := bufio.NewReader(s)
	for {
		protoData, err := readDelimited(br)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Error("Error leyendo proto delimitado: ", err)
			return nil
		}

		if len(protoData) == 0 {
			log.Debug("Fin de stream recibido (proto vacío)")
			break
		}

		var msg global.Envelop
		if err := proto.Unmarshal(protoData, &msg); err != nil {
			log.Error("Error deserializando proto: ", err)
			return nil
		}

		if msg.Payload == nil || len(msg.Payload) == 0 {
			log.Debug("Fin de stream (payload vacío)")
			break
		}

		if _, err := f.Write(msg.Payload); err != nil {
			log.Error("Error escribiendo en disco: ", err)
			return nil
		}

		log.Debugf("Batch de %d bytes guardado", len(msg.Payload))
	}

	log.Info("Transferencia completada con éxito")
	return nil
}

func TransferFile(s network.Stream, fileName string) {

	fileToStream, err := os.Open(fileName)
	if err != nil {
		log.Errorf("Error abriendo para stream: %v", err)
		return
	}
	info, _ := os.Stat(fileName)
	log.Debugf("Tamaño real del archivo en disco: %d bytes", info.Size())

	buffer := make([]byte, 64*1024)

	log.Debug("Iniciando transmisión de bytes...")
	totalEnviado := 0
	for {
		n, err := fileToStream.Read(buffer)
		if n > 0 {
			resp := &global.Envelop{
				Payload: buffer[:n],
			}
			out, _ := proto.Marshal(resp)

			if _, err := writeDelimited(s, out); err != nil {
				log.Errorf("Error enviando chunk: %v", err)
				return
			}
			totalEnviado += n
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("Error leyendo archivo: %v", err)
			break
		}
	}
	fileToStream.Close()

	log.Info("Transmisión Parquet completada con éxito")
}

type MiMemFile struct {
	*bytes.Buffer
}

func (m *MiMemFile) Create(name string) (source.ParquetFile, error) { return m, nil }
func (m *MiMemFile) Open(name string) (source.ParquetFile, error)   { return m, nil }
func (m *MiMemFile) Close() error                                   { return nil }
func (m *MiMemFile) Seek(offset int64, whence int) (int64, error)   { return 0, nil }
func (m *MiMemFile) Read(p []byte) (n int, err error)               { return m.Buffer.Read(p) }
