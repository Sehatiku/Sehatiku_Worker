package repository

import (
	"context"

	"gorm.io/gorm"
)

// TrendCandidate adalah satu pasien aktif yang skor TERBARU-nya 'waswas', beserta jumlah
// hari berisiko (waswas/bahaya) dalam 7 hari terakhir (WIB). Usecase memutuskan apakah
// jumlah itu memenuhi ambang "sustained".
type TrendCandidate struct {
	PatientID       string `gorm:"column:patient_id"`
	FaskesID        string `gorm:"column:faskes_id"`
	AssignedNakesID string `gorm:"column:assigned_nakes_id"`
	RiskScoreID     string `gorm:"column:risk_score_id"`
	RiskyDays7d     int    `gorm:"column:risky_days_7d"`
}

type TrendQuerier interface {
	FindTrendCandidates(ctx context.Context) ([]TrendCandidate, error)
}

type TrendRepository struct {
	DB *gorm.DB
}

// FindTrendCandidates membaca patients + risk_scores (tabel bersama dgn backend). Tidak
// memanggil ML — hanya membaca skor yang sudah ada (worker tetap ringan).
func (r *TrendRepository) FindTrendCandidates(ctx context.Context) ([]TrendCandidate, error) {
	var rows []TrendCandidate
	err := r.DB.WithContext(ctx).Raw(`
		WITH latest_rs AS (
			SELECT DISTINCT ON (rs.patient_id)
				rs.patient_id, rs.id AS risk_score_id, rs.status
			FROM risk_scores rs
			ORDER BY rs.patient_id, rs.scored_at DESC
		),
		risky AS (
			SELECT patient_id,
				COUNT(DISTINCT (scored_at AT TIME ZONE 'Asia/Jakarta')::date) AS risky_days
			FROM risk_scores
			WHERE status IN ('waswas','bahaya')
			  AND scored_at >= now() - interval '7 days'
			GROUP BY patient_id
		)
		SELECT
			p.id                AS patient_id,
			p.faskes_id         AS faskes_id,
			p.assigned_nakes_id AS assigned_nakes_id,
			lr.risk_score_id    AS risk_score_id,
			COALESCE(rk.risky_days, 0) AS risky_days_7d
		FROM patients p
		JOIN latest_rs lr ON lr.patient_id = p.id
		LEFT JOIN risky rk ON rk.patient_id = p.id
		WHERE p.status = 'active' AND lr.status = 'waswas'
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	return rows, nil
}
