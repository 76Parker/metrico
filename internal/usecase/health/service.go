package health

import "context"

type AvailabilityResult struct {
	// Ready показывает доступна ли система в целом принимать трафик на основании результатов проверок
	Ready bool
	// AvailabilityResults содержит результаты проверок доступности внешних систем
	AvailabilityResults []CheckResult
}

type CheckResult struct {
	// System содержит название системы, которая была проверена (например, "postgres")
	System string
	// IsAvailable показывает доступность конкретной внешней системы (в нашем случае это только база данных)
	IsAvailable bool
	// Error содержит информацию об ошибке, если проверка не была успешной
	Error error
}

type Service struct {
	db dbPinger
}

func NewService(db dbPinger) *Service {
	return &Service{db: db}
}

func (s *Service) CheckAvailability(ctx context.Context) AvailabilityResult {
	err := s.db.Ping(ctx)
	result := AvailabilityResult{
		Ready:               err == nil,
		AvailabilityResults: []CheckResult{{System: "postgres", IsAvailable: err == nil, Error: err}},
	}
	return result
}
