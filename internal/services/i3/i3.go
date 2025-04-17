package i3

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"

	"github.com/codelif/pawbar/internal/services"
)

const ipcMagic = "i3-ipc"
const I3_IPC_MESSAGE_TYPE_SUBSCRIBE = 2
const IPC_GET_WORKSPACES = 1
const msgTypeGetTree = 4
var event I3Event

type Service struct {
	callbacks map[string][]chan<- I3Event
	running   bool
}

type WinInfo struct {
		Class string `json:"class"`
		Title string `json:"title"`
}

type WsInfo struct {
		Focused         bool               `json:"focused"`
		WindowInfo      WinInfo            `json:"window_properties"`
}

type WsIdentity struct {
		Focused         bool              `json:"focused"`
		Urgent          bool              `json:"urgent"`
		ScratchpadState string            `json:"scratchpad_state"`
		Name            string            `json:"name"`
		Nodes           []WsInfo          `json:"nodes"`
}

type I3Event struct {
		Change  string     `json:"change"`
		Current WsIdentity `json:"current"`
		Old     WsIdentity `json:"old"`
}

type Workspace struct {
	Id     int    `json:"num"`
	Name    string `json:"name"`
	Focused bool   `json:"focused"`
	Urgent  bool   `json:"urgent"`
}

type WindowProperties struct {
	Class string `json:"class"`
	Title string `json:"title"`
}

type I3Node struct {
	Focused          bool               `json:"focused"`
	Nodes            []I3Node           `json:"nodes"`
	FloatingNodes    []I3Node           `json:"floating_nodes"`
	WindowProperties *WindowProperties  `json:"window_properties"`
}

func Register() {
	services.StartService("i3", &Service{})
}

func GetService() (*Service, bool) {
	if s, ok := services.ServiceRegistry["i3"].(*Service); ok {
		return s, true
	}
	return nil, false
}

func (i *Service) Name() string { return "i3" }

func (i *Service) Start() error {
	if i.running {
		return nil
	}
	i.callbacks = make(map[string][]chan<- I3Event)
	go i.sockMsg()
	i.running = true
	return nil
}

func (i *Service) Stop() error {
	return nil
}

func (i *Service) RegisterChannel(event string, ch chan<- I3Event) {
	i.callbacks[event] = append(i.callbacks[event], ch)
}

func connectToI3() (net.Conn, error) {
	sockPath := os.Getenv("I3SOCK")
	if sockPath == "" {
		return nil, fmt.Errorf("I3SOCK environment variable is not set")
	}

	conn, err := net.Dial("unix", sockPath)
	if err != nil {
		return nil, fmt.Errorf("error connecting to i3 socket: %v", err)
	}

	return conn, nil
}

func sendI3Message(conn net.Conn, messageType uint32, payload []byte) error {
	header := make([]byte, 14)
	copy(header[:6], []byte(ipcMagic))
	binary.LittleEndian.PutUint32(header[6:10], uint32(len(payload)))
	binary.LittleEndian.PutUint32(header[10:14], messageType)

	sendMsg := append(header, payload...)

	if _, err := conn.Write(sendMsg); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}
	return nil
}

func readI3Ack(conn net.Conn) (string, error) {
	const headerSize = 14
	ackHeader := make([]byte, headerSize)
	if _, err := io.ReadFull(conn, ackHeader); err != nil {
		return "", fmt.Errorf("error reading ack header: %v", err)
	}

	if string(ackHeader[0:6]) != "i3-ipc" {
		return "", fmt.Errorf("invalid magic in header: expected 'i3-ipc', got '%s'", string(ackHeader[0:6]))
	}

	ackLen := binary.LittleEndian.Uint32(ackHeader[6:10])
	ackPayload := make([]byte, ackLen)
	if _, err := io.ReadFull(conn, ackPayload); err != nil {
		return "", fmt.Errorf("error reading ack payload: %v", err)
	}

	return string(ackPayload), nil
}

func readResponse(conn net.Conn) ([]byte, error) {
	responseHeader := make([]byte, 14)
	if _, err := io.ReadFull(conn, responseHeader); err != nil {
		return nil, fmt.Errorf("error reading response header: %v", err)
	}
	if string(responseHeader[:6]) != "i3-ipc" {
		return nil, fmt.Errorf("invalid response magic: expected '%s', got '%s'",ipcMagic , string(responseHeader[:6]))
	}

	payloadLength := binary.LittleEndian.Uint32(responseHeader[6:10])
	payloadData := make([]byte, payloadLength)

	if _, err := io.ReadFull(conn, payloadData); err != nil {
		return nil, fmt.Errorf("error reading payload data: %v", err)
	}

	return payloadData, nil
}

func (i *Service) sockMsg(){

	conn, err := connectToI3()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	subscription := []string{"workspace"}
	payload, err := json.Marshal(subscription)
	if err != nil {
		fmt.Printf("Error marshaling subscription payload: %v\n", err)
		os.Exit(1)
	}

	if err := sendI3Message(conn, I3_IPC_MESSAGE_TYPE_SUBSCRIBE, payload); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	ack, err := readI3Ack(conn)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Subscription Acknowledgment:", ack)

	for {
		eventPayload, err := readResponse(conn)
		if err != nil {
			fmt.Println("Error reading response:", err)
			break
		}

		if err := json.Unmarshal(eventPayload, &event); err != nil {
			fmt.Println("Error unmarshaling event:", err)
			continue
		}

		if chans, ok := i.callbacks["workspaces"]; ok {
				for _, ch := range chans {
				ch <- event
			}
		}
	}
}

func GetWorkspaces() []Workspace{
	conn, err := connectToI3()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	payload := []byte("")

	if err := sendI3Message(conn, IPC_GET_WORKSPACES,payload); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	eventPayload, err := readResponse(conn)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	
	var workspaces []Workspace
	if err = json.Unmarshal(eventPayload, &workspaces); err != nil {
		fmt.Println("Error unmarshaling JSON:", err)
		return nil
	}

	return workspaces
}

func GoToWorkspace(name string){
	cmd := exec.Command("i3-msg", "workspace", name)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}

func GetActiveWorkspace() Workspace {
	workspaces := GetWorkspaces()
	for _, ws := range workspaces {
		if ws.Focused {
			return ws
		}
	}
	return Workspace{}
}

func GetTitleClass() (string, string){
	conn, err := connectToI3()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	payload := []byte("")

	if err := sendI3Message(conn, msgTypeGetTree,payload); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	eventPayload, err := readResponse(conn)
	if err != nil {
		fmt.Println(err)
		return "", ""
	}

	var root I3Node
	if err := json.Unmarshal(eventPayload, &root); err != nil {
		fmt.Println("Failed to parse JSON: %v", err)
	}

	var focusedProps *WindowProperties
	var findFocused func(n *I3Node)
	findFocused = func(n *I3Node) {
		if n.Focused && n.WindowProperties != nil {
			focusedProps = n.WindowProperties
			return
		}
		for i := range n.Nodes {
			findFocused(&n.Nodes[i])
		}
		for i := range n.FloatingNodes {
			findFocused(&n.FloatingNodes[i])
		}
	}
	findFocused(&root)

	return focusedProps.Class, focusedProps.Title

}





