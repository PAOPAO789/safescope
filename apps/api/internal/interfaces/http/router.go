package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/safescope/safescope/apps/api/internal/application"
	"github.com/safescope/safescope/apps/api/internal/domain"
	"github.com/safescope/safescope/apps/api/internal/platform/security"
	"go.uber.org/zap"
)

const actorKey = "actor"

type Dependencies struct {
	Auth        *application.AuthService
	Users       *application.UserService
	Projects    *application.ProjectService
	Dashboard   *application.DashboardService
	Tokens      security.JWTManager
	Logger      *zap.Logger
	CORSOrigins []string
	Health      func() error
}

func NewRouter(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(requestID(), accessLog(deps.Logger), gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins: deps.CORSOrigins,
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{"Authorization", "Content-Type", "X-Request-ID"},
		MaxAge:       12 * time.Hour,
	}))

	router.GET("/healthz", func(c *gin.Context) {
		if err := deps.Health(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/openapi.yaml", func(c *gin.Context) {
		c.File("openapi/openapi.yaml")
	})

	api := router.Group("/api/v1")
	api.POST("/auth/register", register(deps.Auth))
	api.POST("/auth/login", login(deps.Auth))

	protected := api.Group("")
	protected.Use(authenticate(deps.Tokens))
	protected.GET("/me", func(c *gin.Context) { c.JSON(http.StatusOK, actor(c)) })
	protected.GET("/dashboard", dashboard(deps.Dashboard))
	protected.GET("/users", listUsers(deps.Users))
	protected.PUT("/users/:userID/role", updateUserRole(deps.Users))
	protected.GET("/projects", listProjects(deps.Projects))
	protected.POST("/projects", createProject(deps.Projects))
	protected.GET("/projects/:projectID", getProject(deps.Projects))
	protected.PUT("/projects/:projectID", updateProject(deps.Projects))
	protected.DELETE("/projects/:projectID", deleteProject(deps.Projects))
	protected.GET("/projects/:projectID/assets", listAssets(deps.Projects))
	protected.POST("/projects/:projectID/assets", createAsset(deps.Projects))
	protected.PUT("/assets/:assetID", updateAsset(deps.Projects))
	protected.DELETE("/assets/:assetID", deleteAsset(deps.Projects))
	return router
}

func listUsers(service *application.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		users, err := service.List(c, actor(c), page(c))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": users})
	}
}

func updateUserRole(service *application.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request struct {
			Role domain.Role `json:"role"`
		}
		if c.ShouldBindJSON(&request) != nil {
			writeError(c, domain.ErrInvalidInput)
			return
		}
		if err := service.UpdateRole(c, actor(c), c.Param("userID"), request.Role); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type authRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func register(service *application.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request authRequest
		if c.ShouldBindJSON(&request) != nil {
			writeError(c, domain.ErrInvalidInput)
			return
		}
		result, err := service.Register(c, request.Name, request.Email, request.Password)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusCreated, result)
	}
}

func login(service *application.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request authRequest
		if c.ShouldBindJSON(&request) != nil {
			writeError(c, domain.ErrInvalidInput)
			return
		}
		result, err := service.Login(c, request.Email, request.Password)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

func dashboard(service *application.DashboardService) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := service.Summary(c, actor(c))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, result)
	}
}

type projectRequest struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Status      domain.ProjectStatus `json:"status"`
}

func createProject(service *application.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request projectRequest
		if c.ShouldBindJSON(&request) != nil {
			writeError(c, domain.ErrInvalidInput)
			return
		}
		project, err := service.Create(c, actor(c), request.Name, request.Description)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusCreated, project)
	}
}

func listProjects(service *application.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		projects, err := service.List(c, actor(c), page(c))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": projects})
	}
}

func getProject(service *application.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		project, err := service.Get(c, actor(c), c.Param("projectID"))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, project)
	}
}

func updateProject(service *application.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request projectRequest
		if c.ShouldBindJSON(&request) != nil {
			writeError(c, domain.ErrInvalidInput)
			return
		}
		project, err := service.Update(c, actor(c), c.Param("projectID"), request.Name, request.Description, request.Status)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, project)
	}
}

func deleteProject(service *application.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := service.Delete(c, actor(c), c.Param("projectID")); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

type assetRequest struct {
	Type     domain.AssetType   `json:"type"`
	Value    string             `json:"value"`
	Status   domain.AssetStatus `json:"status"`
	Tags     []string           `json:"tags"`
	Metadata map[string]any     `json:"metadata"`
}

func createAsset(service *application.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request assetRequest
		if c.ShouldBindJSON(&request) != nil {
			writeError(c, domain.ErrInvalidInput)
			return
		}
		metadata, err := marshalMetadata(request.Metadata)
		if err != nil {
			writeError(c, domain.ErrInvalidInput)
			return
		}
		asset, err := service.CreateAsset(c, actor(c), c.Param("projectID"), request.Type, request.Value, request.Tags, metadata)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusCreated, asset)
	}
}

func listAssets(service *application.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		assets, err := service.ListAssets(c, actor(c), c.Param("projectID"), page(c))
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": assets})
	}
}

func updateAsset(service *application.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request assetRequest
		if c.ShouldBindJSON(&request) != nil {
			writeError(c, domain.ErrInvalidInput)
			return
		}
		metadata, err := marshalMetadata(request.Metadata)
		if err != nil {
			writeError(c, domain.ErrInvalidInput)
			return
		}
		asset, err := service.UpdateAsset(c, actor(c), c.Param("assetID"), request.Status, request.Tags, metadata)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, asset)
	}
}

func deleteAsset(service *application.ProjectService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := service.DeleteAsset(c, actor(c), c.Param("assetID")); err != nil {
			writeError(c, err)
			return
		}
		c.Status(http.StatusNoContent)
	}
}

func authenticate(tokens security.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}
		user, err := tokens.Parse(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			writeError(c, domain.ErrUnauthorized)
			c.Abort()
			return
		}
		c.Set(actorKey, user)
		c.Next()
	}
}

func actor(c *gin.Context) *domain.User {
	value, _ := c.Get(actorKey)
	user, _ := value.(*domain.User)
	return user
}

func page(c *gin.Context) domain.Page {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	return domain.Page{Limit: limit, Offset: offset}
}

func writeError(c *gin.Context, err error) {
	status, code := http.StatusInternalServerError, "internal_error"
	message := "an internal error occurred"
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		status, code = http.StatusBadRequest, "invalid_input"
		message = err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		status, code = http.StatusUnauthorized, "unauthorized"
		message = err.Error()
	case errors.Is(err, domain.ErrForbidden):
		status, code = http.StatusForbidden, "forbidden"
		message = err.Error()
	case errors.Is(err, domain.ErrNotFound):
		status, code = http.StatusNotFound, "not_found"
		message = err.Error()
	case errors.Is(err, domain.ErrConflict):
		status, code = http.StatusConflict, "conflict"
		message = err.Error()
	}
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": message}})
}

func requestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}
		c.Header("X-Request-ID", id)
		c.Set("request_id", id)
		c.Next()
	}
}

func accessLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http_request",
			zap.String("request_id", c.GetString("request_id")),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
		)
	}
}

func marshalMetadata(value map[string]any) ([]byte, error) {
	if value == nil {
		return nil, nil
	}
	return json.Marshal(value)
}
