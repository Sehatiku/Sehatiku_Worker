package usecase_test

import (
	"context"
	"errors"
	"sehatiku-notification-worker/internal/entity"
	"sehatiku-notification-worker/internal/repository"
	"sehatiku-notification-worker/internal/usecase"
	"testing"
	"time"

	"go.uber.org/zap"
)

type mockTrendRepo struct {
	candidates []repository.TrendCandidate
	err        error
}

func (m *mockTrendRepo) FindTrendCandidates(_ context.Context) ([]repository.TrendCandidate, error) {
	return m.candidates, m.err
}

type mockEscStore struct {
	active      bool
	activeErr   error
	createErr   error
	createCalls int
}

func (m *mockEscStore) ExistsActiveOrRecent(_ context.Context, _, _ string, _ time.Time) (bool, error) {
	return m.active, m.activeErr
}
func (m *mockEscStore) Create(_ context.Context, _ *entity.Escalation) error {
	m.createCalls++
	return m.createErr
}

func newTrendUC(tr *mockTrendRepo, es *mockEscStore, inbox *mockInboxRepo, waswasDays int) *usecase.TrendEscalationUseCase {
	return &usecase.TrendEscalationUseCase{
		TrendRepo:      tr,
		EscalationRepo: es,
		InboxRepo:      inbox,
		WaswasDays:     waswasDays,
		Log:            zap.NewNop(),
	}
}

func candidate(id string, riskyDays int) repository.TrendCandidate {
	return repository.TrendCandidate{
		PatientID: id, FaskesID: "f1", AssignedNakesID: "n1",
		RiskScoreID: "rs-" + id, RiskyDays7d: riskyDays,
	}
}

func TestTrendRun_CreatesWhenSustained(t *testing.T) {
	tr := &mockTrendRepo{candidates: []repository.TrendCandidate{candidate("p1", 3)}}
	es := &mockEscStore{}
	inbox := &mockInboxRepo{}
	uc := newTrendUC(tr, es, inbox, 3)

	s := uc.Run(context.Background())
	if s.Created != 1 {
		t.Errorf("Created = %d; want 1", s.Created)
	}
	if es.createCalls != 1 {
		t.Errorf("escalation Create called %d; want 1", es.createCalls)
	}
	if len(inbox.created) != 1 {
		t.Errorf("inbox rows = %d; want 1", len(inbox.created))
	}
}

func TestTrendRun_SkipsBelowThreshold(t *testing.T) {
	tr := &mockTrendRepo{candidates: []repository.TrendCandidate{candidate("p1", 2)}}
	es := &mockEscStore{}
	uc := newTrendUC(tr, es, &mockInboxRepo{}, 3)

	s := uc.Run(context.Background())
	if s.Created != 0 || s.Skipped != 1 {
		t.Errorf("Created/Skipped = %d/%d; want 0/1", s.Created, s.Skipped)
	}
}

func TestTrendRun_DedupSkips(t *testing.T) {
	tr := &mockTrendRepo{candidates: []repository.TrendCandidate{candidate("p1", 5)}}
	es := &mockEscStore{active: true}
	uc := newTrendUC(tr, es, &mockInboxRepo{}, 3)

	s := uc.Run(context.Background())
	if s.Created != 0 || s.Skipped != 1 {
		t.Errorf("Created/Skipped = %d/%d; want 0/1 (dedup)", s.Created, s.Skipped)
	}
}

func TestTrendRun_PartialFailureContinues(t *testing.T) {
	tr := &mockTrendRepo{candidates: []repository.TrendCandidate{candidate("p1", 4), candidate("p2", 4)}}
	es := &mockEscStore{createErr: errors.New("db down")}
	uc := newTrendUC(tr, es, &mockInboxRepo{}, 3)

	s := uc.Run(context.Background())
	if s.Failed != 2 {
		t.Errorf("Failed = %d; want 2 (both create errors, batch continued)", s.Failed)
	}
}
