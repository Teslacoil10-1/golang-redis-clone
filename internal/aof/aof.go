// the ffi and implementation module for the multi part aof system
package aof

/*
#cgo LDFLAGS: -L../../aof_engine/target/release -laof_engine -lpthread -ldl
#include <stdlib.h>

extern void write_to_aof(const char* command);
extern void rotate_aof(const char* tmp_path);
*/
import "C"

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"redis-clone/internal/store"
)

var cmdChan = make(chan string, 4096)

func StartWorker() {
	go func() {
		for cmd := range cmdChan {
			c_cmd := C.CString(cmd)
			C.write_to_aof(c_cmd)
			C.free(unsafe.Pointer(c_cmd))
		}
	}()
}

func RecordCommand(cmd string) {
	cmdChan <- cmd
}

func TriggerRewrite(db *store.Store) error {
	dir := "/data/appendonlydir"

	tmpPath := filepath.Join(dir, fmt.Sprintf("temp-%d.base.aof", time.Now().UnixNano()))

	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	if err := db.DumpState(f); err != nil {
		f.Close()
		return err
	}
	f.Close()

	cTmpPath := C.CString(tmpPath)
	defer C.free(unsafe.Pointer(cTmpPath))

	C.rotate_aof(cTmpPath)

	return nil
}

func ReplayState(db *store.Store) {
	dir := "/data/appendonlydir"

	files, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("[AOF] no aof file found")
			return
		}
		log.Fatalf("[AOF] failed to read aof: %v", err)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".aof") {
			continue
		}

		path := filepath.Join(dir, file.Name())

		f, err := os.Open(path)
		if err != nil {
			log.Printf("[AOF] Failed to open aof file %s: %v", file.Name(), err)
			continue
		}
		log.Printf("[AOF] Replaying aof file %s", file.Name())
		scanner := bufio.NewScanner(f)

		for scanner.Scan() {
			line := scanner.Text()

			parts := strings.SplitN(line, " ", 4)

			if len(parts) == 0 {
				continue
			}

			switch parts[0] {
			case "SET":
				if len(parts) == 3 {
					db.Set(parts[1], parts[2])
				} else if len(parts) == 4 {
					expiresAt, err := strconv.ParseInt(parts[3], 10, 64)
					if err == nil {
						if time.Now().UnixNano() < expiresAt {
							db.SetWithTTL(parts[1], parts[2], expiresAt)
						}
					}
				}
			case "DELETE":
				if len(parts) >= 2 {
					db.Delete(parts[1])
				}

			}
		}
	}
}
