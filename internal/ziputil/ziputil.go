// package ziputil

// import (
// 	"archive/zip"
// 	"bufio"
// 	"bytes"
// 	"context"
// 	"fmt"
// 	"io"
// 	"os"
// 	"os/exec"
// 	"path"
// 	"path/filepath"
// 	"strconv"
// 	"strings"
// 	"syscall"
// 	"time"

// 	// "github.com/gen2brain/go-unarr"
// 	// "github.com/gen2brain/go-unarr"
// 	"github.com/jhillyerd/enmime"
// 	"github.com/mholt/archives"
// )

// func GetFileFromZipArchive(zipBytes []byte, filename string) (io.ReadCloser, error) {
// 	rdr := bytes.NewReader(zipBytes)
// 	r, err := zip.NewReader(rdr, int64(len(zipBytes)))
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, f := range r.File {
// 		if f.Name == filename {
// 			rc, err := f.Open()
// 			if err != nil {
// 				return nil, fmt.Errorf("nepavyko atidaryti failo %q: %w", filename, err)
// 			}
// 			return rc, nil
// 		}
// 	}
// 	return nil, fmt.Errorf("failas %q zip’e nerastas", filename)
// }

// func detectArchiveType(b []byte) string {
// 	if len(b) >= 4 && bytes.Equal(b[:4], []byte("PK\x03\x04")) {
// 		return "zip"
// 	}
// 	// RAR4
// 	if len(b) >= 7 && bytes.Equal(b[:7], []byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x00}) {
// 		return "rar"
// 	}
// 	// RAR5
// 	if len(b) >= 8 && bytes.Equal(b[:8], []byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x01, 0x00}) {
// 		return "rar"
// 	}
// 	if len(b) >= 6 && bytes.Equal(b[:6], []byte("7z\xBC\xAF\x27\x1C")) {
// 		return "7z"
// 	}
// 	return "unknown"
// }

// func IdentityFilesV2(ctx context.Context, archiveBytes []byte) ([]string, error) {
// 	switch detectArchiveType(archiveBytes) {

// 	case "zip":
// 		return listZip(archiveBytes)

// 	// case "7z":
// 	// 	return listWith7z(ctx, archiveBytes)

// 	default:
// 		return IdentityFilesV3(ctx, archiveBytes)
// 	}
// }

// func listZip(b []byte) ([]string, error) {
// 	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
// 	if err != nil {
// 		return nil, err
// 	}

// 	var out []string
// 	for _, f := range r.File {
// 		if !strings.HasSuffix(f.Name, "/") {
// 			out = append(out, f.Name)
// 		}
// 	}
// 	return out, nil
// }

// func GetFileFromArchiveV2(ctx context.Context, archiveBytes []byte, filename string) (io.ReadCloser, error) {
// 	switch detectArchiveType(archiveBytes) {

// 	case "zip":
// 		return getFromZip(archiveBytes, filename)

// 	// case "7z":
// 	// 	return getWith7z(ctx, archiveBytes, filename)

// 	default:
// 		return GetFileFromArchiveV3(ctx, archiveBytes, filename)
// 	}
// }

// func getFromZip(b []byte, filename string) (io.ReadCloser, error) {
// 	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
// 	if err != nil {
// 		return nil, err
// 	}

// 	for _, f := range r.File {
// 		if f.Name == filename {
// 			return f.Open()
// 		}
// 	}
// 	return nil, fmt.Errorf("failas nerastas: %s", filename)
// }

// func listWith7z(rctx context.Context, b []byte) ([]string, error) {
// 	tmp, err := os.CreateTemp("", "arc-*")
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer os.Remove(tmp.Name())
// 	os.WriteFile(tmp.Name(), b, 0644)
// 	ctx, cancel := context.WithTimeout(rctx, 15*time.Second)
// 	defer cancel()
// 	cmd := exec.CommandContext(ctx, "7z", "l", "-slt", tmp.Name())
// 	cmd.SysProcAttr = &syscall.SysProcAttr{
// 		Setpgid: true,
// 	}

// 	cmd.Cancel = func() error {
// 		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
// 	}
// 	out, err := cmd.Output()
// 	if err != nil {
// 		return nil, err
// 	}

// 	var files []string
// 	scanner := bufio.NewScanner(bytes.NewReader(out))
// 	for scanner.Scan() {
// 		line := strings.TrimSpace(scanner.Text())
// 		if strings.HasPrefix(line, "Path = ") {
// 			p := strings.TrimPrefix(line, "Path = ")
// 			if !strings.HasSuffix(p, "/") {
// 				files = append(files, p)
// 			}
// 		}
// 	}
// 	return files, scanner.Err()
// }
// func getWith7z(rctx context.Context, b []byte, filename string) (io.ReadCloser, error) {
// 	tmp, err := os.CreateTemp("", "arc-*")
// 	if err != nil {
// 		return nil, err
// 	}
// 	os.WriteFile(tmp.Name(), b, 0644)
// 	defer os.Remove(tmp.Name())
// 	ctx, cancel := context.WithTimeout(rctx, 15*time.Second)

// 	cmd := exec.CommandContext(
// 		ctx,
// 		"7z", "x",
// 		"-so",
// 		tmp.Name(),
// 		filename,
// 	)
// 	cmd.SysProcAttr = &syscall.SysProcAttr{
// 		Setpgid: true,
// 	}

// 	cmd.Cancel = func() error {
// 		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
// 	}

// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		cancel()
// 		return nil, err
// 	}
// 	stderr, err := cmd.StderrPipe()
// 	if err != nil {
// 		cancel()
// 		return nil, err
// 	}

// 	go io.Copy(os.Stderr, stderr)

// 	if err := cmd.Start(); err != nil {
// 		cancel()
// 		return nil, err
// 	}

// 	return struct {
// 		io.Reader
// 		io.Closer
// 	}{
// 		Reader: stdout,
// 		Closer: closerFunc(func() error {
// 			defer cancel()
// 			cmd.Wait()
// 			os.Remove(tmp.Name())
// 			return nil
// 		}),
// 	}, nil
// }

// type closerFunc func() error

// func (c closerFunc) Close() error { return c() }

// func IdentityFilesV3(rctx context.Context, archiveBytes []byte) ([]string, error) {
// 	format, stream, err := archives.Identify(rctx, "file", bytes.NewReader(archiveBytes))
// 	if err != nil {
// 		return nil, fmt.Errorf("nepavyko atidaryti archyvo: %w", err)
// 	}
// 	extractor, ok := format.(archives.Extractor)
// 	if !ok {
// 		return nil, fmt.Errorf("formatas %T nepalaiko failų išskleidimo (gali būti, kad tai ne archyvas)", format)
// 	}
// 	var names []string
// 	dir := ""
// 	err = extractor.Extract(context.TODO(), stream, func(ctx context.Context, info archives.FileInfo) error {
// 		if info.IsDir() {
// 			dir = filepath.Join(dir, info.Name())
// 			return nil
// 		}
// 		names = append(names, filepath.Join(dir, info.Name()))
// 		return nil
// 	})

// 	return names, nil
// }

// func GetFileFromArchiveV3(rctx context.Context, archiveBytes []byte, filename string) (io.ReadCloser, error) {
// 	var buf bytes.Buffer
// 	format, stream, err := archives.Identify(rctx, filename, bytes.NewReader(archiveBytes))
// 	if err != nil {
// 		return nil, fmt.Errorf("nepavyko atidaryti archyvo: %w", err)
// 	}
// 	extractor, ok := format.(archives.Extractor)
// 	if !ok {
// 		return nil, fmt.Errorf("formatas %T nepalaiko failų išskleidimo (gali būti, kad tai ne archyvas)", format)
// 	}
// 	dir := ""
// 	err = extractor.Extract(context.TODO(), stream, func(ctx context.Context, info archives.FileInfo) error {
// 		if info.IsDir() {
// 			dir = filepath.Join(dir, info.Name())
// 			return nil
// 		}
// 		if filepath.Join(dir, info.Name()) != filename {
// 			return nil
// 		}
// 		fh, err := info.Open()
// 		if err != nil {
// 			return fmt.Errorf("nepavyko atidaryti failo %q: %w", filename, err)
// 		}
// 		defer fh.Close()
// 		buf.ReadFrom(fh)
// 		return nil
// 	})
// 	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil

// }

// // GetFileFromZip suranda faile esantį įrašą pagal filename ir grąžina jo turinį kaip io.ReadCloser.
// // filename lyginamas pagal basename (pvz. "failas.pdf" ras "dir/sub/failas.pdf").
// // Grąžina nil, nil jei failas nerastas.
// func GetFileFromZip(zipBytes []byte, filename string) (io.ReadCloser, error) {
// 	r := bytes.NewReader(zipBytes)
// 	zr, err := zip.NewReader(r, int64(len(zipBytes)))
// 	if err != nil {
// 		return nil, fmt.Errorf("nepavyko atidaryti zip: %w", err)
// 	}

// 	target := strings.ToLower(path.Base(filename))

// 	for _, f := range zr.File {
// 		if strings.HasSuffix(f.Name, "/") {
// 			continue // katalogas
// 		}
// 		if strings.ToLower(path.Base(f.Name)) == target {
// 			rc, err := f.Open()
// 			if err != nil {
// 				return nil, fmt.Errorf("nepavyko atidaryti failo %q: %w", f.Name, err)
// 			}
// 			return rc, nil
// 		}
// 	}

// 	return nil, fmt.Errorf("failas %q zip’e nerastas", filename)
// }

// func ExtractEmlAttachments(in []byte, filename string, idx string) (io.ReadCloser, error) {
// 	// 1. Atidarome failą
// 	f := bytes.NewReader(in)
// 	index, _ := strconv.Atoi(idx)

// 	// 2. Išparsiname (Enmime padaro visą sunkų darbą)
// 	env, err := enmime.ReadEnvelope(f)
// 	if err != nil {
// 		return nil, fmt.Errorf("klaida skaitant EML: %w", err)
// 	}

// 	// 3. Išsaugome prisegtukus
// 	var buf bytes.Buffer
// 	i := 0
// 	for _, att := range env.Attachments {
// 		if att.FileName != filename {
// 			continue
// 		}
// 		i++
// 		if i < index && index != 0 {
// 			continue
// 		}
// 		// err := os.WriteFile(fullPath, att.Content, 0644)
// 		buf.ReadFrom(bytes.NewReader(att.Content))
// 		break
// 		// if err != nil {
// 		// return fmt.Errorf("nepavyko įrašyti %s: %w", att.FileName, err)
// 		// }
// 	}
// 	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
// }

// func ConvertMsgToEml(in []byte) ([]byte, error) {
// 	file := bytes.NewReader(in)
// 	_ = file

// 	tmpFileName, _ := os.CreateTemp("", "msg-*.msg")
// 	defer os.Remove(tmpFileName.Name())
// 	// log.Printf("Laikinas MSG failas: %s", tmpFileName.Name())
// 	os.WriteFile(tmpFileName.Name(), in, 0755)

// 	cmd := exec.Command("msgconvert", "--outfile", "-", tmpFileName.Name())
// 	stdout, _ := cmd.StdoutPipe()
// 	stderr, _ := cmd.StderrPipe()
// 	go func() {
// 		io.Copy(os.Stderr, stderr)
// 	}()
// 	var buf bytes.Buffer
// 	go func() {
// 		io.Copy(&buf, stdout)
// 	}()
// 	err := cmd.Start()
// 	if err != nil {
// 		return nil, fmt.Errorf("nepavyko konvertuoti MSG į EML: %w", err)
// 	}
// 	err = cmd.Wait()
// 	if err != nil {
// 		return nil, fmt.Errorf("nepavyko konvertuoti MSG į EML: %w", err)
// 	}

// 	return buf.Bytes(), nil
// }

// func normalizeArchivePath(v string) string {
// 	v = strings.ReplaceAll(v, "\\", "/")
// 	v = filepath.ToSlash(v)
// 	v = strings.TrimSpace(v)
// 	v = strings.TrimPrefix(v, "./")
// 	v = strings.TrimPrefix(v, "/")
// 	v = path.Clean(v)
// 	if v == "." {
// 		return ""
// 	}
// 	return v
// }

package ziputil

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	// "github.com/gen2brain/go-unarr"
	// "github.com/gen2brain/go-unarr"
	"github.com/jhillyerd/enmime"
	"github.com/mholt/archives"
)

func GetFileFromZipArchive(zipBytes []byte, filename string) (io.ReadCloser, error) {
	rdr := bytes.NewReader(zipBytes)
	r, err := zip.NewReader(rdr, int64(len(zipBytes)))
	if err != nil {
		return nil, err
	}

	for _, f := range r.File {
		if f.Name == filename {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("nepavyko atidaryti failo %q: %w", filename, err)
			}
			return rc, nil
		}
	}
	return nil, fmt.Errorf("failas %q zip’e nerastas", filename)
}

func detectArchiveType(b []byte) string {
	if len(b) >= 4 && bytes.Equal(b[:4], []byte("PK\x03\x04")) {
		return "zip"
	}
	// RAR4
	if len(b) >= 7 && bytes.Equal(b[:7], []byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x00}) {
		return "rar"
	}
	// RAR5
	if len(b) >= 8 && bytes.Equal(b[:8], []byte{0x52, 0x61, 0x72, 0x21, 0x1a, 0x07, 0x01, 0x00}) {
		return "rar"
	}
	if len(b) >= 6 && bytes.Equal(b[:6], []byte("7z\xBC\xAF\x27\x1C")) {
		return "7z"
	}
	return "unknown"
}

func IdentityFilesV2(ctx context.Context, archiveBytes []byte) ([]string, error) {
	switch detectArchiveType(archiveBytes) {

	case "zip":
		return listZip(archiveBytes)

	// case "7z":
	// 	return listWith7z(archiveBytes)

	default:
		return IdentityFilesV3(archiveBytes)
	}
}

func listZip(b []byte) ([]string, error) {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}

	var out []string
	for _, f := range r.File {
		if !strings.HasSuffix(f.Name, "/") {
			out = append(out, f.Name)
		}
	}
	return out, nil
}

func GetFileFromArchiveV2(ctx context.Context, archiveBytes []byte, filename string) (io.ReadCloser, error) {
	switch detectArchiveType(archiveBytes) {

	case "zip":
		return getFromZip(archiveBytes, filename)

	// case "7z":
	// 	return getWith7z(archiveBytes, filename)

	default:
		return GetFileFromArchiveV3(archiveBytes, filename)
	}
}

func getFromZip(b []byte, filename string) (io.ReadCloser, error) {
	r, err := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if err != nil {
		return nil, err
	}

	for _, f := range r.File {
		if f.Name == filename {
			return f.Open()
		}
	}
	return nil, fmt.Errorf("failas nerastas: %s", filename)
}

func listWith7z(b []byte) ([]string, error) {
	tmp, err := os.CreateTemp("", "arc-*")
	if err != nil {
		return nil, err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return nil, err
	}
	defer os.Remove(tmp.Name())
	if err := os.WriteFile(tmp.Name(), b, 0600); err != nil {
		return nil, err
	}

	cmd := exec.Command("7z", "l", "-slt", tmp.Name())
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var files []string
	var currentPath string
	inEntry := false
	isDir := false
	flushEntry := func() {
		if currentPath != "" && !isDir {
			files = append(files, currentPath)
		}
		currentPath = ""
		isDir = false
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "----------" {
			inEntry = true
			flushEntry()
			continue
		}
		if !inEntry {
			continue
		}
		if line == "" {
			flushEntry()
			continue
		}
		if strings.HasPrefix(line, "Path = ") {
			currentPath = strings.TrimPrefix(line, "Path = ")
		}
		if strings.HasPrefix(line, "Folder = ") {
			isDir = strings.TrimPrefix(line, "Folder = ") == "+"
		}
	}
	flushEntry()
	return files, scanner.Err()
}

func getWith7z(b []byte, filename string) (io.ReadCloser, error) {
	tmp, err := os.CreateTemp("", "arc-*")
	if err != nil {
		return nil, err
	}
	tmpName := tmp.Name()
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return nil, err
	}
	if err := os.WriteFile(tmpName, b, 0600); err != nil {
		_ = os.Remove(tmpName)
		return nil, err
	}

	cmd := exec.Command(
		"7z", "x",
		"-so",
		tmpName,
		filename,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = os.Remove(tmpName)
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		_ = stdout.Close()
		_ = os.Remove(tmpName)
		return nil, err
	}

	return &sevenZipReadCloser{
		stdout:  stdout,
		cmd:     cmd,
		tmpPath: tmpName,
		stderr:  &stderr,
	}, nil
}

type closerFunc func() error

func (c closerFunc) Close() error { return c() }

type sevenZipReadCloser struct {
	stdout  io.ReadCloser
	cmd     *exec.Cmd
	tmpPath string
	stderr  *bytes.Buffer

	once     sync.Once
	closeErr error
}

func (s *sevenZipReadCloser) Read(p []byte) (int, error) {
	return s.stdout.Read(p)
}

func (s *sevenZipReadCloser) Close() error {
	s.once.Do(func() {
		var errs []error

		if s.stdout != nil {
			if err := s.stdout.Close(); err != nil {
				errs = append(errs, err)
			}
		}

		if s.cmd != nil && s.cmd.Process != nil {
			if err := s.cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
				errs = append(errs, err)
			}
			if err := s.cmd.Wait(); err != nil {
				if s.stderr != nil && s.stderr.Len() > 0 {
					errs = append(errs, fmt.Errorf("7z wait failed: %w: %s", err, strings.TrimSpace(s.stderr.String())))
				} else {
					errs = append(errs, fmt.Errorf("7z wait failed: %w", err))
				}
			}
		}

		if err := os.Remove(s.tmpPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err)
		}

		s.closeErr = errors.Join(errs...)
	})

	return s.closeErr
}

func IdentityFilesV3(archiveBytes []byte) ([]string, error) {
	format, stream, err := archives.Identify(context.TODO(), "file", bytes.NewReader(archiveBytes))
	if err != nil {
		return nil, fmt.Errorf("nepavyko atidaryti archyvo: %w", err)
	}
	extractor, ok := format.(archives.Extractor)
	if !ok {
		return nil, fmt.Errorf("formatas %T nepalaiko failų išskleidimo (gali būti, kad tai ne archyvas)", format)
	}
	var names []string
	seen := make(map[string]struct{})
	var dirs []string
	err = extractor.Extract(context.TODO(), stream, func(ctx context.Context, info archives.FileInfo) error {
		if info.IsDir() {
			dirs = append(dirs, info.Name())
			log.Printf("debug1: %v", info.IsDir())
			return nil
		}
		name := normalizeArchivePath(info.Name())
		log.Printf("debug2: %v", name)
		if name == "" {
			return nil
		}
		if _, ok := seen[name]; ok {
			return nil
		}
		seen[name] = struct{}{}
		names = append(names, name)
		var tmpDir string
		for _, d := range dirs {
			tmpDir += d + "/"
			names = append(names, tmpDir+name)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return names, nil
}

func GetFileFromArchiveV3(archiveBytes []byte, filename string) (io.ReadCloser, error) {
	log.Printf("debug10: get file from archive3 %s", filename)
	var buf bytes.Buffer
	target := normalizeArchivePath(filename)
	targetBase := strings.ToLower(path.Base(target))
	log.Printf("debug20: target=%q targetBase=%q", target, strings.ToLower(path.Base(target)))
	if target == "" {
		return nil, fmt.Errorf("failas nerastas: %s", filename)
	}

	format, stream, err := archives.Identify(context.TODO(), filename, bytes.NewReader(archiveBytes))
	if err != nil {
		return nil, fmt.Errorf("nepavyko atidaryti archyvo: %w", err)
	}
	extractor, ok := format.(archives.Extractor)
	if !ok {
		return nil, fmt.Errorf("formatas %T nepalaiko failų išskleidimo (gali būti, kad tai ne archyvas)", format)
	}
	matched := false
	err = extractor.Extract(context.TODO(), stream, func(ctx context.Context, info archives.FileInfo) error {
		if info.IsDir() {
			log.Printf("debug5: %v", info.IsDir())
			return nil
		}
		entryName := normalizeArchivePath(info.Name())
		log.Printf("debug5: %v", entryName)
		if entryName == "" {
			return nil
		}

		if entryName != target && strings.ToLower(path.Base(entryName)) != targetBase {
			return nil
		}
		fh, err := info.Open()
		if err != nil {
			return fmt.Errorf("nepavyko atidaryti failo %q: %w", filename, err)
		}
		defer fh.Close()
		if _, err := buf.ReadFrom(fh); err != nil {
			return fmt.Errorf("nepavyko nuskaityti failo %q: %w", filename, err)
		}
		matched = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, fmt.Errorf("failas nerastas: %s", filename)
	}
	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil

}

func normalizeArchivePath(v string) string {
	v = strings.ReplaceAll(v, "\\", "/")
	v = filepath.ToSlash(v)
	v = strings.TrimSpace(v)
	v = strings.TrimPrefix(v, "./")
	v = strings.TrimPrefix(v, "/")
	v = path.Clean(v)
	if v == "." {
		return ""
	}
	return v
}

// GetFileFromZip suranda faile esantį įrašą pagal filename ir grąžina jo turinį kaip io.ReadCloser.
// filename lyginamas pagal basename (pvz. "failas.pdf" ras "dir/sub/failas.pdf").
// Grąžina nil, nil jei failas nerastas.
func GetFileFromZip(zipBytes []byte, filename string) (io.ReadCloser, error) {
	r := bytes.NewReader(zipBytes)
	zr, err := zip.NewReader(r, int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("nepavyko atidaryti zip: %w", err)
	}

	target := strings.ToLower(path.Base(filename))

	for _, f := range zr.File {
		if strings.HasSuffix(f.Name, "/") {
			continue // katalogas
		}
		if strings.ToLower(path.Base(f.Name)) == target {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("nepavyko atidaryti failo %q: %w", f.Name, err)
			}
			return rc, nil
		}
	}

	return nil, fmt.Errorf("failas %q zip’e nerastas", filename)
}

func ExtractEmlAttachments(in []byte, filename string, idx string) (io.ReadCloser, error) {
	// 1. Atidarome failą
	f := bytes.NewReader(in)
	index, _ := strconv.Atoi(idx)

	// 2. Išparsiname (Enmime padaro visą sunkų darbą)
	env, err := enmime.ReadEnvelope(f)
	if err != nil {
		return nil, fmt.Errorf("klaida skaitant EML: %w", err)
	}

	// 3. Išsaugome prisegtukus
	var buf bytes.Buffer
	i := 0
	for _, att := range env.Attachments {
		if att.FileName != filename {
			continue
		}
		i++
		if i < index && index != 0 {
			continue
		}
		// err := os.WriteFile(fullPath, att.Content, 0644)
		buf.ReadFrom(bytes.NewReader(att.Content))
		break
		// if err != nil {
		// return fmt.Errorf("nepavyko įrašyti %s: %w", att.FileName, err)
		// }
	}
	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

func ConvertMsgToEml(in []byte) ([]byte, error) {
	tmpFile, err := os.CreateTemp("", "msg-*.msg")
	if err != nil {
		return nil, err
	}
	tmpName := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		_ = os.Remove(tmpName)
		return nil, err
	}
	defer os.Remove(tmpName)

	if err := os.WriteFile(tmpName, in, 0600); err != nil {
		return nil, err
	}

	cmd := exec.Command("msgconvert", "--outfile", "-", tmpName)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return nil, fmt.Errorf("nepavyko konvertuoti MSG į EML: %w: %s", err, strings.TrimSpace(stderr.String()))
		}
		return nil, fmt.Errorf("nepavyko konvertuoti MSG į EML: %w", err)
	}

	return stdout.Bytes(), nil
}
