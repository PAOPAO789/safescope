package domain

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("resource not found")
	ErrConflict     = errors.New("resource already exists")
	ErrForbidden    = errors.New("access denied")
	ErrUnauthorized = errors.New("authentication required")
	ErrInvalidInput = errors.New("invalid input")
)

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleAnalyst Role = "analyst"
	RoleViewer  Role = "viewer"
)

func (r Role) Valid() bool {
	return r == RoleAdmin || r == RoleAnalyst || r == RoleViewer
}

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProjectStatus string

const (
	ProjectActive   ProjectStatus = "active"
	ProjectArchived ProjectStatus = "archived"
)

type Project struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Status      ProjectStatus `json:"status"`
	OwnerID     string        `json:"owner_id"`
	AssetCount  int64         `json:"asset_count,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type AssetType string

const (
	AssetDomain  AssetType = "domain"
	AssetIP      AssetType = "ip"
	AssetURL     AssetType = "url"
	AssetService AssetType = "service"
)

func (t AssetType) Valid() bool {
	return t == AssetDomain || t == AssetIP || t == AssetURL || t == AssetService
}

type AssetStatus string

const (
	AssetUnknown AssetStatus = "unknown"
	AssetAlive   AssetStatus = "alive"
	AssetDown    AssetStatus = "down"
)

func (s AssetStatus) Valid() bool {
	return s == AssetUnknown || s == AssetAlive || s == AssetDown
}

type Asset struct {
	ID        string          `json:"id"`
	ProjectID string          `json:"project_id"`
	Type      AssetType       `json:"type"`
	Value     string          `json:"value"`
	Status    AssetStatus     `json:"status"`
	Tags      []string        `json:"tags"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type Dashboard struct {
	Projects     int64    `json:"projects"`
	Assets       int64    `json:"assets"`
	AliveAssets  int64    `json:"alive_assets"`
	Archived     int64    `json:"archived_projects"`
	RecentAssets []Asset  `json:"recent_assets"`
	AssetsByType []Metric `json:"assets_by_type"`
}

type Metric struct {
	Label string `json:"label"`
	Value int64  `json:"value"`
}

type Page struct {
	Limit  int
	Offset int
}

type UserRepository interface {
	Create(context.Context, *User) error
	ByEmail(context.Context, string) (*User, error)
	ByID(context.Context, string) (*User, error)
	Count(context.Context) (int64, error)
	List(context.Context, Page) ([]User, error)
	UpdateRole(context.Context, string, Role) error
}

type ProjectRepository interface {
	Create(context.Context, *Project) error
	List(context.Context, *User, Page) ([]Project, error)
	ByID(context.Context, string) (*Project, error)
	Update(context.Context, *Project) error
	Delete(context.Context, string) error
	CanAccess(context.Context, *User, string) (bool, error)
}

type AssetRepository interface {
	Create(context.Context, *Asset) error
	ListByProject(context.Context, string, Page) ([]Asset, error)
	ByID(context.Context, string) (*Asset, error)
	Update(context.Context, *Asset) error
	Delete(context.Context, string) error
}

type DashboardRepository interface {
	Summary(context.Context, *User) (*Dashboard, error)
}

type PasswordHasher interface {
	Hash(string) (string, error)
	Compare(string, string) error
}

type TokenIssuer interface {
	Issue(*User) (string, time.Time, error)
}

type Cache interface {
	Ping(context.Context) error
}

// Scanner and AIProvider are extension ports for future httpx/naabu/nuclei and LLM adapters.
type Scanner interface {
	Name() string
	Run(context.Context, Project, []Asset) error
}

type AIProvider interface {
	SummarizeRisk(context.Context, Project, []Asset) (string, error)
}
