package panel

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	kittyCmd = "kitty"
)

type Vector struct {
	X, Y int
}

//go:generate stringer -type=Layer -linecomment
type Layer int

const (
	LayerBackground Layer = iota // background
	LayerBottom                  // bottom
	LayerTop                     // top
	LayerOverlay                 // overlay
)

//go:generate stringer -type=FocusPolicy -linecomment
type FocusPolicy int

const (
	FocusExclusive  FocusPolicy = iota // exclusive
	FocusNotAllowed                    // not-allowed
	FocusOnDemand                      // on-demand
)

//go:generate stringer -type=Edge -linecomment
type Edge int

const (
	EdgeBackground  Edge = iota // background
	EdgeBottom                  // bottom
	EdgeCenter                  // center
	EdgeCenterSized             // center-sized
	EdgeLeft                    // left
	EdgeNone                    // none
	EdgeRight                   // right
	EdgeTop                     // top
)

type Config struct {
	Position    Vector
	Size        Vector
	Name        string
	Layer       Layer
	FocusPolicy FocusPolicy
	Edge        Edge

	// WithSignals causes Panel to install signal handlers to cancel panel context.
	// Specifically os.Interrupt (which is SIGINT for most systems), and SIGHUP
	// on receving these signals, the panel process will be terminated and context
	// be cancelled.
	// This differs from calling Panel.Stop since Panel.Stop tries to gracefully shutdown
	// the child panel process first before
	WithSignals bool
}

type Panel struct {
	process    *exec.Cmd
	socketPath string
	pos        Vector
	size       Vector
	config     Config
	cancel     context.CancelFunc
}

type KittySockMsg struct {
	Command       string          `json:"cmd"`
	Version       [3]uint64       `json:"version"`
	NoResponse    bool            `json:"no_response,omitempty"`
	KittyWindowId uint64          `json:"kitty_window_id,omitempty"`
	Payload       json.RawMessage `json:"payload,omitempty"`
}

var kittyRCHeader = append([]byte{0x1b}, "P@kitty-cmd"...)

func packMsg(msg []byte) []byte {
	return append(append(kittyRCHeader, msg...), 0x1b, '\\')
}

// For dispatching commands only, no response is assumed.
func (w *Panel) Dispatch(cmd string, payload any) error {
	// new dispatch, new socket; coz I am too lazy,
	// and its a pain to manage. Though I'll keep a todo below:
	// TODO: keep only one socket connection for Panel's lifetime
	sock, err := net.Dial("unix", w.socketPath)
	if err != nil {
		return err
	}

	defer sock.Close()

	var p []byte
	if payload == nil {
		// easiest way to induce omitempty for payload
		p = []byte{}
	} else {
		p, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	mraw := KittySockMsg{Command: cmd, Version: [3]uint64{0, 42, 0}, Payload: p}

	msg, err := json.Marshal(mraw)
	if err != nil {
		return err
	}

	fmt.Println(string(msg))
	_, err = sock.Write(packMsg(msg))
	if err != nil {
		return err
	}

	resp_header := make([]byte, len(kittyRCHeader))
	sock.Read(resp_header)

	var resp map[string]any
	dec := json.NewDecoder(sock)
	err = dec.Decode(&resp)
	if err != nil {
		return err
	}

	fmt.Println(resp)
	resp_ok_any, ok := resp["ok"]
	if !ok {
		return fmt.Errorf("invalid response schema: field 'ok' not found")
	}

	resp_ok, ok := resp_ok_any.(bool)
	if !ok {
		return fmt.Errorf("invalid response schema: field 'ok' not a boolean")
	}

	if !resp_ok {
		resp_error_any, ok := resp["error"]
		if !ok {
			return fmt.Errorf("invalid response schema: field 'error' not found")
		}

		resp_error, ok := resp_error_any.(string)
		if !ok {
			return fmt.Errorf("invalid response schema: field 'error' not a string")
		}

		return fmt.Errorf(resp_error)
	}

	return nil
}

// Stop tries to gracefully shutdown the panel first
// if panel does not close within a second, it kills it.
func (p *Panel) Stop() {
	p.Dispatch("close-window", nil)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	go func() {
		p.process.Wait()
		cancel()
	}()
	<-ctx.Done()
	p.cancel()
}

func NewPanel(parent context.Context, config Config) (*Panel, context.Context, error) {
	if len(config.Name) == 0 {
		return nil, nil, fmt.Errorf("Name should be a non-zero string.")
	}

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	if config.WithSignals {
		ctx, cancel = signal.NotifyContext(parent, os.Interrupt, syscall.SIGHUP)
	} else {
		ctx, cancel = context.WithCancel(parent)
	}

	socketPath := fmt.Sprintf("/tmp/dbusmenukitty-%s-%d", config.Name, os.Getpid())
	args := []string{
		"+kitten", "panel",
		"--edge", config.Edge.String(),
		"--layer", config.Layer.String(),
		"--focus-policy", config.FocusPolicy.String(),
		"--listen-on", "unix:" + socketPath,
		"-o", "allow_remote_control=socket-only",
		"--class", config.Name,
		"--lines", strconv.Itoa(config.Size.Y),
		"--columns", strconv.Itoa(config.Size.X),
		"--margin-top", strconv.Itoa(config.Position.Y),
		"--margin-left", strconv.Itoa(config.Position.X),
		"box",
	}
	c := exec.CommandContext(ctx, kittyCmd, args...)
	c.Env = append(c.Environ(), "DBUSMENUKITTY_SOCKET="+socketPath)

	if err := c.Start(); err != nil {
		cancel()
		return nil, nil, err
	}
	go func() {
		c.Wait()
		cancel()
	}()
	return &Panel{c, socketPath, config.Position, config.Size, config, cancel}, ctx, nil
}
