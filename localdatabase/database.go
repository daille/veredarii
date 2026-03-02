package localdatabase

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
	"fmt"

	"github.com/cockroachdb/pebble"
)

type Database struct {
	Storage *pebble.DB
}

var DB *Database

func NewDatabase(path string) (*Database, error) {
	db, err := pebble.Open(path, &pebble.Options{})
	if err != nil {
		return nil, fmt.Errorf("error abriendo pebble: %w", err)
	}

	DB = &Database{Storage: db}
	return DB, nil
}

func (d *Database) Set(key string, value string) error {
	return d.Storage.Set([]byte(key), []byte(value), pebble.Sync)
}

func (d *Database) Get(key string) (string, error) {
	val, closer, err := d.Storage.Get([]byte(key))
	if err != nil {
		return "", err
	}
	defer closer.Close()
	return string(val), nil
}

func (d *Database) Close() error {
	return d.Storage.Close()
}
