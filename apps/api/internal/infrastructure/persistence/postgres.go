package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/safescope/safescope/apps/api/internal/domain"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, databaseURL string) (*Postgres, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) Ping(ctx context.Context) error {
	return p.pool.Ping(ctx)
}

func (p *Postgres) Create(ctx context.Context, value any) error {
	switch v := value.(type) {
	case *domain.User:
		_, err := p.pool.Exec(ctx, `
			INSERT INTO users (id, email, name, password_hash, role, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			v.ID, v.Email, v.Name, v.PasswordHash, v.Role, v.CreatedAt, v.UpdatedAt)
		return translate(err)
	case *domain.Project:
		_, err := p.pool.Exec(ctx, `
			INSERT INTO projects (id, name, description, status, owner_id, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7)`,
			v.ID, v.Name, v.Description, v.Status, v.OwnerID, v.CreatedAt, v.UpdatedAt)
		return translate(err)
	case *domain.Asset:
		_, err := p.pool.Exec(ctx, `
			INSERT INTO assets (id, project_id, type, value, status, tags, metadata, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
			v.ID, v.ProjectID, v.Type, v.Value, v.Status, v.Tags, v.Metadata, v.CreatedAt, v.UpdatedAt)
		return translate(err)
	default:
		return fmt.Errorf("unsupported create type %T", value)
	}
}

func (p *Postgres) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT id, email, name, password_hash, role, created_at, updated_at
		FROM users WHERE email = $1`, email)
	return scanUser(row)
}

func (p *Postgres) ByID(ctx context.Context, id string) (*domain.User, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT id, email, name, password_hash, role, created_at, updated_at
		FROM users WHERE id = $1`, id)
	return scanUser(row)
}

func (p *Postgres) Count(ctx context.Context) (int64, error) {
	var count int64
	err := p.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (p *Postgres) ListUsers(ctx context.Context, page domain.Page) ([]domain.User, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, email, name, password_hash, role, created_at, updated_at
		FROM users ORDER BY created_at ASC LIMIT $1 OFFSET $2`, page.Limit, page.Offset)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	users := make([]domain.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *user)
	}
	return users, rows.Err()
}

func (p *Postgres) UpdateUserRole(ctx context.Context, id string, role domain.Role) error {
	tag, err := p.pool.Exec(ctx, `UPDATE users SET role=$2, updated_at=NOW() WHERE id=$1`, id, role)
	return ensureAffected(tag, err)
}

func (p *Postgres) List(ctx context.Context, actor *domain.User, page domain.Page) ([]domain.Project, error) {
	query := `
		SELECT p.id, p.name, p.description, p.status, p.owner_id, p.created_at, p.updated_at, COUNT(a.id)
		FROM projects p
		LEFT JOIN assets a ON a.project_id = p.id`
	args := []any{}
	if actor.Role != domain.RoleAdmin {
		query += ` WHERE p.owner_id = $1`
		args = append(args, actor.ID)
	}
	query += ` GROUP BY p.id ORDER BY p.updated_at DESC`
	args = append(args, page.Limit, page.Offset)
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	projects := make([]domain.Project, 0)
	for rows.Next() {
		var project domain.Project
		if err := rows.Scan(&project.ID, &project.Name, &project.Description, &project.Status, &project.OwnerID, &project.CreatedAt, &project.UpdatedAt, &project.AssetCount); err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func (p *Postgres) ProjectByID(ctx context.Context, id string) (*domain.Project, error) {
	var project domain.Project
	err := p.pool.QueryRow(ctx, `
		SELECT p.id, p.name, p.description, p.status, p.owner_id, p.created_at, p.updated_at, COUNT(a.id)
		FROM projects p LEFT JOIN assets a ON a.project_id = p.id
		WHERE p.id = $1 GROUP BY p.id`, id).
		Scan(&project.ID, &project.Name, &project.Description, &project.Status, &project.OwnerID, &project.CreatedAt, &project.UpdatedAt, &project.AssetCount)
	if err != nil {
		return nil, translate(err)
	}
	return &project, nil
}

func (p *Postgres) UpdateProject(ctx context.Context, project *domain.Project) error {
	tag, err := p.pool.Exec(ctx, `
		UPDATE projects SET name=$2, description=$3, status=$4, updated_at=$5 WHERE id=$1`,
		project.ID, project.Name, project.Description, project.Status, project.UpdatedAt)
	return ensureAffected(tag, err)
}

func (p *Postgres) DeleteProject(ctx context.Context, id string) error {
	tag, err := p.pool.Exec(ctx, `DELETE FROM projects WHERE id=$1`, id)
	return ensureAffected(tag, err)
}

func (p *Postgres) CanAccess(ctx context.Context, actor *domain.User, projectID string) (bool, error) {
	if actor.Role == domain.RoleAdmin {
		var exists bool
		err := p.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE id=$1)`, projectID).Scan(&exists)
		return exists, err
	}
	var exists bool
	err := p.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE id=$1 AND owner_id=$2)`, projectID, actor.ID).Scan(&exists)
	return exists, err
}

func (p *Postgres) ListByProject(ctx context.Context, projectID string, page domain.Page) ([]domain.Asset, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT id, project_id, type, value, status, tags, metadata, created_at, updated_at
		FROM assets WHERE project_id=$1 ORDER BY updated_at DESC LIMIT $2 OFFSET $3`,
		projectID, page.Limit, page.Offset)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	assets := make([]domain.Asset, 0)
	for rows.Next() {
		asset, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, *asset)
	}
	return assets, rows.Err()
}

func (p *Postgres) AssetByID(ctx context.Context, id string) (*domain.Asset, error) {
	row := p.pool.QueryRow(ctx, `
		SELECT id, project_id, type, value, status, tags, metadata, created_at, updated_at
		FROM assets WHERE id=$1`, id)
	return scanAsset(row)
}

func (p *Postgres) UpdateAsset(ctx context.Context, asset *domain.Asset) error {
	tag, err := p.pool.Exec(ctx, `
		UPDATE assets SET status=$2, tags=$3, metadata=$4, updated_at=$5 WHERE id=$1`,
		asset.ID, asset.Status, asset.Tags, asset.Metadata, asset.UpdatedAt)
	return ensureAffected(tag, err)
}

func (p *Postgres) DeleteAsset(ctx context.Context, id string) error {
	tag, err := p.pool.Exec(ctx, `DELETE FROM assets WHERE id=$1`, id)
	return ensureAffected(tag, err)
}

func (p *Postgres) Summary(ctx context.Context, actor *domain.User) (*domain.Dashboard, error) {
	filter, args := "", []any{}
	if actor.Role != domain.RoleAdmin {
		filter = " WHERE p.owner_id=$1"
		args = append(args, actor.ID)
	}
	var dashboard domain.Dashboard
	dashboard.AssetsByType = make([]domain.Metric, 0)
	err := p.pool.QueryRow(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE p.status='archived')
		FROM projects p`+filter, args...).Scan(&dashboard.Projects, &dashboard.Archived)
	if err != nil {
		return nil, err
	}
	assetFilter := ""
	if actor.Role != domain.RoleAdmin {
		assetFilter = " WHERE p.owner_id=$1"
	}
	err = p.pool.QueryRow(ctx, `
		SELECT COUNT(a.id), COUNT(a.id) FILTER (WHERE a.status='alive')
		FROM assets a JOIN projects p ON p.id=a.project_id`+assetFilter, args...).Scan(&dashboard.Assets, &dashboard.AliveAssets)
	if err != nil {
		return nil, err
	}
	rows, err := p.pool.Query(ctx, `
		SELECT a.type, COUNT(a.id) FROM assets a JOIN projects p ON p.id=a.project_id`+
		assetFilter+` GROUP BY a.type ORDER BY COUNT(a.id) DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var metric domain.Metric
		if err := rows.Scan(&metric.Label, &metric.Value); err != nil {
			return nil, err
		}
		dashboard.AssetsByType = append(dashboard.AssetsByType, metric)
	}
	recent, err := p.recentAssets(ctx, actor, 6)
	if err != nil {
		return nil, err
	}
	dashboard.RecentAssets = recent
	return &dashboard, nil
}

func (p *Postgres) recentAssets(ctx context.Context, actor *domain.User, limit int) ([]domain.Asset, error) {
	query := `
		SELECT a.id, a.project_id, a.type, a.value, a.status, a.tags, a.metadata, a.created_at, a.updated_at
		FROM assets a JOIN projects p ON p.id=a.project_id`
	args := []any{}
	if actor.Role != domain.RoleAdmin {
		query += ` WHERE p.owner_id=$1`
		args = append(args, actor.ID)
	}
	args = append(args, limit)
	query += fmt.Sprintf(" ORDER BY a.updated_at DESC LIMIT $%d", len(args))
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assets := make([]domain.Asset, 0)
	for rows.Next() {
		asset, err := scanAsset(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, *asset)
	}
	return assets, rows.Err()
}

type scanner interface {
	Scan(...any) error
}

func scanUser(row scanner) (*domain.User, error) {
	var user domain.User
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, translate(err)
	}
	return &user, nil
}

func scanAsset(row scanner) (*domain.Asset, error) {
	var asset domain.Asset
	var metadata []byte
	err := row.Scan(&asset.ID, &asset.ProjectID, &asset.Type, &asset.Value, &asset.Status, &asset.Tags, &metadata, &asset.CreatedAt, &asset.UpdatedAt)
	if err != nil {
		return nil, translate(err)
	}
	asset.Metadata = json.RawMessage(metadata)
	return &asset, nil
}

func translate(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return domain.ErrConflict
	}
	return err
}

func ensureAffected(tag pgconn.CommandTag, err error) error {
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

type ProjectRepository struct{ DB *Postgres }

func (r ProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	return r.DB.Create(ctx, p)
}
func (r ProjectRepository) List(ctx context.Context, u *domain.User, page domain.Page) ([]domain.Project, error) {
	return r.DB.List(ctx, u, page)
}
func (r ProjectRepository) ByID(ctx context.Context, id string) (*domain.Project, error) {
	return r.DB.ProjectByID(ctx, id)
}
func (r ProjectRepository) Update(ctx context.Context, p *domain.Project) error {
	return r.DB.UpdateProject(ctx, p)
}
func (r ProjectRepository) Delete(ctx context.Context, id string) error {
	return r.DB.DeleteProject(ctx, id)
}
func (r ProjectRepository) CanAccess(ctx context.Context, u *domain.User, id string) (bool, error) {
	return r.DB.CanAccess(ctx, u, id)
}

type AssetRepository struct{ DB *Postgres }

func (r AssetRepository) Create(ctx context.Context, a *domain.Asset) error {
	return r.DB.Create(ctx, a)
}
func (r AssetRepository) ListByProject(ctx context.Context, id string, page domain.Page) ([]domain.Asset, error) {
	return r.DB.ListByProject(ctx, id, page)
}
func (r AssetRepository) ByID(ctx context.Context, id string) (*domain.Asset, error) {
	return r.DB.AssetByID(ctx, id)
}
func (r AssetRepository) Update(ctx context.Context, a *domain.Asset) error {
	return r.DB.UpdateAsset(ctx, a)
}
func (r AssetRepository) Delete(ctx context.Context, id string) error {
	return r.DB.DeleteAsset(ctx, id)
}

type UserRepository struct{ DB *Postgres }

func (r UserRepository) Create(ctx context.Context, u *domain.User) error {
	return r.DB.Create(ctx, u)
}
func (r UserRepository) ByEmail(ctx context.Context, email string) (*domain.User, error) {
	return r.DB.ByEmail(ctx, email)
}
func (r UserRepository) ByID(ctx context.Context, id string) (*domain.User, error) {
	return r.DB.ByID(ctx, id)
}
func (r UserRepository) Count(ctx context.Context) (int64, error) {
	return r.DB.Count(ctx)
}
func (r UserRepository) List(ctx context.Context, page domain.Page) ([]domain.User, error) {
	return r.DB.ListUsers(ctx, page)
}
func (r UserRepository) UpdateRole(ctx context.Context, id string, role domain.Role) error {
	return r.DB.UpdateUserRole(ctx, id, role)
}
