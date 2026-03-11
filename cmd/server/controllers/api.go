package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/xid"
	"gorm.io/gorm"

	mkpgo "github.com/elvuel/mkp-go"
	servercfg "github.com/elvuel/mkp-go/cmd/server/config"
	"github.com/elvuel/mkp-go/cmd/server/models"
	mkpcontroller "github.com/elvuel/mkp-go/controller"
)

type API struct {
	mkpCtrl *mkpcontroller.Controller
	sfport  *mkpgo.SFSerialPort
	auth    servercfg.AuthConfig
	db      *gorm.DB

	alogMu        sync.Mutex
	alogRunning   bool
	currentAlogID string
	currentAlog   *alogSession

	replayMu        sync.Mutex
	replayRunning   bool
	currentReplayID string
	replayDoneAt    time.Time
	replayTimer     *time.Timer
}

type alogSession struct {
	Name         string
	UniqueID     string
	MKPPath      string
	StartPointX  int
	StartPointY  int
	ScreenWidth  int
	ScreenHeight int
	OS           string
}

type jwtClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type tokenRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type alogRequest struct {
	LogName string `json:"log_name" binding:"required"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	StPosX  *int   `json:"stposx"`
	StPosY  *int   `json:"stposy"`
}

func (r alogRequest) logOption() *mkpgo.LogOption {
	hasOption := r.Width > 0 || r.Height > 0 || r.StPosX != nil || r.StPosY != nil
	if !hasOption {
		return nil
	}

	opt := &mkpgo.LogOption{}
	opt.StPos.X = -1
	opt.StPos.Y = -1
	if r.Width > 0 {
		opt.Width = r.Width
	}
	if r.Height > 0 {
		opt.Height = r.Height
	}
	if r.StPosX != nil {
		opt.StPos.X = *r.StPosX
	}
	if r.StPosY != nil {
		opt.StPos.Y = *r.StPosY
	}
	return opt
}

func NewAPI(sfportName string, auth servercfg.AuthConfig, db *gorm.DB) (*API, error) {
	sfport := mkpgo.NewSFSerialPort()
	sfport.Name = sfportName
	if err := sfport.Open(); err != nil {
		return nil, err
	}
	go sfport.Read()

	return &API{
		mkpCtrl: mkpcontroller.NewController(sfport),
		sfport:  sfport,
		auth:    auth,
		db:      db,
	}, nil
}

func (a *API) Close() {
	if a.sfport != nil {
		_ = a.sfport.Close()
	}
}

func (a *API) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	api.POST("/auth/token", a.handleToken)

	protected := api.Group("/directives")
	if !(a.auth.MuteInDebug && gin.Mode() == gin.DebugMode) {
		protected.Use(a.jwtAuthMiddleware())
	}
	protected.GET("/list", a.handleList)
	protected.GET("/records/:id", a.handleGetRecord)
	protected.POST("/aplay/:id", a.handleAplay)
	protected.POST("/alog", a.handleAlog)
	protected.POST("/astop", a.handleAstop)
}

func (a *API) unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"ok":    false,
		"error": message,
	})
	c.Abort()
}

func (a *API) jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			a.unauthorized(c, "missing Authorization header")
			return
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			a.unauthorized(c, "invalid Authorization header format")
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			a.unauthorized(c, "missing bearer token")
			return
		}

		claims := &jwtClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(a.auth.JWTSecret), nil
		})
		if err != nil || !token.Valid {
			a.unauthorized(c, "invalid or expired token")
			return
		}

		c.Set("jwt_username", claims.Username)
		c.Next()
	}
}

func (a *API) handleToken(c *gin.Context) {
	var req tokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}
	if req.Username != a.auth.User || req.Password != a.auth.Password {
		a.unauthorized(c, "invalid username or password")
		return
	}

	now := time.Now()
	claims := jwtClaims{
		Username: req.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    a.auth.Issuer,
			Subject:   req.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(a.auth.TokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(a.auth.JWTSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": "failed to sign token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"token":      tokenString,
		"token_type": "Bearer",
		"expires_in": int(a.auth.TokenTTL.Seconds()),
	})
}

func (a *API) handleAlog(c *gin.Context) {
	var req alogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	a.alogMu.Lock()
	defer a.alogMu.Unlock()

	if a.alogRunning {
		c.JSON(http.StatusConflict, gin.H{"ok": false, "error": "alog is already running"})
		return
	}

	alogID := xid.New().String()
	if err := a.mkpCtrl.StartRecord(alogID, req.logOption()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	startPointX := -1
	if req.StPosX != nil {
		startPointX = *req.StPosX
	}
	startPointY := -1
	if req.StPosY != nil {
		startPointY = *req.StPosY
	}

	a.alogRunning = true
	a.currentAlogID = alogID
	a.currentAlog = &alogSession{
		Name:         req.LogName,
		UniqueID:     alogID,
		MKPPath:      a.mkpCtrl.ComposeLogFullpath(req.LogName),
		StartPointX:  startPointX,
		StartPointY:  startPointY,
		ScreenWidth:  req.Width,
		ScreenHeight: req.Height,
		OS:           runtime.GOOS,
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"directive": "alog",
		"id":        alogID,
		"log_name":  req.LogName,
		"status":    "started",
	})
}

func (a *API) handleAstop(c *gin.Context) {
	a.alogMu.Lock()
	defer a.alogMu.Unlock()

	if !a.alogRunning {
		c.JSON(http.StatusConflict, gin.H{"ok": false, "error": "alog is not running"})
		return
	}
	if err := a.mkpCtrl.StopRecord(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	stoppedID := a.currentAlogID
	stoppedSession := a.currentAlog
	a.alogRunning = false
	a.currentAlogID = ""
	a.currentAlog = nil

	resp := gin.H{
		"ok":        true,
		"directive": "astop",
		"id":        stoppedID,
		"status":    "stopped",
	}
	if stoppedSession != nil {
		if record, err := a.persistMacroRecord(stoppedSession); err != nil {
			resp["persisted"] = false
			resp["persist_error"] = err.Error()
		} else {
			resp["persisted"] = true
			resp["macro_record_id"] = record.ID
		}
	} else {
		resp["persisted"] = false
		resp["persist_error"] = "missing alog session state"
	}
	c.JSON(http.StatusOK, resp)
}

func (a *API) handleList(c *gin.Context) {
	limits := 10
	if raw := strings.TrimSpace(c.Query("limits")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			limits = 10
		}
		if parsed > 10 {
			limits = 10
		} else {
			limits = parsed
		}
	}
	nameFilter := strings.TrimSpace(c.Query("name"))

	if a.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "database is not initialized",
		})
		return
	}

	query := a.db
	if nameFilter != "" {
		query = query.Where("name LIKE ?", "%"+nameFilter+"%")
	}

	records := make([]models.MacroRecord, 0, limits)
	if err := query.Order("created_at DESC").Limit(limits).Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	resp := gin.H{
		"ok":        true,
		"directive": "list",
		"limits":    limits,
		"count":     len(records),
		"records":   records,
	}
	if nameFilter != "" {
		resp["name"] = nameFilter
	}
	c.JSON(http.StatusOK, resp)
}

func (a *API) handleGetRecord(c *gin.Context) {
	if a.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "database is not initialized",
		})
		return
	}

	rawUniqueID := strings.TrimSpace(c.Param("id"))
	if rawUniqueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "missing record id",
		})
		return
	}

	var record models.MacroRecord
	if err := a.db.Where("unique_id = ?", rawUniqueID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"ok":    false,
				"error": "record not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"directive": "record",
		"record":    record,
	})
}

func (a *API) handleAplay(c *gin.Context) {
	if a.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "database is not initialized",
		})
		return
	}

	a.replayMu.Lock()
	if a.replayRunning {
		if time.Now().Before(a.replayDoneAt) {
			remaining := time.Until(a.replayDoneAt)
			a.replayMu.Unlock()
			c.JSON(http.StatusConflict, gin.H{
				"ok":                  false,
				"error":               "replay is already running",
				"retry_after_seconds": int(remaining.Seconds()) + 1,
			})
			return
		}

		a.replayRunning = false
		a.currentReplayID = ""
		a.replayDoneAt = time.Time{}
		if a.replayTimer != nil {
			a.replayTimer.Stop()
			a.replayTimer = nil
		}
	}
	a.replayMu.Unlock()

	rawUniqueID := strings.TrimSpace(c.Param("id"))
	if rawUniqueID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"ok":    false,
			"error": "missing record id",
		})
		return
	}

	var record models.MacroRecord
	if err := a.db.Where("unique_id = ?", rawUniqueID).First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"ok":    false,
				"error": "record not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	if err := a.sfport.StartReplaying(record.MKPPath, -1); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	timeoutSeconds := record.Seconds + 1
	if timeoutSeconds < 1 {
		timeoutSeconds = 1
	}
	finishAt := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)

	a.replayMu.Lock()
	a.replayRunning = true
	a.currentReplayID = record.UniqueID
	a.replayDoneAt = finishAt
	if a.replayTimer != nil {
		a.replayTimer.Stop()
	}
	replayID := record.UniqueID
	a.replayTimer = time.AfterFunc(time.Until(finishAt), func() {
		a.replayMu.Lock()
		defer a.replayMu.Unlock()
		if a.currentReplayID == replayID {
			a.replayRunning = false
			a.currentReplayID = ""
			a.replayDoneAt = time.Time{}
			a.replayTimer = nil
		}
	})
	a.replayMu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"directive": "aplay",
		"id":        record.UniqueID,
		"mkp_path":  record.MKPPath,
		"status":    "started",
	})
}

func (a *API) persistMacroRecord(session *alogSession) (*models.MacroRecord, error) {
	if a.db == nil {
		return nil, fmt.Errorf("database is not initialized")
	}

	record := &models.MacroRecord{
		Name:         session.Name,
		UniqueID:     session.UniqueID,
		MKPPath:      session.MKPPath,
		StartPointX:  session.StartPointX,
		StartPointY:  session.StartPointY,
		ScreenWidth:  session.ScreenWidth,
		ScreenHeight: session.ScreenHeight,
		OS:           session.OS,
	}

	logLength, err := a.mkpCtrl.Atime(session.MKPPath)
	if err == nil && logLength != nil {
		record.Seconds = logLength.Seconds
		record.Milliseconds = logLength.Milliseconds
	}
	if err := a.db.Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}
