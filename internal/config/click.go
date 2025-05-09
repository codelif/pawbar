package config

import (
	"fmt"
	"os/exec"
	"reflect"

	"git.sr.ht/~rockorager/vaxis"
	"gopkg.in/yaml.v3"
)

func ButtonName(ev vaxis.Mouse) string {
	switch ev.Button {
	case vaxis.MouseLeftButton:
		return "left"
	case vaxis.MouseRightButton:
		return "right"
	case vaxis.MouseMiddleButton:
		return "middle"
	case vaxis.MouseWheelUp:
		return "wheel-up"
	case vaxis.MouseWheelDown:
		return "wheel-down"
	case 66:
		return "wheel-left"
	case 67:
		return "wheel-right"
	default:
		return ""
	}
}

var allowedButtons = []string{
	"left", "right", "middle",
	"wheel-up", "wheel-down", "wheel-left", "wheel-right",
	"hover", // yea its a button now ;)
}

var allowedButtonsSet = func() map[string]struct{} {
	m := make(map[string]struct{}, len(allowedButtons))
	for _, b := range allowedButtons {
		m[b] = struct{}{}
	}
	return m
}()

func ParseCursor(s string) (vaxis.MouseShape, error) {
	switch s {
	case "pointer":
		return vaxis.MouseShapeClickable, nil
	case "default", "":
		return vaxis.MouseShapeDefault, nil
	case "text":
		return vaxis.MouseShapeTextInput, nil
	default:
		return "", fmt.Errorf("invalid cursor name: %q", s)
	}
}

func validateOnMouseNode(node *yaml.Node) error {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == "onmouse" {
			val := node.Content[i+1]
			if val.Kind != yaml.MappingNode {
				return fmt.Errorf("onmouse must be a map")
			}
			for j := 0; j < len(val.Content); j += 2 {
				btn := val.Content[j].Value
				if _, ok := allowedButtonsSet[btn]; !ok {
					return fmt.Errorf("invalid onmouse key %q; must be one of %v",
						btn, allowedButtons)
				}
			}
		}
	}
	return nil
}

type MouseActions[T any] struct {
	Actions map[string]*MouseAction[T]

	hoverActive   bool
	hoverSnapshot T
	hoverIndex    int
}

// yea, couldn't be bothered to validate the whole thing here
// TODO: unmarshal and validate here
func (m *MouseActions[T]) UnmarshalYAML(node *yaml.Node) error {
	var tmp map[string]*MouseAction[T]
	if err := node.Decode(&tmp); err != nil {
		return err
	}
	m.Actions = tmp
	return nil
}

type MouseAction[T any] struct {
	Run     []string `yaml:"run"`
	Notify  string   `yaml:"notify"`
	Configs []T      `yaml:"config"`

	clickIndex int
	inited     bool
}

func (m *MouseActions[T]) Dispatch(
	button string,
	initOpts, liveOpts any,
) bool {
	act, ok := m.Actions[button]
	if !ok {
		return false
	}

	act.DispatchAction()

	clicked := act.Next(initOpts, liveOpts)

	if m.hoverActive && clicked {
		m.hoverSnapshot = snapTFields[T](liveOpts, act.Configs)
	}

	return clicked
}

func (m *MouseActions[T]) HoverIn(liveOpts any) bool {
	hoverAct, ok := m.Actions["hover"]

	if ok {
		hoverAct.DispatchAction()
	}

	if !ok || len(hoverAct.Configs) == 0 {
		return false
	}

	if !m.hoverActive {
		m.hoverActive = true
		m.hoverSnapshot = snapTFields[T](liveOpts, hoverAct.Configs)
	}

	idx := m.hoverIndex % len(hoverAct.Configs)
	applyOverride(hoverAct.Configs[idx], liveOpts)
	m.hoverIndex++
	return true
}

func (m *MouseActions[T]) HoverOut(liveOpts any) bool {
	if !m.hoverActive {
		return false
	}
	m.hoverActive = false

	restoreTFields[T](m.hoverSnapshot, liveOpts)
	return true
}

func snapTFields[T any](liveOpts any, cfgs []T) T {
	liveV := reflect.ValueOf(liveOpts)
	if liveV.Kind() == reflect.Ptr {
		liveV = liveV.Elem()
	}
	tT := reflect.TypeOf(cfgs).Elem()
	snap := reflect.New(tT).Elem()
	for i := 0; i < tT.NumField(); i++ {
		fd := tT.Field(i)
		lf := liveV.FieldByName(fd.Name)
		if !lf.IsValid() {
			continue
		}
		p := reflect.New(fd.Type.Elem())
		p.Elem().Set(lf)
		snap.Field(i).Set(p)
	}
	return snap.Interface().(T)
}

func restoreTFields[T any](snap T, liveOpts any) {
	liveV := reflect.ValueOf(liveOpts)
	if liveV.Kind() == reflect.Ptr {
		liveV = liveV.Elem()
	}
	snapV := reflect.ValueOf(snap)
	if snapV.Kind() == reflect.Ptr {
		snapV = snapV.Elem()
	}
	for i := 0; i < snapV.NumField(); i++ {
		f := snapV.Field(i)
		if f.Kind() == reflect.Ptr && !f.IsNil() {
			name := snapV.Type().Field(i).Name
			if lf := liveV.FieldByName(name); lf.CanSet() {
				lf.Set(f.Elem())
			}
		}
	}
}

func applyOverride[T any](cfg T, liveOpts any) {
	liveV := reflect.ValueOf(liveOpts)
	if liveV.Kind() == reflect.Ptr {
		liveV = liveV.Elem()
	}
	part := reflect.ValueOf(cfg)
	if part.Kind() == reflect.Ptr {
		part = part.Elem()
	}
	for i := 0; i < part.NumField(); i++ {
		f := part.Field(i)
		if f.Kind() == reflect.Ptr && !f.IsNil() {
			name := part.Type().Field(i).Name
			if lf := liveV.FieldByName(name); lf.CanSet() {
				lf.Set(f.Elem())
			}
		}
	}
}

func (a *MouseAction[T]) UnmarshalYAML(n *yaml.Node) error {
	// must be a mapping: run:, notify:, config:
	if n.Kind != yaml.MappingNode {
		return fmt.Errorf("onmouse action must be a map, got %v", n.Kind)
	}

	var runNode, notifyNode, configNode *yaml.Node
	for i := 0; i < len(n.Content); i += 2 {
		key := n.Content[i].Value
		val := n.Content[i+1]
		switch key {
		case "run":
			runNode = val
		case "notify":
			notifyNode = val
		case "config":
			configNode = val
		}
	}

	if runNode != nil {
		if runNode.Kind == yaml.SequenceNode {
			if err := runNode.Decode(&a.Run); err != nil {
				return fmt.Errorf("onmouse: bad run: %w", err)
			}
		} else {
			var s string
			if err := runNode.Decode(&s); err != nil {
				return fmt.Errorf("onmouse: bad run: %w", err)
			}
			a.Run = []string{s}
		}
	}

	if notifyNode != nil {
		if err := notifyNode.Decode(&a.Notify); err != nil {
			return fmt.Errorf("onmouse: bad notify: %w", err)
		}
	}

	if configNode != nil {
		switch configNode.Kind {
		case yaml.SequenceNode:
			var list []T
			if err := configNode.Decode(&list); err != nil {
				return fmt.Errorf("onmouse: bad config list: %w", err)
			}
			a.Configs = list

		case yaml.MappingNode:
			var single T
			if err := configNode.Decode(&single); err != nil {
				return fmt.Errorf("onmouse: bad config map: %w", err)
			}
			a.Configs = []T{single}

		default:
			return fmt.Errorf("onmouse.config must be a map or sequence of maps, got %v", configNode.Kind)
		}
	}

	return nil
}

func (a *MouseAction[T]) DispatchAction() {
	if a.Notify != "" {
		go exec.Command("notify-send", a.Notify).Start()
	}
	if len(a.Run) > 0 {
		go exec.Command(a.Run[0], a.Run[1:]...).Start()
	}
}

func (a *MouseAction[T]) appendInitial(opts any) {
	if len(a.Configs) == 0 {
		return
	}

	tType := reflect.TypeOf(a.Configs).Elem()
	initOpts := reflect.ValueOf(opts)
	if initOpts.Kind() == reflect.Ptr {
		initOpts = initOpts.Elem()
	}

	newCfg := reflect.New(tType).Elem()
	for i := range tType.NumField() {
		fType := tType.Field(i)
		ptrField := newCfg.Field(i)
		actualField := initOpts.FieldByName(fType.Name)
		if !actualField.IsValid() {
			continue
		}

		p := reflect.New(actualField.Type())
		p.Elem().Set(actualField)
		ptrField.Set(p)
	}

	a.Configs = append(a.Configs, newCfg.Interface().(T))
}

func (a *MouseAction[T]) Next(initOpts, liveOpts any) bool {
	if !a.inited && len(a.Configs) > 0 {
		a.inited = true
		a.appendInitial(initOpts)
	}

	if len(a.Configs) < 2 {
		return false
	}

	initV := reflect.ValueOf(initOpts)
	liveV := reflect.ValueOf(liveOpts)
	if initV.Kind() == reflect.Ptr {
		initV = initV.Elem()
	}
	if liveV.Kind() == reflect.Ptr {
		liveV = liveV.Elem()
	}

	clickOptions := reflect.TypeOf(a.Configs).Elem()

	for i := range clickOptions.NumField() {
		name := clickOptions.Field(i).Name
		liveOptsField := liveV.FieldByName(name)
		if liveOptsField.IsValid() && liveOptsField.CanSet() {
			liveOptsField.Set(initV.FieldByName(name))
		}

	}

	partV := reflect.ValueOf(a.Configs[a.clickIndex])
	if partV.Kind() == reflect.Ptr {
		partV = partV.Elem()
	}

	for i := range partV.NumField() {
		f := partV.Field(i)
		if f.Kind() != reflect.Ptr || f.IsNil() {
			continue
		}

		name := clickOptions.Field(i).Name
		liveOptsField := liveV.FieldByName(name)
		if liveOptsField.IsValid() && liveOptsField.CanSet() {
			liveOptsField.Set(f.Elem())
		}
	}

	a.clickIndex = (a.clickIndex + 1) % len(a.Configs)

	return true
}
