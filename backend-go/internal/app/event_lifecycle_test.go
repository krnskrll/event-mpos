package app_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krnskrll/event-mpos/internal/event"
	"github.com/krnskrll/event-mpos/internal/order"
	"github.com/krnskrll/event-mpos/internal/testutil"
)

// Ported from the Kotlin EventLifecycleTest, plus the "selling only while ACTIVE" rule.

func TestEventTransitions(t *testing.T) {
	legal := [][2]event.EventStatus{
		{event.StatusDraft, event.StatusActive},
		{event.StatusDraft, event.StatusClosed},
		{event.StatusActive, event.StatusClosed},
	}
	for _, tr := range legal {
		if err := tr[0].AssertCanTransitionTo(tr[1]); err != nil {
			t.Errorf("%s -> %s rejected: %v", tr[0], tr[1], err)
		}
	}
	illegal := [][2]event.EventStatus{
		{event.StatusClosed, event.StatusActive},
		{event.StatusClosed, event.StatusDraft},
		{event.StatusActive, event.StatusDraft},
		{event.StatusDraft, event.StatusDraft},
		{event.StatusActive, event.StatusActive},
	}
	for _, tr := range illegal {
		err := tr[0].AssertCanTransitionTo(tr[1])
		if testutil.StatusOf(err) != http.StatusConflict {
			t.Errorf("%s -> %s: got %v, want 409", tr[0], tr[1], err)
		}
	}
}

func TestSellingOnlyWhileActive(t *testing.T) {
	a := testutil.NewApp(t)
	p := a.Product("100.00")

	for _, status := range []event.EventStatus{event.StatusDraft, event.StatusClosed} {
		ev := &event.Event{Code: testutil.Unique("T"), Name: "Not selling", Status: status}
		if err := a.DB.Create(ev).Error; err != nil {
			t.Fatal(err)
		}
		a.Allocate(ev.ID, p.ID, 10)
		_, err := a.Orders.Submit(ev.ID, uuid.New(), order.SubmitOrderRequest{
			IdempotencyKey: uuid.NewString(), Customer: order.CustomerInfo{Name: "X"},
			Lines: []order.OrderLineRequest{{ProductID: p.ID, Qty: 1}},
		})
		if testutil.StatusOf(err) != http.StatusConflict {
			t.Errorf("submit to %s event: got %v, want 409", status, err)
		}
		if got := a.Remaining(ev.ID, p.ID); got != 10 {
			t.Errorf("%s event stock changed: remaining %d, want 10", status, got)
		}
	}

	// The same order on an ACTIVE event goes through.
	ev := a.ActiveEvent()
	a.Allocate(ev.ID, p.ID, 10)
	if _, err := a.Orders.Submit(ev.ID, uuid.New(), order.SubmitOrderRequest{
		IdempotencyKey: uuid.NewString(), Customer: order.CustomerInfo{Name: "X"},
		Lines: []order.OrderLineRequest{{ProductID: p.ID, Qty: 1}},
	}); err != nil {
		t.Fatalf("submit to ACTIVE event: %v", err)
	}
}

func TestEventStatusEndpointRejectsIllegalTransition(t *testing.T) {
	a := testutil.NewApp(t)
	created, err := a.Events.CreateEvent(event.CreateEventRequest{Code: testutil.Unique("T"), Name: "Lifecycle"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.Events.ChangeStatus(created.ID, event.StatusActive); err != nil {
		t.Fatalf("DRAFT -> ACTIVE: %v", err)
	}
	if _, err := a.Events.ChangeStatus(created.ID, event.StatusClosed); err != nil {
		t.Fatalf("ACTIVE -> CLOSED: %v", err)
	}
	if _, err := a.Events.ChangeStatus(created.ID, event.StatusActive); testutil.StatusOf(err) != http.StatusConflict {
		t.Errorf("CLOSED -> ACTIVE: got %v, want 409", err)
	}
}
