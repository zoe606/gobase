package grpcserver

import "context"

func (s *Server) Shutdown(ctx context.Context) error {
	s.shutdownOnce.Do(func() {
		s.mu.Lock()
		s.stopping = true
		s.mu.Unlock()
		s.health.Shutdown()
		go func() { s.GRPC.GracefulStop(); close(s.shutdownDone) }()
	})
	select {
	case <-s.shutdownDone:
		return nil
	case <-ctx.Done():
		// Stop can wait on GracefulStop's lock while a handler ignores cancellation.
		s.forceOnce.Do(func() { go s.GRPC.Stop() })
		return ctx.Err()
	}
}
