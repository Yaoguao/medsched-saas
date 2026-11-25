package users

import (
	"net/http"

	"base/internal/domain/models"
	"base/internal/http-server/handlers/common"
	authmiddleware "base/internal/http-server/middleware/auth"
	"base/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.Service
}

func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(r gin.IRouter) {
	h.RegisterPublic(r)
	h.RegisterProtected(r)
}

func (h *Handler) RegisterPublic(r gin.IRouter) {
	r.POST("/users", h.create)
	r.POST("/auth/login", h.login)
}

func (h *Handler) RegisterProtected(r gin.IRouter) {
	r.GET("/users", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleRegistrator, models.UserRoleITSupport), h.list)
	r.POST("/users/staff", authmiddleware.RequireRoles(models.UserRoleAdmin), h.createStaff)
	r.GET("/users/me", h.me)
	r.GET("/users/:user_id", authmiddleware.RequireRoles(models.UserRoleAdmin, models.UserRoleITSupport), h.get)
	r.PATCH("/users/:user_id", authmiddleware.RequireRoles(models.UserRoleAdmin), h.update)
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if !common.BindJSON(c, &req) {
		return
	}

	user, err := h.svc.CreateUser(c.Request.Context(), service.CreateUserRequest{
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, userResponseFromModel(user))
}

func (h *Handler) createStaff(c *gin.Context) {
	var req createStaffRequest
	if !common.BindJSON(c, &req) {
		return
	}

	user, err := h.svc.CreateStaffUser(c.Request.Context(), service.CreateStaffUserRequest{
		Role:     req.Role,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, userResponseFromModel(user))
}

func (h *Handler) list(c *gin.Context) {
	page := common.Pagination(c)
	users, err := h.svc.ListUsers(c.Request.Context(), page)
	if common.HandleError(c, err) {
		return
	}
	common.RespondList(c, userResponsesFromModels(users), page)
}

func (h *Handler) get(c *gin.Context) {
	id, ok := common.PathUUID(c, "user_id")
	if !ok {
		return
	}
	user, err := h.svc.GetUser(c.Request.Context(), id)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, userResponseFromModel(user))
}

func (h *Handler) update(c *gin.Context) {
	id, ok := common.PathUUID(c, "user_id")
	if !ok {
		return
	}

	var req updateRequest
	if !common.BindJSON(c, &req) {
		return
	}

	user, err := h.svc.UpdateUser(c.Request.Context(), id, service.UserPatch{
		Role:   req.Role,
		Status: req.Status,
		Email:  req.Email,
		Phone:  req.Phone,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, userResponseFromModel(user))
}

func (h *Handler) me(c *gin.Context) {
	user, err := h.svc.CurrentUserFromAuthorization(c.Request.Context(), c.GetHeader("Authorization"))
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, userResponseFromModel(user))
}

func (h *Handler) login(c *gin.Context) {
	var req loginRequest
	if !common.BindJSON(c, &req) {
		return
	}
	token, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusOK, authResponseFromToken(token))
}

type createRequest struct {
	Email    string `json:"email" binding:"required"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required"`
}

type createStaffRequest struct {
	Role     models.UserRole `json:"role" binding:"required"`
	Email    string          `json:"email" binding:"required"`
	Phone    string          `json:"phone" binding:"required"`
	Password string          `json:"password" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type updateRequest struct {
	Role   *models.UserRole `json:"role"`
	Status *string          `json:"status"`
	Email  *string          `json:"email"`
	Phone  *string          `json:"phone"`
}
