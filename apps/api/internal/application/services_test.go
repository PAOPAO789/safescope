package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/safescope/safescope/apps/api/internal/domain"
)

type fakeUsers struct{ users map[string]*domain.User }

func (f *fakeUsers) Create(_ context.Context, u *domain.User) error {
	if _, ok := f.users[u.Email]; ok {
		return domain.ErrConflict
	}
	f.users[u.Email] = u
	return nil
}
func (f *fakeUsers) ByEmail(_ context.Context, email string) (*domain.User, error) {
	u, ok := f.users[email]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return u, nil
}
func (f *fakeUsers) ByID(_ context.Context, id string) (*domain.User, error) {
	for _, u := range f.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, domain.ErrNotFound
}
func (f *fakeUsers) Count(context.Context) (int64, error) { return int64(len(f.users)), nil }
func (f *fakeUsers) List(context.Context, domain.Page) ([]domain.User, error) {
	out := make([]domain.User, 0, len(f.users))
	for _, user := range f.users {
		out = append(out, *user)
	}
	return out, nil
}
func (f *fakeUsers) UpdateRole(_ context.Context, id string, role domain.Role) error {
	for _, user := range f.users {
		if user.ID == id {
			user.Role = role
			return nil
		}
	}
	return domain.ErrNotFound
}

type fakeHasher struct{}

func (fakeHasher) Hash(v string) (string, error) { return "hash:" + v, nil }
func (fakeHasher) Compare(hash, v string) error {
	if hash != "hash:"+v {
		return errors.New("mismatch")
	}
	return nil
}

type fakeTokens struct{}

func (fakeTokens) Issue(*domain.User) (string, time.Time, error) {
	return "token", time.Now().Add(time.Hour), nil
}

func TestRegisterAndLogin(t *testing.T) {
	users := &fakeUsers{users: map[string]*domain.User{}}
	service := NewAuthService(users, fakeHasher{}, fakeTokens{})
	result, err := service.Register(context.Background(), "Alice", "ALICE@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if result.User.Email != "alice@example.com" || result.User.Role != domain.RoleAdmin {
		t.Fatalf("unexpected user: %#v", result.User)
	}
	second, err := service.Register(context.Background(), "Bob", "bob@example.com", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if second.User.Role != domain.RoleAnalyst {
		t.Fatalf("expected analyst, got %s", second.User.Role)
	}
	if _, err := service.Login(context.Background(), "alice@example.com", "wrong"); !errors.Is(err, domain.ErrUnauthorized) {
		t.Fatalf("expected unauthorized, got %v", err)
	}
}

type fakeProjects struct{ project *domain.Project }

func (f *fakeProjects) Create(_ context.Context, p *domain.Project) error { f.project = p; return nil }
func (f *fakeProjects) List(context.Context, *domain.User, domain.Page) ([]domain.Project, error) {
	return []domain.Project{*f.project}, nil
}
func (f *fakeProjects) ByID(_ context.Context, _ string) (*domain.Project, error) {
	return f.project, nil
}
func (f *fakeProjects) Update(_ context.Context, p *domain.Project) error { f.project = p; return nil }
func (f *fakeProjects) Delete(context.Context, string) error              { return nil }
func (f *fakeProjects) CanAccess(_ context.Context, actor *domain.User, _ string) (bool, error) {
	return actor.Role == domain.RoleAdmin || actor.ID == f.project.OwnerID, nil
}

type fakeAssets struct{ asset *domain.Asset }

func (f *fakeAssets) Create(_ context.Context, a *domain.Asset) error { f.asset = a; return nil }
func (f *fakeAssets) ListByProject(context.Context, string, domain.Page) ([]domain.Asset, error) {
	return []domain.Asset{*f.asset}, nil
}
func (f *fakeAssets) ByID(context.Context, string) (*domain.Asset, error) { return f.asset, nil }
func (f *fakeAssets) Update(_ context.Context, a *domain.Asset) error     { f.asset = a; return nil }
func (f *fakeAssets) Delete(context.Context, string) error                { return nil }

func TestViewerCannotCreateProjectAndOwnerCanCreateAsset(t *testing.T) {
	projects, assets := &fakeProjects{}, &fakeAssets{}
	service := NewProjectService(projects, assets)
	viewer := &domain.User{ID: "viewer", Role: domain.RoleViewer}
	if _, err := service.Create(context.Background(), viewer, "Nope", ""); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}

	owner := &domain.User{ID: "owner", Role: domain.RoleAnalyst}
	project, err := service.Create(context.Background(), owner, "External perimeter", "")
	if err != nil {
		t.Fatal(err)
	}
	asset, err := service.CreateAsset(context.Background(), owner, project.ID, domain.AssetDomain, "example.com", domain.AssetAlive, []string{"prod"}, json.RawMessage(`{"source":"manual"}`))
	if err != nil {
		t.Fatal(err)
	}
	if asset.Value != "example.com" || asset.Status != domain.AssetAlive {
		t.Fatalf("unexpected asset: %#v", asset)
	}
	if _, err := service.CreateAsset(context.Background(), owner, project.ID, domain.AssetURL, "https://example.com", "bad", nil, nil); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("expected invalid status, got %v", err)
	}
	untagged, err := service.CreateAsset(context.Background(), owner, project.ID, domain.AssetIP, "192.0.2.1", "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if untagged.Tags == nil || string(untagged.Metadata) != "{}" || untagged.Status != domain.AssetUnknown {
		t.Fatalf("expected normalized empty values: %#v", untagged)
	}
}

func TestOnlyAdminCanUpdateAnotherUsersRole(t *testing.T) {
	users := &fakeUsers{users: map[string]*domain.User{
		"admin@example.com": {ID: "admin", Role: domain.RoleAdmin},
		"user@example.com":  {ID: "user", Role: domain.RoleViewer},
	}}
	service := NewUserService(users)
	if err := service.UpdateRole(context.Background(), users.users["user@example.com"], "admin", domain.RoleViewer); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	if err := service.UpdateRole(context.Background(), users.users["admin@example.com"], "user", domain.RoleAnalyst); err != nil {
		t.Fatal(err)
	}
	if users.users["user@example.com"].Role != domain.RoleAnalyst {
		t.Fatal("role was not updated")
	}
}
