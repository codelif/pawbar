package main

import (
	"context"
	"log"
	"time"

	"github.com/codelif/pawbar/pkg/dbusmenukitty/panel"
)

// "io"
// "log"
// "os"
// "strconv"
// "syscall"

// "git.sr.ht/~rockorager/vaxis"

func main() {
	// if os.Getenv("DBUSMENUKITTY_FD") == "" {
	// 	log.Fatalln("FD not set")
	// }
	// fdint, err := strconv.Atoi(os.Getenv("DBUSMENUKITTY_FD"))
	// if err != nil {
	// 	log.Fatalln("invalid FD '", os.Getenv("DBUSMENUKITTY_FD"), "'")
	// }
	// fd := uintptr(fdint)

	// // os.NewFile()
	// _, _, e := syscall.Syscall(syscall.SYS_FCNTL, fd, syscall.F_GETFD, 0)
	// if e != 0 {
	// 	log.Fatalln("FD not open")
	// }

	// f := os.NewFile(fd, "pipe")
	// log.SetOutput(f)

	// if os.Getenv("DBUSMENUKITTY_SOCKET") == "" {
	// 	log.Fatalln("DBUSMENUKITTY_SOCKET not set")
	// }
	p, pcon, err := panel.NewPanel(context.Background(), panel.Config{
		Name:        "pawbar",
		Size:        panel.Vector{20, 20},
		Edge:        panel.EdgeNone,
		FocusPolicy: panel.FocusOnDemand,
		WithSignals: true,
		Layer:       panel.LayerTop,
	})

	if err != nil {
		log.Fatalf("can't spawn a panel: %v\n", err)
	}

	time.Sleep(time.Second)
	p.Dispatch("action", nil)
	p.Dispatch("env", nil)
	p.Dispatch("get-text", nil)

	<-pcon.Done()
}
