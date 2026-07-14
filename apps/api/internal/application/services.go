package application

import (
	"context"
	"encoding/json"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/safescope/safescope/apps/api/internal/domain"
)

type AuthService struct {
	users  domain.UserRepository
	hasher domain.PasswordHasher
	tokens domain.TokenIssuer
}

type UserService struct {
	users domain.UserRepository
}

func NewUserService(users domain.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) List(ctx context.Context, actor *domain.User, page domain.Page) ([]domain.User, error) {
	if actor.Role != domain.RoleAdmin {
		return nil, domain.ErrForbidden
	}
	return s.users.List(ctx, normalizePage(page))
}

func (s *UserService) UpdateRole(ctx context.Context, actor *domain.User, userID string, role domain.Role) error {
	if actor.Role != domain.RoleAdmin {
		return domain.ErrForbidden
	}
	if actor.ID == userID || !role.Valid() {
		return domain.ErrInvalidInput
	}
	return s.users.UpdateRole(ctx, userID, role)
}

func NewAuthService(users domain.UserRepository, hasher domain.PasswordHasher, tokens domain.TokenIssuer) *AuthService {
	return &AuthService{users: users, hasher: hasher, tokens: tokens}
}

type AuthResult struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      *domain.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, name, email, password string) (*AuthResult, error) {
	name, email = strings.TrimSpace(name), strings.ToLower(strings.TrimSpace(email))
	if name == "" || len(password) < 8 {
		return nil, domain.ErrInvalidInput
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, domain.ErrInvalidInput
	}
	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, err
	}
	count, err := s.users.Count(ctx)
	if err != nil {
		return nil, err
	}
	role := domain.RoleAnalyst
	if count == 0 {
		role = domain.RoleAdmin
	}
	now := time.Now().UTC()
	user := &domain.User{
		ID: uuid.NewString(), Email: email, Name: name, PasswordHash: hash,
		Role: role, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return s.result(user)
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.users.ByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil || s.hasher.Compare(user.PasswordHash, password) != nil {
		return nil, domain.ErrUnauthorized
	}
	return s.result(user)
}

func (s *AuthService) result(user *domain.User) (*AuthResult, error) {
	token, expires, err := s.tokens.Issue(user)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: token, ExpiresAt: expires, User: user}, nil
}

type ProjectService struct {
	projects domain.ProjectRepository
	assets   domain.AssetRepository
}

func NewProjectService(projects domain.ProjectRepository, assets domain.AssetRepository) *ProjectService {
	return &ProjectService{projects: projects, assets: assets}
}

func (s *ProjectService) Create(ctx context.Context, actor *domain.User, name, description string) (*domain.Project, error) {
	if actor.Role == domain.RoleViewer || strings.TrimSpace(name) == "" {
		return nil, domain.ErrForbidden
	}
	now := time.Now().UTC()
	project := &domain.Project{
		ID: uuid.NewString(), Name: strings.TrimSpace(name), Description: strings.TrimSpace(description),
		Status: domain.ProjectActive, OwnerID: actor.ID, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.projects.Create(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) List(ctx context.Context, actor *domain.User, page domain.Page) ([]domain.Project, error) {
	return s.projects.List(ctx, actor, normalizePage(page))
}

func (s *ProjectService) Get(ctx context.Context, actor *domain.User, id string) (*domain.Project, error) {
	if err := s.authorize(ctx, actor, id, false); err != nil {
		return nil, err
	}
	return s.projects.ByID(ctx, id)
}

func (s *ProjectService) Update(ctx context.Context, actor *domain.User, id, name, description string, status domain.ProjectStatus) (*domain.Project, error) {
	if err := s.authorize(ctx, actor, id, true); err != nil {
		return nil, err
	}
	project, err := s.projects.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(name) != "" {
		project.Name = strings.TrimSpace(name)
	}
	project.Description = strings.TrimSpace(description)
	if status == domain.ProjectActive || status == domain.ProjectArchived {
		project.Status = status
	}
	project.UpdatedAt = time.Now().UTC()
	if err := s.projects.Update(ctx, project); err != nil {
		return nil, err
	}
	return project, nil
}

func (s *ProjectService) Delete(ctx context.Context, actor *domain.User, id string) error {
	if err := s.authorize(ctx, actor, id, true); err != nil {
		return err
	}
	return s.projects.Delete(ctx, id)
}

func (s *ProjectService) CreateAsset(ctx context.Context, actor *domain.User, projectID string, assetType domain.AssetType, value string, status domain.AssetStatus, tags []string, metadata json.RawMessage) (*domain.Asset, error) {
	if err := s.authorize(ctx, actor, projectID, true); err != nil {
		return nil, err
	}
	if !assetType.Valid() || strings.TrimSpace(value) == "" {
		return nil, domain.ErrInvalidInput
	}
	if status == "" {
		status = domain.AssetUnknown
	}
	if !status.Valid() {
		return nil, domain.ErrInvalidInput
	}
	if len(metadata) == 0 {
		metadata = json.RawMessage(`{}`)
	}
	if tags == nil {
		tags = []string{}
	}
	if !json.Valid(metadata) {
		return nil, domain.ErrInvalidInput
	}
	now := time.Now().UTC()
	asset := &domain.Asset{
		ID: uuid.NewString(), ProjectID: projectID, Type: assetType, Value: strings.TrimSpace(value),
		Status: status, Tags: tags, Metadata: metadata, CreatedAt: now, UpdatedAt: now,
	}
	if err := s.assets.Create(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *ProjectService) ListAssets(ctx context.Context, actor *domain.User, projectID string, page domain.Page) ([]domain.Asset, error) {
	if err := s.authorize(ctx, actor, projectID, false); err != nil {
		return nil, err
	}
	return s.assets.ListByProject(ctx, projectID, normalizePage(page))
}

func (s *ProjectService) UpdateAsset(ctx context.Context, actor *domain.User, id string, status domain.AssetStatus, tags []string, metadata json.RawMessage) (*domain.Asset, error) {
	asset, err := s.assets.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := s.authorize(ctx, actor, asset.ProjectID, true); err != nil {
		return nil, err
	}
	if status.Valid() {
		asset.Status = status
	}
	if tags != nil {
		asset.Tags = tags
	}
	if len(metadata) > 0 {
		if !json.Valid(metadata) {
			return nil, domain.ErrInvalidInput
		}
		asset.Metadata = metadata
	}
	asset.UpdatedAt = time.Now().UTC()
	if err := s.assets.Update(ctx, asset); err != nil {
		return nil, err
	}
	return asset, nil
}

func (s *ProjectService) DeleteAsset(ctx context.Context, actor *domain.User, id string) error {
	asset, err := s.assets.ByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.authorize(ctx, actor, asset.ProjectID, true); err != nil {
		return err
	}
	return s.assets.Delete(ctx, id)
}

func (s *ProjectService) authorize(ctx context.Context, actor *domain.User, projectID string, write bool) error {
	if write && actor.Role == domain.RoleViewer {
		return domain.ErrForbidden
	}
	ok, err := s.projects.CanAccess(ctx, actor, projectID)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrForbidden
	}
	return nil
}

func normalizePage(page domain.Page) domain.Page {
	if page.Limit <= 0 || page.Limit > 100 {
		page.Limit = 50
	}
	if page.Offset < 0 {
		page.Offset = 0
	}
	return page
}

type DashboardService struct {
	repo domain.DashboardRepository
}

func NewDashboardService(repo domain.DashboardRepository) *DashboardService {
	return &DashboardService{repo: repo}
}

func (s *DashboardService) Summary(ctx context.Context, actor *domain.User) (*domain.Dashboard, error) {
	return s.repo.Summary(ctx, actor)
}
