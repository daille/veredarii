package util

/*
MIT License

Copyright (c) 2025 Juan Carlos Daille

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
	"archive/zip"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
)

func CreateZip(name string) (*zip.Writer, bool) {
	archive, err := os.Create("./" + name)
	if err != nil {
		log.Error(err)
		return nil, false
	}
	return zip.NewWriter(archive), true
}

func WriteZipStringFile(zipWriter *zip.Writer, name string, content string) {
	w1, err := zipWriter.Create(name)
	if err != nil {
		log.Error(err)
		return
	}
	w1.Write([]byte(content))
}

func WriteZipFile(zipWriter *zip.Writer, name string) {
	f2, err := os.Open(name)
	if err != nil {
		log.Error(err)
		return
	}
	defer f2.Close()

	w2, err := zipWriter.Create(name)
	if err != nil {
		log.Error(err)
		return
	}
	if _, err := io.Copy(w2, f2); err != nil {
		log.Error(err)
		return
	}
}
