package zeroruntime

import (
	"context"
	"testing"
)

// A malformed (nameless) tool call must never reach the agent — it would dispatch
// an empty tool name ("Unknown tool \"\""). Valid calls still pass through.
func TestCollectStreamDropsNamelessToolCalls(t *testing.T) {
	events := make(chan StreamEvent, 8)
	// valid call
	events <- StreamEvent{Type: StreamEventToolCallStart, ToolCallID: "a", ToolName: "read_file"}
	events <- StreamEvent{Type: StreamEventToolCallDelta, ToolCallID: "a", ArgumentsFragment: `{"path":"x"}`}
	events <- StreamEvent{Type: StreamEventToolCallEnd, ToolCallID: "a"}
	// malformed: a delta/end for an id that never got a (named) start
	events <- StreamEvent{Type: StreamEventToolCallDelta, ToolCallID: "b", ArgumentsFragment: `{"path":"y"}`}
	events <- StreamEvent{Type: StreamEventToolCallEnd, ToolCallID: "b"}
	events <- StreamEvent{Type: StreamEventDone}
	close(events)

	got := CollectStream(context.Background(), events)
	if len(got.ToolCalls) != 1 {
		t.Fatalf("expected 1 valid tool call, got %d: %+v", len(got.ToolCalls), got.ToolCalls)
	}
	if got.ToolCalls[0].Name != "read_file" {
		t.Errorf("kept call name = %q, want read_file", got.ToolCalls[0].Name)
	}
	for _, c := range got.ToolCalls {
		if c.Name == "" {
			t.Error("a nameless tool call leaked to the agent")
		}
	}
}

// A provider-signalled dropped (nameless) tool call must be counted so the agent
// can tell the model to retry instead of silently treating it as a final answer.
func TestCollectStreamCountsDroppedToolCalls(t *testing.T) {
	events := make(chan StreamEvent, 4)
	events <- StreamEvent{Type: StreamEventText, Content: "I'll write the file."}
	events <- StreamEvent{Type: StreamEventToolCallDropped}
	events <- StreamEvent{Type: StreamEventDone}
	close(events)

	got := CollectStream(context.Background(), events)
	if len(got.ToolCalls) != 0 {
		t.Fatalf("expected no usable tool calls, got %d", len(got.ToolCalls))
	}
	if got.DroppedToolCalls != 1 {
		t.Fatalf("expected DroppedToolCalls=1, got %d", got.DroppedToolCalls)
	}
}
