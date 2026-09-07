package frpcclient

import (
	"context"
	"time"

	"cnb.cool/sixkun/sixfrp/v2/cmd/frppc/facades"
	"cnb.cool/sixkun/sixfrp/v2/haokun/defs"
	"cnb.cool/sixkun/sixfrp/v2/utils/systemevent"

	pb "cnb.cool/sixkun/sixfrp-client-proto/sixfrp/client_pb"
)

// PushEvents reports operational system events (started/stopped) to master for
// the admin event center. It attaches the client identity for authentication
// and is a no-op when the runtime is not running or the batch is empty.
// Best-effort: callers ignore the returned error beyond logging.
func (s *Service) PushEvents(ctx context.Context, events []*pb.SystemEventPayload) error {
	if !s.ShouldRun() {
		return nil
	}
	if len(events) == 0 {
		return nil
	}

	_, err := s.masterClient.PushClientEvent(ctx, &pb.PushClientEventReq{
		Base: &pb.ClientBase{
			ClientId:     s.opts.clientID,
			ClientSecret: s.opts.clientSecret,
		},
		Events: events,
	})
	return err
}

// events returns the per-service debouncer for client lifecycle events. The
// first started/stopped emits immediately; rapid same-key repeats (a
// re-register loop) collapse into one aggregated trailing push.
func (s *Service) events() *systemevent.Debouncer[*pb.SystemEventPayload] {
	s.eventDebouncerOnce.Do(func() {
		s.eventDebouncer = systemevent.New(defs.SystemEventDebounceWindow, s.pushDebouncedEvent)
	})
	return s.eventDebouncer
}

// pushDebouncedEvent is the debouncer's emit sink: annotate with the aggregate
// count and push to master on a short-lived context.
func (s *Service) pushDebouncedEvent(event *pb.SystemEventPayload, count int) {
	event = withAggregateCount(event, count)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.PushEvents(ctx, []*pb.SystemEventPayload{event}); err != nil {
		facades.Log().Warningf("emit lifecycle event %s failed: %v", event.GetEventType(), err)
	}
}

// emitLifecycleEvent submits a client lifecycle event to the debouncer.
func (s *Service) emitLifecycleEvent(eventType, level string) {
	if !s.ShouldRun() {
		return
	}
	s.events().Submit(&pb.SystemEventPayload{
		EventType:   eventType,
		Level:       level,
		SubjectKind: defs.SystemEventSubjectClient,
		SubjectId:   s.opts.clientID,
		DedupeKey:   eventType + ":" + s.opts.clientID,
	})
}
