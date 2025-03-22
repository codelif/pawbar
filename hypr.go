package main

import (
	"bufio"
	"encoding/json"
	"net"
	"os"
	"path"
	"strings"
)

type HyprService struct {
	callbacks map[string][]chan<- HyprEvent
	running   bool
}

func MakeHypr() *HyprService {
	h := HyprService{}
	h.callbacks = make(map[string][]chan<- HyprEvent)
	return &h
}
func (h *HyprService) Name() string { return "hypr" }

func (h *HyprService) Start() error {
	if h.running {
		return nil
	}
  logger.Println("Hypr started")
	h.callbacks = make(map[string][]chan<- HyprEvent)
	go h.run()
	h.running = true
	return nil
}
func (h *HyprService) Stop() error {
	return nil
}

func (h *HyprService) RegisterChannel(event string, ch chan<- HyprEvent) {
	h.callbacks[event] = append(h.callbacks[event], ch)
}

func (h *HyprService) run() {
	_, sockaddr2 := GetHyprSocketAddrs()

	sock2, err := net.Dial("unix", sockaddr2)
	if err != nil {
		panic(err)
	}
	defer sock2.Close()

	scanner := bufio.NewScanner(sock2)
	for scanner.Scan() {
		e := NewHyprEvent(scanner.Text())
		c, ok := h.callbacks[e.event]
    logger.Println("Hypr is active")
		if ok {
			for _, ch := range c {
				ch <- e
			}
		}
	}
}

func GetHyprSocketAddrs() (string, string) {
	instance_signature := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	runtime_dir := os.Getenv("XDG_RUNTIME_DIR")
	socket_addr := path.Join(runtime_dir, "/hypr", instance_signature)

	return path.Join(socket_addr, "/.socket.sock"), path.Join(socket_addr, "/.socket2.sock")
}

type HyprEvent struct {
	event string
	data  string
}

func NewHyprEvent(s string) HyprEvent {
	e, d, _ := strings.Cut(s, ">>")
	return HyprEvent{e, strings.Trim(d, " \n")}
}

type Workspace struct {
	Id              int    `json:"id"`
	Name            string `json:"name"`
	Monitor         string `json:"monitor"`
	MonitorID       int    `json:"monitorID"`
	Windows         int    `json:"windows"`
	Hasfullscreen   bool   `json:"hasfullscreen"`
	Lastwindow      string `json:"lastwindow"`
	Lastwindowtitle string `json:"lastwindowtitle"`
}

func GetWorkspaces() []Workspace {
	sockaddr1, _ := GetHyprSocketAddrs()
	sock, err := net.Dial("unix", sockaddr1)
	if err != nil {
		panic(err)
	}
	defer sock.Close()
	scanner := json.NewDecoder(sock)

	sock.Write([]byte("-j/workspaces"))
	var o []Workspace

	err = scanner.Decode(&o)
	if err != nil {
		panic(err)
	}
	return o
}

func GetActiveWorkspace() Workspace {
	sockaddr1, _ := GetHyprSocketAddrs()
	sock, err := net.Dial("unix", sockaddr1)
	if err != nil {
		panic(err)
	}
	defer sock.Close()
	scanner := json.NewDecoder(sock)

	sock.Write([]byte("-j/activeworkspace"))
	var o Workspace

	err = scanner.Decode(&o)
	if err != nil {
		panic(err)
	}
	return o
}

type ClientWS struct {
	Id   int
	Name string
}

type Client struct {
	Address          string      `json:"address"`
	Mapped           bool        `json:"mapped"`
	Hidden           bool        `json:"hidden"`
	At               []int       `json:"at"`
	Size             []int       `json:"size"`
	Workspace        ClientWS    `json:"workspace"`
	Floating         bool        `json:"floating"`
	Pseudo           bool        `json:"pseudo"`
	Monitor          int         `json:"monitor"`
	Class            string      `json:"class"`
	Title            string      `json:"title"`
	InitialClass     string      `json:"initialClass"`
	InitialTitle     string      `json:"initialTitle"`
	Pid              int         `json:"pid"`
	Xwayland         bool        `json:"xwayland"`
	Pinned           bool        `json:"pinned"`
	Fullscreen       int         `json:"fullscreen"`
	FullscreenClient int         `json:"fullscreenClient"`
	Grouped          interface{} `json:"grouped"`
	Tags             interface{} `json:"tags"`
	Swallowing       string      `json:"swallowing"`
	FocusHistoryID   int         `json:"focusHistoryID"`
	InhibitingIdle   bool        `json:"inhibitingIdle"`
}

func GetClients() []Client {
	sockaddr1, _ := GetHyprSocketAddrs()
	sock, err := net.Dial("unix", sockaddr1)
	if err != nil {
		panic(err)
	}
	defer sock.Close()
	scanner := json.NewDecoder(sock)

	sock.Write([]byte("-j/clients"))
	var o []Client

	err = scanner.Decode(&o)
	if err != nil {
		panic(err)
	}
	return o
}

func GoToWorkspace(name string) {
	sockaddr1, _ := GetHyprSocketAddrs()
	sock, err := net.Dial("unix", sockaddr1)
	if err != nil {
		panic(err)
	}
	defer sock.Close()

	sock.Write([]byte("/dispatch workspace " + name))
}
