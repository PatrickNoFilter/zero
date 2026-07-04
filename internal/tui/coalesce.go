package tui

import (
	"sync"
	"time"

	tea "charm.land/bubbletea/v2"
)

// streamCoalesceInterval is roughly one 60fps frame. Assistant-text deltas that
// arrive within this window are merged into a single agentTextMsg, so the render
// rate decouples from the token rate: a fast local model (100+ tok/s) no longer
// forces 100+ full Update→View cycles (each re-parsing the growing markdown) per
// second. Rendering stays smooth regardless of provider speed.
const streamCoalesceInterval = 16 * time.Millisecond

// textCoalescer batches agentTextMsg deltas before forwarding them to the Bubble
// Tea program. Any OTHER message flushes the pending text first, so ordering
// between streamed prose and tool-call / reasoning / row / usage messages is
// preserved. The turn's final agentResponseMsg does not pass through here (it is
// a tea.Cmd return, not a sink message), but the model drops deltas whose runID
// is no longer active, so a flush that races just past end-of-turn is harmless.
//
// Sink messages originate from the single agent goroutine and so arrive
// serially; the only concurrent caller is the flush timer. The mutex guards the
// buffer/timer against that one race.
type textCoalescer struct {
	forward func(tea.Msg) // downstream sink (external sink + program.Send)

	mu    sync.Mutex
	buf   []byte
	runID int
	timer *time.Timer
}

func newTextCoalescer(forward func(tea.Msg)) *textCoalescer {
	return &textCoalescer{forward: forward}
}

// send is the coalescing entry point installed as the RuntimeMessageSink.
func (c *textCoalescer) send(msg tea.Msg) {
	text, ok := msg.(agentTextMsg)
	if !ok {
		// Non-text message: flush buffered text first (preserving order), then
		// forward it unchanged.
		c.flush()
		c.forward(msg)
		return
	}

	c.mu.Lock()
	// A delta for a different run than the one buffered: flush the old run's text
	// before buffering the new run's. In practice runs are sequential (the prior
	// run's end already flushed via a non-text message), so this is belt-and-braces.
	if len(c.buf) > 0 && text.runID != c.runID {
		pending := c.drainLocked()
		c.mu.Unlock()
		c.forward(pending)
		c.mu.Lock()
	}
	c.runID = text.runID
	c.buf = append(c.buf, text.delta...)
	if c.timer == nil {
		c.timer = time.AfterFunc(streamCoalesceInterval, c.flush)
	}
	c.mu.Unlock()
}

// flush forwards any buffered text as one agentTextMsg. Safe to call from the
// timer goroutine and inline; a no-op when nothing is buffered.
func (c *textCoalescer) flush() {
	c.mu.Lock()
	if len(c.buf) == 0 {
		c.mu.Unlock()
		return
	}
	msg := c.drainLocked()
	c.mu.Unlock()
	c.forward(msg)
}

// drainLocked packages the buffer into an agentTextMsg and stops the timer. The
// caller holds c.mu. string(c.buf) copies, so reusing the backing array via [:0]
// is safe.
func (c *textCoalescer) drainLocked() agentTextMsg {
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}
	msg := agentTextMsg{runID: c.runID, delta: string(c.buf)}
	c.buf = c.buf[:0]
	return msg
}
