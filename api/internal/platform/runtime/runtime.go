package runtime

import "context"

type Shutdowner interface {
	Shutdown(context.Context) error
}

type AdmissionGate interface {
	Pause()
}

type Service struct {
	server    Shutdowner
	admission AdmissionGate
	drainers  []Shutdowner
}

func New(server Shutdowner, admission AdmissionGate, drainers ...Shutdowner) Service {
	return Service{server: server, admission: admission, drainers: drainers}
}

func (s Service) Shutdown(ctx context.Context) error {
	s.admission.Pause()
	if err := s.server.Shutdown(ctx); err != nil {
		return err
	}
	for _, drainer := range s.drainers {
		if err := drainer.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
