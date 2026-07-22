package portfolio

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

func (r *PostgresRepository) GetPlatformsWithStats(ctx context.Context, userID uuid.UUID) ([]PlatformStats, error) {
	rows, err := r.db.Query(ctx, `
		SELECT
			is_.id,
			is_.name,
			COALESCE(is_.description, ''),
			is_.source_type,
			is_.is_active,
			is_.created_at,
			is_.updated_at,
			COUNT(pe.id)::bigint AS investments,
			COALESCE(SUM(pe.quantity::numeric * pe.price::numeric), 0)::text AS total_value
		FROM investment_sources is_
		LEFT JOIN portfolio_entries pe ON pe.source_id = is_.id
		WHERE is_.user_id = $1
		GROUP BY is_.id
		ORDER BY is_.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]PlatformStats, 0)
	for rows.Next() {
		var p PlatformStats
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Description, &p.SourceType,
			&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
			&p.Investments, &p.TotalValue,
		); err != nil {
			return nil, err
		}
		result = append(result, p)
	}
	return result, nil
}

func (r *PostgresRepository) UpdatePlatform(ctx context.Context, userID, sourceID uuid.UUID, name, description string, sourceType SourceType, isActive bool) (PlatformStats, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE investment_sources
		SET name = $3, description = $4, source_type = $5, is_active = $6, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`, sourceID, userID, name, description, sourceType, isActive)
	if err != nil {
		return PlatformStats{}, err
	}
	if tag.RowsAffected() == 0 {
		return PlatformStats{}, errors.New("platform not found")
	}

	var p PlatformStats
	if err := r.db.QueryRow(ctx, `
		SELECT
			is_.id, is_.name, COALESCE(is_.description, ''),
			is_.source_type, is_.is_active, is_.created_at, is_.updated_at,
			COUNT(pe.id)::bigint,
			COALESCE(SUM(pe.quantity::numeric * pe.price::numeric), 0)::text
		FROM investment_sources is_
		LEFT JOIN portfolio_entries pe ON pe.source_id = is_.id
		WHERE is_.id = $1 AND is_.user_id = $2
		GROUP BY is_.id
	`, sourceID, userID).Scan(
		&p.ID, &p.Name, &p.Description, &p.SourceType,
		&p.IsActive, &p.CreatedAt, &p.UpdatedAt,
		&p.Investments, &p.TotalValue,
	); err != nil {
		return PlatformStats{}, err
	}
	return p, nil
}

func (r *PostgresRepository) DeletePlatform(ctx context.Context, userID, sourceID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `
		DELETE FROM investment_sources WHERE id = $1 AND user_id = $2
	`, sourceID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("platform not found")
	}
	return nil
}

func (r *PostgresRepository) CreatePlatform(ctx context.Context, userID uuid.UUID, sourceType SourceType, name, desciption string) (InvestmentSource, error) {
	var platform InvestmentSource
	err := r.db.QueryRow(ctx, "INSERT INTO investment_sources(user_id, source_type, name, description) VALUES ($1, $2, $3, $4) RETURNING id, name, description, created_at, updated_at", userID, sourceType, name, desciption).Scan(&platform.ID, &platform.Name, &platform.Description, &platform.CreatedAt, &platform.UpdatedAt)
	if err != nil {
		return InvestmentSource{}, err
	}

	return platform, nil
}
