package handler

import (
	"net/http"

	"game-link/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func (h *Handler) Register(c *gin.Context) {
	var body struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Fullname string `json:"fullname" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.store.NewUser(c.Request.Context(), body.Username, body.Email, body.Password, body.Fullname)

	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id":       u.ID,
		"username": u.Username,
		"email":    u.Email,
		"fullname": u.Fullname,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	u, err := h.store.Authenticate(c.Request.Context(), body.Email, body.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	token, err := auth.GenerateToken(u)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) SearchUsers(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name not set"})
		return
	}

	claims := c.MustGet("claims").(jwt.MapClaims)
	currentUserID := claims["sub"].(string)

	users, err := h.store.SearchUsersByFullname(c.Request.Context(), name, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handler) GetUserByUsername(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	currentUserID := claims["sub"].(string)

	user, err := h.store.GetUserByUsername(c.Request.Context(), c.Param("username"), currentUserID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	availability, err := h.store.GetUserAvailability(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"fullname":     user.Fullname,
		"biography":    user.Biography,
		"verified":     user.Verified,
		"avatarUrl":    user.AvatarURL,
		"linkStatus":   user.LinkStatus,
		"availability": availability,
	})
}

func (h *Handler) CheckEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email not set"})
		return
	}

	exists, err := h.store.EmailExists(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

func (h *Handler) CheckUsername(c *gin.Context) {
	username := c.Query("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username not set"})
		return
	}

	exists, err := h.store.UsernameExists(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"exists": exists})
}

func (h *Handler) Me(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)

	user, err := h.store.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	games, err := h.store.GetUserGames(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	availability, err := h.store.GetUserAvailability(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":           user.ID,
		"username":     user.Username,
		"email":        user.Email,
		"fullname":     user.Fullname,
		"biography":    user.Biography,
		"verified":     user.Verified,
		"avatarUrl":    user.AvatarURL,
		"games":        games,
		"availability": availability,
	})
}

func (h *Handler) UpdateProfile(c *gin.Context) {
	claims := c.MustGet("claims").(jwt.MapClaims)
	userID := claims["sub"].(string)

	var body struct {
		Fullname  string `json:"fullname" binding:"required"`
		Biography string `json:"biography"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.store.UpdateUserProfile(c.Request.Context(), userID, body.Fullname, body.Biography); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated"})
}
