// the ffi and implementation module for the multi part aof system
package aof

/*
#cgo LDFLAGS: -L../../aof_engine/target/release -laof_engine -lpthread -ldl
#include <stdlib.h>

extern void write_to_aof(const char* command);
extern void trigger_bgrewriteaof();
*/
import "C"

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func TriggerRewrite() {
	C.trigger_bgrewriteaof()
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

			parts := strings.SplitN(line, " ", 3)

			if len(parts) == 0 {
				continue
			}

			switch parts[0] {
			case "SET":
				if len(parts) >= 3 {
					db.Set(parts[1], parts[2])
				}
			case "DELETE":
				if len(parts) >= 2 {
					db.Delete(parts[1])
				}

			}
		}

	}
}
