package usecase_test

import (
	"context"
	"errors"
	"sehatiku-notification-worker/internal/entity"
	"sehatiku-notification-worker/internal/usecase"
	"testing"

	"go.uber.org/zap"
)

// ── Mocks ──────────────────────────────────────────────────────────────────

type mockPatientRepo struct {
	patients []entity.Patient
	err      error
}

func (m *mockPatientRepo) FindPatientsWithNoLogToday(_ context.Context) ([]entity.Patient, error) {
	return m.patients, m.err
}

type mockNotifRepo struct {
	existsResult bool
	existsErr    error
	createErr    error
	updates      []struct{ id, status string }
}

func (m *mockNotifRepo) ExistsForToday(_ context.Context, _ string) (bool, error) {
	return m.existsResult, m.existsErr
}

func (m *mockNotifRepo) Create(_ context.Context, _ *entity.Notification) error {
	return m.createErr
}

func (m *mockNotifRepo) UpdateStatus(_ context.Context, id, status string, _ *string) error {
	m.updates = append(m.updates, struct{ id, status string }{id, status})
	return nil
}

type mockSender struct {
	patientErr     error
	companionErr   error
	patientCalls   []string
	companionCalls []string
}

func (m *mockSender) SendPatientReminder(_ context.Context, phone, _, _ string) error {
	m.patientCalls = append(m.patientCalls, phone)
	return m.patientErr
}

func (m *mockSender) SendCompanionReminder(_ context.Context, phone, _, _ string) error {
	m.companionCalls = append(m.companionCalls, phone)
	return m.companionErr
}

type statefulMockSender struct {
	failPhone    string
	patientCalls []string
}

func (m *statefulMockSender) SendPatientReminder(_ context.Context, phone, _, _ string) error {
	m.patientCalls = append(m.patientCalls, phone)
	if phone == m.failPhone {
		return errors.New("send failed")
	}
	return nil
}

func (m *statefulMockSender) SendCompanionReminder(_ context.Context, _, _, _ string) error {
	return nil
}

func newUC(pr *mockPatientRepo, nr *mockNotifRepo, wa usecase.WhatsAppSender) *usecase.DailyReminderUseCase {
	return &usecase.DailyReminderUseCase{
		PatientRepo:      pr,
		NotificationRepo: nr,
		WhatsApp:         wa,
		Log:              zap.NewNop(),
	}
}

// ── Tests ──────────────────────────────────────────────────────────────────

func TestRun_SendsToPatient(t *testing.T) {
	patients := []entity.Patient{{ID: "p1", FullName: "Budi", PhoneNumber: "628111", Status: "active"}}
	wa := &mockSender{}
	uc := newUC(&mockPatientRepo{patients: patients}, &mockNotifRepo{}, wa)

	summary := uc.Run(context.Background(), "noon")

	if summary.Sent != 1 {
		t.Errorf("Sent = %d; want 1", summary.Sent)
	}
	if len(wa.patientCalls) != 1 || wa.patientCalls[0] != "628111" {
		t.Errorf("patientCalls = %v; want [628111]", wa.patientCalls)
	}
}

func TestRun_SendsToCompanionWhenPhoneSet(t *testing.T) {
	cp, cn := "628222", "Ibu Budi"
	patients := []entity.Patient{{
		ID: "p1", FullName: "Budi", PhoneNumber: "628111",
		CompanionPhone: &cp, CompanionName: &cn, Status: "active",
	}}
	wa := &mockSender{}
	uc := newUC(&mockPatientRepo{patients: patients}, &mockNotifRepo{}, wa)

	uc.Run(context.Background(), "noon")

	if len(wa.companionCalls) != 1 || wa.companionCalls[0] != cp {
		t.Errorf("companionCalls = %v; want [%s]", wa.companionCalls, cp)
	}
}

func TestRun_SkipsCompanionWhenPhoneNil(t *testing.T) {
	patients := []entity.Patient{{ID: "p1", FullName: "Budi", PhoneNumber: "628111", Status: "active"}}
	wa := &mockSender{}
	uc := newUC(&mockPatientRepo{patients: patients}, &mockNotifRepo{}, wa)

	uc.Run(context.Background(), "noon")

	if len(wa.companionCalls) != 0 {
		t.Errorf("expected no companion calls, got %v", wa.companionCalls)
	}
}

func TestRun_SkipsIfAlreadySentToday(t *testing.T) {
	patients := []entity.Patient{{ID: "p1", FullName: "Budi", PhoneNumber: "628111", Status: "active"}}
	wa := &mockSender{}
	uc := newUC(&mockPatientRepo{patients: patients}, &mockNotifRepo{existsResult: true}, wa)

	summary := uc.Run(context.Background(), "noon")

	if summary.SkippedAlreadySent != 1 {
		t.Errorf("SkippedAlreadySent = %d; want 1", summary.SkippedAlreadySent)
	}
	if len(wa.patientCalls) != 0 {
		t.Errorf("expected no WA calls when already sent, got %v", wa.patientCalls)
	}
}

func TestRun_MarksFailedOnWAError(t *testing.T) {
	patients := []entity.Patient{{ID: "p1", FullName: "Budi", PhoneNumber: "628111", Status: "active"}}
	wa := &mockSender{patientErr: errors.New("not connected")}
	nr := &mockNotifRepo{}
	uc := newUC(&mockPatientRepo{patients: patients}, nr, wa)

	summary := uc.Run(context.Background(), "noon")

	if summary.Failed != 1 {
		t.Errorf("Failed = %d; want 1", summary.Failed)
	}
	hasFailed := false
	for _, u := range nr.updates {
		if u.status == "failed" {
			hasFailed = true
		}
	}
	if !hasFailed {
		t.Error("expected UpdateStatus('failed') to be called")
	}
}

func TestRun_ContinuesAfterOnePatientFailure(t *testing.T) {
	cp := "628333"
	patients := []entity.Patient{
		{ID: "p1", FullName: "Budi", PhoneNumber: "628111", Status: "active"},
		{ID: "p2", FullName: "Siti", PhoneNumber: "628222", Status: "active",
			CompanionPhone: &cp},
	}
	wa := &statefulMockSender{failPhone: "628111"}
	uc := newUC(&mockPatientRepo{patients: patients}, &mockNotifRepo{}, wa)

	summary := uc.Run(context.Background(), "noon")

	if summary.Sent != 1 {
		t.Errorf("Sent = %d; want 1 (only second patient)", summary.Sent)
	}
	if summary.Failed != 1 {
		t.Errorf("Failed = %d; want 1 (first patient)", summary.Failed)
	}
	if len(wa.patientCalls) != 2 {
		t.Errorf("expected 2 WA patient calls (both attempted), got %d", len(wa.patientCalls))
	}
}
