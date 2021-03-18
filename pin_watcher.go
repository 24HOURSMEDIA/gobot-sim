package gobot_sim

type WatchFuncs struct {
	Read    PinReadFunc
	Changed PinChangedFunc
}

type PinWatcher struct {
	name       string
	pin        string
	onValue    byte
	offValue   byte
	watchFuncs *WatchFuncs
	previous   int
	triggering bool
}

// Name returns the name of the watcher and can be used in
// change handlers, for logging etc
func (w *PinWatcher) Name() string {
	return w.name
}

// SetName sets the name of the watcher and can be used in
// change handlers, for logging etc
func (w *PinWatcher) SetName(name string) {
	w.name = name
}

// NewPinWatcher creates a new watcher for value changes of
// GPIO pins
func NewPinWatcher(pin string, watchFuncs *WatchFuncs) *PinWatcher {
	w := &PinWatcher{
		pin:        pin,
		onValue:    PIN_ON,
		offValue:   PIN_OFF,
		watchFuncs: watchFuncs,
	}
	return w
}

// Pin returns the pin number
func (w *PinWatcher) Pin() string {
	return w.pin
}

// Observe must be called periodically by the owner
// and detects changes in state.
func (w *PinWatcher) Observe() error {
	v, err := w.watchFuncs.Read(w.pin)
	if err != nil {
		return err
	}
	if w.triggering {
		if v != w.previous && w.watchFuncs.Changed != nil {
			ev := PinChangedEvent{
				Pin:       w.pin,
				LastValue: w.previous,
				Value:     v,
				source:    w,
			}
			w.watchFuncs.Changed(ev)
		}
	} else {
		w.triggering = true
	}
	w.previous = v
	return nil
}
