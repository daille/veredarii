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
	"Veredarii/configuration"
	"Veredarii/global"
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func (n *Network) handleFileFetch(s network.Stream) {
	defer s.Close()

	reader := bufio.NewReader(s)
	filePath, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	filePath = strings.TrimSpace(filePath)

	for _, resource := range n.Resources.FILE {
		if resource.Name == filePath {
			file, err := os.Open(resource.ResourcePath)
			if err != nil {
				s.Write([]byte("ERR: " + err.Error() + "\n"))
				return
			}
			defer file.Close()

			s.Write([]byte("OK\n"))

			written, err := io.Copy(s, file)
			if err != nil {
				fmt.Printf("Error durante el envío: %v\n", err)
				return
			}

			fmt.Printf("📤 Archivo '%s' enviado correctamente (%d bytes)\n", filePath, written)
			return
		}
	}
	s.Write([]byte("-1\n"))

}

func (n *Network) handleFileStat(s network.Stream) {
	defer s.Close()

	reader := bufio.NewReader(s)
	filePath, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	filePath = strings.TrimSpace(filePath)

	for _, resource := range n.Resources.FILE {
		if resource.Name == filePath {
			fileInfo, err := os.Stat(resource.ResourcePath)
			if err != nil {
				s.Write([]byte("-1\n"))
				return
			}

			s.Write([]byte(fmt.Sprintf("%d\n", fileInfo.Size())))
			break
		}
	}
}

func (n *Network) GetRemoteStat(dest peer.ID, path string) (int64, error) {
	s, err := n.Host.NewStream(context.Background(), dest, global.ProtocolFileSystemStat)
	if err != nil {
		log.Error("Error al abrir stream: ", err)
		return 0, err
	}
	defer s.Close()

	s.Write([]byte(path + "\n"))

	reader := bufio.NewReader(s)
	line, _ := reader.ReadString('\n')
	size, _ := strconv.ParseInt(strings.TrimSpace(line), 10, 64)

	return size, nil
}

func (n *Network) RequestFile(dest peer.ID, remotePath string, localDest string) ([]byte, error) {
	log.Debug("Abrir stream con el protocolo")
	s, err := n.Host.NewStream(context.Background(), dest, global.ProtocolFileSystem)
	if err != nil {
		log.Error("Error al abrir stream: ", err)
		return nil, err
	}
	defer s.Close()

	s.Write([]byte(remotePath + "\n"))
	reader := bufio.NewReader(s)
	status, err := reader.ReadString('\n')
	if err != nil || !strings.HasPrefix(status, "OK") {
		log.Error("Error del servidor: ", status)
		return nil, fmt.Errorf("error del servidor: %s", status)
	}

	var buf bytes.Buffer

	fmt.Println("📥 Recibiendo datos...")
	nBytes, err := io.Copy(&buf, s)
	if err != nil {
		log.Error("Error al recibir el stream: ", err)
		return nil, err
	}

	fmt.Printf("✅ Descarga completada: %d bytes guardados en %s\n", nBytes, localDest)
	return buf.Bytes(), nil
}

func (n *Network) FileSystem() {
	log.Debug("init handleFileSystem")
	mountpoint := flag.String("mount", configuration.CM.GetConfig().Networks[0].FS, "punto de montaje")
	flag.Parse()

	c, err := fuse.Mount(*mountpoint,
		fuse.FSName("MiP2P_VFS"),
		fuse.Subtype("vfs"),
	)

	if err != nil {
		log.Error(err)
		return
	}
	defer c.Close()

	err = fs.Serve(c, &FS{N: n})
	if err != nil {
		log.Error(err)
	}
}

type FS struct {
	N *Network
}

func (f *FS) Root() (fs.Node, error) {
	return &Dir{N: f.N}, nil
}

type Dir struct {
	N *Network
}

func (d *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = os.ModeDir | 0555 // Solo lectura para este ejemplo
	return nil
}

func (d *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	f := []fuse.Dirent{}
	for _, resource := range d.N.RemoteResources.FILE {
		f = append(f, fuse.Dirent{Name: resource.Name, Type: fuse.DT_File})
	}

	return f, nil
}

func (d *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	for _, resource := range d.N.RemoteResources.FILE {
		if name == resource.Name {

			targetID := d.N.BuscarServicio(context.Background(), resource.Name)
			if targetID == "" {
				log.Error("Servicio no encontrado")
				return nil, fuse.ENOENT
			}

			log.Debug("Buscando Stat del archivo: ", name)
			remoteSize, err := d.N.GetRemoteStat(targetID, name)
			if err != nil {
				log.Error("Error al obtener el stat del archivo: ", err)
				return nil, fuse.ENOENT
			}
			log.Debug("Stat del archivo: ", remoteSize)
			return &File{N: d.N, FileName: name, Size: uint64(remoteSize)}, nil
		}
	}
	return nil, fuse.ENOENT
}

type File struct {
	N        *Network
	FileName string
	Content  []byte
	Size     uint64
}

func (f *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	resp.Flags |= fuse.OpenDirectIO
	return f, nil
}

func (f *File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Mode = 0444
	a.Size = f.Size
	return nil
}

func (f *File) ReadAll(ctx context.Context) ([]byte, error) {
	log.Debug("Leyendo archivo...")
	fmt.Printf("🚀 Iniciando descarga P2P para: %s...\n", f.FileName)

	targetID := f.N.BuscarServicio(context.Background(), f.FileName)
	if targetID == "" {
		log.Error("Servicio no encontrado")
		return nil, fuse.ENOENT
	}
	content, err := f.N.RequestFile(targetID, f.FileName, f.FileName)
	if err != nil {
		return nil, fuse.ENOENT
	}

	return content, nil
}
