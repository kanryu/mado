package mado

import "sync"

// QueueBackend is a small backend utility for adapters. SDL3 and Winit
// implementations can translate native events into mado.Event and enqueue them.
type QueueBackend struct {
	kind   BackendKind
	mu     sync.Mutex
	cond   *sync.Cond
	events []Event
	quit   bool
	nextID WindowID
}

func NewQueueBackend(kind BackendKind) *QueueBackend {
	b := &QueueBackend{kind: kind}
	b.cond = sync.NewCond(&b.mu)
	return b
}
func (b *QueueBackend) Name() BackendKind { return b.kind }
func (b *QueueBackend) CreateWindow(options WindowOptions) (Window, error) {
	if options.Size.Width <= 0 || options.Size.Height <= 0 {
		return nil, ErrInvalidSize
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.nextID++
	return &StubWindow{
		IDValue:      b.nextID,
		TitleValue:   options.Title,
		SizeValue:    options.Size,
		Handle:       NativeHandle{Kind: HandleWin32HWND},
		SessionValue: NewSession(b.kind),
	}, nil
}
func (b *QueueBackend) Push(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.quit {
		b.events = append(b.events, event)
		b.cond.Signal()
	}
}
func (b *QueueBackend) Quit() {
	b.mu.Lock()
	b.quit = true
	b.cond.Broadcast()
	b.mu.Unlock()
}
func (b *QueueBackend) Run(handle func(Event) error) error {
	for {
		b.mu.Lock()
		for len(b.events) == 0 && !b.quit {
			b.cond.Wait()
		}
		if len(b.events) == 0 && b.quit {
			b.mu.Unlock()
			return nil
		}
		e := b.events[0]
		b.events = b.events[1:]
		b.mu.Unlock()
		if err := handle(e); err != nil {
			if err == ErrClosed {
				return nil
			}
			return err
		}
	}
}
