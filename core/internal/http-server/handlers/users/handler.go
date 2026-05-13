package users

import (
	"net/http"

	"base/internal/http-server/handlers/common"
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
	r.POST("/users", h.create)
	r.GET("/users", h.list)
	r.GET("/users/me", h.me)
	r.GET("/users/:user_id", h.get)

	r.POST("/auth/login", h.login)
}

func (h *Handler) create(c *gin.Context) {
	var req createRequest
	if !common.BindJSON(c, &req) {
		return
	}

	registered, err := h.svc.CreateUser(c.Request.Context(), service.CreateUserRequest{
		FullName: req.FullName,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	})
	if common.HandleError(c, err) {
		return
	}
	c.JSON(http.StatusCreated, registeredUserResponseFromModel(registered))
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
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Phone    string `json:"phone"`
	Password string `json:"password" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
