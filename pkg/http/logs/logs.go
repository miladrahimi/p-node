// Package logs serves the log files of a storage/logs directory over HTTP: an
// index of the files and a tail -f style chunked stream of one file. It is
// shared by P-Node (its own logs) and P-Manager (its logs, and node logs
// proxied through it).
package logs

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	// DefaultTail is the number of trailing lines sent before following.
	DefaultTail = 200
	// MaxTail caps the trailing lines a client may ask for.
	MaxTail = 5000
	// tailBytes bounds how far back the initial tail reads.
	tailBytes = 4 << 20
	// PollInterval is how often the stream checks the file for new data.
	PollInterval = 500 * time.Millisecond
)

// File describes one log file in the index response.
type File struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	ModifiedAt int64  `json:"modified_at"`
}

// Index lists the log files in the logs directory.
func Index(dir string) echo.HandlerFunc {
	return func(c echo.Context) error {
		entries, err := os.ReadDir(dir)
		if err != nil && !os.IsNotExist(err) {
			return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Cannot read the logs directory."})
		}
		files := make([]File, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".log") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			files = append(files, File{Name: e.Name(), Size: info.Size(), ModifiedAt: info.ModTime().UnixMilli()})
		}
		sort.Slice(files, func(i, j int) bool { return files[i].Name < files[j].Name })
		return c.JSON(http.StatusOK, files)
	}
}

// ValidName reports whether name is a plain *.log file name (no path parts).
func ValidName(name string) bool {
	return name != "" && name == filepath.Base(name) && strings.HasSuffix(name, ".log")
}

// Stream streams a log file as chunked plain text: the last `lines` lines
// first, then every byte appended to the file until the client disconnects or
// done is closed (tail -f). A truncated or replaced file is picked up from its
// start.
func Stream(dir string, done <-chan struct{}) echo.HandlerFunc {
	return func(c echo.Context) error {
		name := c.Param("name")
		if !ValidName(name) {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Log file not found."})
		}
		path := filepath.Join(dir, name)

		tail := DefaultTail
		if v, err := strconv.Atoi(c.QueryParam("lines")); err == nil && v >= 0 {
			tail = min(v, MaxTail)
		}

		file, err := os.Open(path)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"message": "Log file not found."})
		}
		defer func() { _ = file.Close() }()

		res := c.Response()
		res.Header().Set(echo.HeaderContentType, "text/plain; charset=utf-8")
		res.Header().Set("X-Accel-Buffering", "no")
		res.WriteHeader(http.StatusOK)

		offset, err := writeTail(res, file, tail)
		if err != nil {
			return nil
		}
		if _, err = file.Seek(offset, io.SeekStart); err != nil {
			return nil
		}
		res.Flush()

		ticker := time.NewTicker(PollInterval)
		defer ticker.Stop()
		ctx := c.Request().Context()
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-done:
				return nil
			case <-ticker.C:
			}

			opened, err := file.Stat()
			if err != nil {
				return nil
			}
			current, err := os.Stat(path)
			replaced := err == nil && !os.SameFile(opened, current)
			if replaced || opened.Size() < offset {
				// Rotated (removed and recreated) or truncated: start over.
				if replaced {
					_ = file.Close()
					if file, err = os.Open(path); err != nil {
						return nil
					}
				}
				offset = 0
				if _, err = file.Seek(0, io.SeekStart); err != nil {
					return nil
				}
			}

			n, err := io.Copy(res, file)
			offset += n
			if err != nil {
				return nil
			}
			if n > 0 {
				res.Flush()
			}
		}
	}
}

// writeTail writes the last n lines of file to w (reading at absolute offsets,
// so the file position is left untouched) and returns the file size the
// following should continue from.
func writeTail(w io.Writer, file *os.File, n int) (int64, error) {
	info, err := file.Stat()
	if err != nil {
		return 0, err
	}
	size := info.Size()
	if n == 0 {
		return size, nil
	}

	start := max(size-tailBytes, 0)
	buf := make([]byte, size-start)
	if _, err = file.ReadAt(buf, start); err != nil && err != io.EOF {
		return 0, err
	}
	if start > 0 {
		// Drop the partial first line of a bounded read.
		if i := bytes.IndexByte(buf, '\n'); i >= 0 {
			buf = buf[i+1:]
		} else {
			buf = nil
		}
	}

	// Walk back n newlines (ignoring a trailing one) to find the tail start.
	end := len(buf)
	if end > 0 && buf[end-1] == '\n' {
		end--
	}
	pos := end
	for ; n > 0 && pos > 0; n-- {
		i := bytes.LastIndexByte(buf[:pos], '\n')
		if i < 0 {
			pos = 0
			break
		}
		pos = i
	}
	if pos > 0 {
		pos++
	}

	_, err = w.Write(buf[pos:])
	return size, err
}
