package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-vgo/robotgo"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/xid"
	"gorm.io/gorm"

	mkpgo "github.com/elvuel/mkp-go"
	servercfg "github.com/elvuel/mkp-go/cmd/server/config"
	"github.com/elvuel/mkp-go/cmd/server/models"
	mkpcontroller "github.com/elvuel/mkp-go/controller"
)

const (
	serialSyncTimeoutGrace = 2 * time.Second
)

type API struct {
	mkpCtrl *mkpcontroller.Controller
	sfport  *mkpgo.SFSerialPort
	auth    servercfg.AuthConfig
	db      *gorm.DB
	runtime servercfg.RuntimeConfig
	version string

	alogMu        sync.Mutex
	alogRunning   bool
	alogStarted   bool
	currentAlogID string
	currentAlog   *alogSession

	replayMu        sync.Mutex
	replayRunning   bool
	currentReplayID string
	replayStartedAt time.Time
	replayDoneAt    time.Time
	replayReleaseAt time.Time
	replayTimer     *time.Timer

	readLoopCancel context.CancelFunc
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

func NewAPI(sfportName string, auth servercfg.AuthConfig, db *gorm.DB, runtimeCfg servercfg.RuntimeConfig, version string) (*API, error) {
	runtimeCfg = servercfg.NormalizeRuntimeConfig(runtimeCfg)

	sfport := mkpgo.NewSFSerialPort()
	sfport.SyncOuputEnabled = true
	sfport.Name = sfportName
	sfport.SyncOutputTimeout = runtimeCfg.SyncDirectiveTimeout
	if err := sfport.Open(); err != nil {
		return nil, err
	}
	readLoopCtx, readLoopCancel := context.WithCancel(context.Background())
	go func() {
		_ = sfport.ReadLoopContext(readLoopCtx)
	}()

	sfport.StopRecording()

	log.Printf(
		"api initialized: sf_port=%s sync_timeout=%s async_timeout=%s db_timeout=%s replay_grace=%s replay_min_duration=%s",
		sfportName,
		runtimeCfg.SyncDirectiveTimeout,
		runtimeCfg.AsyncDirectiveTimeout,
		runtimeCfg.DBTimeout,
		runtimeCfg.ReplayCompletionGrace,
		runtimeCfg.ReplayMinimumDuration,
	)

	return &API{
		mkpCtrl:        mkpcontroller.NewController(sfport),
		sfport:         sfport,
		auth:           auth,
		db:             db,
		runtime:        runtimeCfg,
		version:        strings.TrimSpace(version),
		readLoopCancel: readLoopCancel,
	}, nil
}

func (a *API) Close() {
	if a.readLoopCancel != nil {
		a.readLoopCancel()
	}
	if a.sfport != nil {
		_ = a.sfport.Close()
	}
}

func (a *API) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	api.POST("/auth/token", a.handleToken)
	api.GET("/version", a.handleVersion)

	protected := api.Group("/directives")
	if !(a.auth.MuteInDebug && gin.Mode() == gin.DebugMode) {
		protected.Use(a.jwtAuthMiddleware())
	}
	protected.GET("/list", a.handleList)
	protected.GET("/records/:id", a.handleGetRecord)
	protected.DELETE("/records/:id", a.handleDeleteRecord)
	protected.GET("/replay/status", a.handleReplayStatus)
	protected.POST("/aplay/:id", a.handleAplay)
	protected.POST("/acancel", a.handleAcancel)
	protected.POST("/alog", a.handleAlog)
	protected.POST("/astop", a.handleAstop)
}

func (a *API) HandleHealth(c *gin.Context) {
	serialReady := false
	var serialStatus *mkpgo.SerialPortStatus
	if a.sfport != nil {
		snapshot := a.sfport.SnapshotStatus()
		serialStatus = &snapshot
		serialReady = snapshot.PortOpen && snapshot.ReadLoopRunning && !snapshot.ReadClosed
	}

	alog := a.snapshotAlogState()
	replay := a.snapshotReplayState()
	dbReady := a.db != nil
	ready := serialReady && dbReady
	status := "up"
	if !ready {
		status = "degraded"
		log.Printf("health degraded: ready=%t serial_ready=%t db_ready=%t serial=%+v alog=%+v replay=%+v", ready, serialReady, dbReady, serialStatus, alog, replay)
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"status":   status,
		"ready":    ready,
		"db_ready": dbReady,
		"serial":   serialStatus,
		"alog":     alog,
		"replay":   replay,
	})
}

func (a *API) handleReplayStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"directive": "replay_status",
		"replay":    a.snapshotReplayState(),
	})
}

func (a *API) handleVersion(c *gin.Context) {
	if a.sfport == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "serial port is not initialized",
		})
		return
	}

	ctx, cancel := a.newSyncDirectiveContext(c.Request.Context())
	defer cancel()

	mkpVersion, err := a.mkpCtrl.AversionContext(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":         true,
		"directive":  "version",
		"mkp_server": a.version,
		"mkp_device": mkpVersion,
	})
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
	ctx, cancel := a.newSyncDirectiveContext(c.Request.Context())
	defer cancel()
	var req alogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "error": err.Error()})
		return
	}

	alogID := xid.New().String()
	startPointX := -1
	if req.StPosX != nil {
		startPointX = *req.StPosX
	}
	startPointY := -1
	if req.StPosY != nil {
		startPointY = *req.StPosY
	}

	if startPointX < 0 {
		startPointX = 0
	}

	if startPointY < 0 {
		startPointY = 0
	}

	req.StPosX = &startPointX
	req.StPosY = &startPointY

	if req.Width <= 0 || req.Height <= 0 {
		req.Width, req.Height = robotgo.GetScreenSize()
	}

	session := &alogSession{
		Name:         req.LogName,
		UniqueID:     alogID,
		MKPPath:      a.mkpCtrl.ComposeLogFullpath(alogID),
		StartPointX:  startPointX,
		StartPointY:  startPointY,
		ScreenWidth:  req.Width,
		ScreenHeight: req.Height,
		OS:           runtime.GOOS,
	}

	log.Printf(
		"alog requested: id=%s name=%q start=(%d,%d) size=%dx%d",
		alogID,
		req.LogName,
		startPointX,
		startPointY,
		req.Width,
		req.Height,
	)

	a.alogMu.Lock()
	if a.alogRunning {
		runningID := a.currentAlogID
		a.alogMu.Unlock()
		log.Printf("alog rejected: existing session still running current_id=%s requested_id=%s", runningID, alogID)
		c.JSON(http.StatusConflict, gin.H{"ok": false, "error": "alog is already running"})
		return
	}
	a.alogRunning = true
	a.alogStarted = false
	a.currentAlogID = alogID
	a.currentAlog = session
	a.alogMu.Unlock()

	robotgo.Move(startPointX, startPointY)

	if _, err := a.mkpCtrl.AlogContext(ctx, alogID, req.logOption()); err != nil {
		a.alogMu.Lock()
		if a.currentAlogID == alogID {
			a.alogRunning = false
			a.alogStarted = false
			a.currentAlogID = ""
			a.currentAlog = nil
		}
		a.alogMu.Unlock()
		log.Printf("alog failed: id=%s name=%q err=%v", alogID, req.LogName, err)
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	a.alogMu.Lock()
	if a.currentAlogID == alogID {
		a.alogStarted = true
	}
	a.alogMu.Unlock()

	log.Printf("alog started: id=%s name=%q mkp_path=%s", alogID, req.LogName, session.MKPPath)

	c.JSON(http.StatusOK, gin.H{
		"ok":        true,
		"directive": "alog",
		"id":        alogID,
		"log_name":  req.LogName,
		"status":    "started",
	})
}

func (a *API) handleAstop(c *gin.Context) {
	ctx, cancel := a.newAsyncDirectiveContext(c.Request.Context())
	defer cancel()

	log.Printf("astop requested")
	if err := a.mkpCtrl.StopRecordContext(ctx); err != nil {
		log.Printf("astop failed to send: err=%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	a.alogMu.Lock()
	if !a.alogRunning {
		a.alogMu.Unlock()
		log.Printf("astop sent without active alog session")
		c.JSON(http.StatusOK, gin.H{
			"ok":        true,
			"directive": "astop",
			"status":    "The stop command has been sent.",
		})
		return
	}

	stoppedID := a.currentAlogID
	stoppedStarted := a.alogStarted
	stoppedSession := a.currentAlog
	a.alogRunning = false
	a.alogStarted = false
	a.currentAlogID = ""
	a.currentAlog = nil
	a.alogMu.Unlock()

	resp := gin.H{
		"ok":        true,
		"directive": "astop",
		"id":        stoppedID,
		"status":    "stopped",
	}
	if stoppedSession != nil && stoppedStarted {
		dbCtx, dbCancel := a.newDBContext(c.Request.Context())
		defer dbCancel()
		if record, err := a.persistMacroRecord(dbCtx, stoppedSession); err != nil {
			log.Printf("astop completed but persist failed: id=%s err=%v", stoppedID, err)
			resp["persisted"] = false
			resp["persist_error"] = err.Error()
		} else {
			log.Printf("astop completed: id=%s persisted=true macro_record_id=%d", stoppedID, record.ID)
			resp["persisted"] = true
			resp["macro_record_id"] = record.ID
		}
	} else if stoppedSession != nil {
		log.Printf("astop completed: id=%s persisted=false reason=alog session was not fully started", stoppedID)
		resp["persisted"] = false
		resp["persist_error"] = "alog session was not fully started"
	} else {
		log.Printf("astop completed: persisted=false reason=missing alog session state")
		resp["persisted"] = false
		resp["persist_error"] = "missing alog session state"
	}
	c.JSON(http.StatusOK, resp)
}

func (a *API) handleList(c *gin.Context) {
	ctx, cancel := a.newDBContext(c.Request.Context())
	defer cancel()
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

	query := a.db.WithContext(ctx)
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
	ctx, cancel := a.newDBContext(c.Request.Context())
	defer cancel()
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
	if err := a.db.WithContext(ctx).Where("unique_id = ?", rawUniqueID).First(&record).Error; err != nil {
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

func (a *API) handleDeleteRecord(c *gin.Context) {
	ctx, cancel := a.newDBContext(c.Request.Context())
	defer cancel()
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
	if err := a.db.WithContext(ctx).Where("unique_id = ?", rawUniqueID).First(&record).Error; err != nil {
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

	a.replayMu.Lock()
	isReplaying, retryAfter := a.isReplayBlockingLocked(time.Now(), record.UniqueID)
	a.replayMu.Unlock()
	if isReplaying {
		resp := gin.H{
			"ok":    false,
			"error": "record is currently being replayed",
		}
		if retryAfter > 0 {
			resp["retry_after_seconds"] = int(retryAfter.Seconds()) + 1
		}
		c.JSON(http.StatusConflict, resp)
		return
	}

	resp := gin.H{
		"ok":        true,
		"directive": "record_delete",
		"id":        record.UniqueID,
	}

	if record.MKPPath != "" {
		serialCtx, serialCancel := a.newSyncDirectiveContext(c.Request.Context())
		defer serialCancel()
		if err := a.mkpCtrl.DeleteFileContext(serialCtx, record.MKPPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"ok":    false,
				"error": err.Error(),
			})
			return
		}
		resp["file_deleted"] = true
	}

	if err := a.db.WithContext(ctx).Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (a *API) handleAplay(c *gin.Context) {
	dbCtx, dbCancel := a.newDBContext(c.Request.Context())
	defer dbCancel()
	if a.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": "database is not initialized",
		})
		return
	}

	a.replayMu.Lock()
	if a.replayRunning {
		if blocking, remaining := a.isReplayBlockingLocked(time.Now(), ""); blocking {
			currentID := a.currentReplayID
			a.replayMu.Unlock()
			log.Printf("aplay rejected: current_replay_id=%s retry_after=%s", currentID, remaining)
			c.JSON(http.StatusConflict, gin.H{
				"ok":                  false,
				"error":               "replay is already running",
				"retry_after_seconds": int(remaining.Seconds()) + 1,
				"estimated_done_at":   a.replayDoneAt,
				"release_after":       a.replayReleaseAt,
			})
			return
		}

		a.resetReplayStateLocked()
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

	log.Printf("aplay requested: id=%s", rawUniqueID)

	var record models.MacroRecord
	if err := a.db.WithContext(dbCtx).Where("unique_id = ?", rawUniqueID).First(&record).Error; err != nil {
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

	if record.StartPointX >= 0 && record.StartPointY >= 0 {
		robotgo.Move(record.StartPointX, record.StartPointY)
	} else {
		robotgo.Move(0, 0)
	}

	serialCtx, serialCancel := a.newAsyncDirectiveContext(c.Request.Context())
	defer serialCancel()
	if err := a.sfport.StartReplayingContext(serialCtx, record.MKPPath, -1); err != nil {
		log.Printf("aplay failed: id=%s path=%s err=%v", record.UniqueID, record.MKPPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"ok":    false,
			"error": err.Error(),
		})
		return
	}

	startedAt := time.Now()
	finishAt, releaseAt := a.replayWindowForRecord(record, startedAt)

	a.replayMu.Lock()
	a.replayRunning = true
	a.currentReplayID = record.UniqueID
	a.replayStartedAt = startedAt
	a.replayDoneAt = finishAt
	a.replayReleaseAt = releaseAt
	if a.replayTimer != nil {
		a.replayTimer.Stop()
	}
	replayID := record.UniqueID
	a.replayTimer = time.AfterFunc(time.Until(releaseAt), func() {
		a.replayMu.Lock()
		defer a.replayMu.Unlock()
		if a.currentReplayID == replayID {
			a.resetReplayStateLocked()
		}
	})
	a.replayMu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"ok":                true,
		"directive":         "aplay",
		"id":                record.UniqueID,
		"mkp_path":          record.MKPPath,
		"status":            "started",
		"estimated_done_at": finishAt,
		"release_after":     releaseAt,
	})

	log.Printf("aplay started: id=%s path=%s estimated_done_at=%s release_after=%s", record.UniqueID, record.MKPPath, finishAt.Format(time.RFC3339Nano), releaseAt.Format(time.RFC3339Nano))
}

func (a *API) handleAcancel(c *gin.Context) {
	ctx, cancel := a.newAsyncDirectiveContext(c.Request.Context())
	defer cancel()

	log.Printf("acancel requested")
	if err := a.mkpCtrl.CancelContext(ctx); err != nil {
		log.Printf("acancel failed: err=%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "error": err.Error()})
		return
	}

	a.replayMu.Lock()
	canceledID := a.currentReplayID
	wasRunning := a.replayRunning
	a.resetReplayStateLocked()
	a.replayMu.Unlock()

	log.Printf("acancel sent: id=%s was_running=%t", canceledID, wasRunning)

	c.JSON(http.StatusOK, gin.H{
		"ok":          true,
		"directive":   "acancel",
		"id":          canceledID,
		"was_running": wasRunning,
		"status":      "cancel_sent",
	})
}

func (a *API) persistMacroRecord(ctx context.Context, session *alogSession) (*models.MacroRecord, error) {
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

	logLength, err := a.mkpCtrl.AtimeContext(ctx, session.MKPPath)
	if err == nil && logLength != nil {
		record.Seconds = logLength.Seconds
		record.Milliseconds = logLength.Milliseconds
	}
	if err := a.db.WithContext(ctx).Create(record).Error; err != nil {
		return nil, err
	}
	return record, nil
}

func (a *API) runtimeConfig() servercfg.RuntimeConfig {
	return servercfg.NormalizeRuntimeConfig(a.runtime)
}

func (a *API) newSyncDirectiveContext(parent context.Context) (context.Context, context.CancelFunc) {
	timeout := a.runtimeConfig().SyncDirectiveTimeout + serialSyncTimeoutGrace
	if a.sfport != nil && a.sfport.SyncOutputTimeout > 0 {
		timeout = a.sfport.SyncOutputTimeout + serialSyncTimeoutGrace
	}
	return withTimeoutIfNone(parent, timeout)
}

func (a *API) newAsyncDirectiveContext(parent context.Context) (context.Context, context.CancelFunc) {
	return withTimeoutIfNone(parent, a.runtimeConfig().AsyncDirectiveTimeout)
}

func (a *API) newDBContext(parent context.Context) (context.Context, context.CancelFunc) {
	return withTimeoutIfNone(parent, a.runtimeConfig().DBTimeout)
}

func withTimeoutIfNone(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return parent, func() {}
	}
	if _, hasDeadline := parent.Deadline(); hasDeadline {
		return parent, func() {}
	}
	return context.WithTimeout(parent, timeout)
}

func (a *API) replayWindowForRecord(record models.MacroRecord, startedAt time.Time) (time.Time, time.Time) {
	runtimeCfg := a.runtimeConfig()
	duration := time.Duration(record.Seconds)*time.Second + time.Duration(record.Milliseconds)*time.Millisecond
	if duration < runtimeCfg.ReplayMinimumDuration {
		duration = runtimeCfg.ReplayMinimumDuration
	}
	finishAt := startedAt.Add(duration)
	return finishAt, finishAt.Add(runtimeCfg.ReplayCompletionGrace)
}

func (a *API) resetReplayStateLocked() {
	a.replayRunning = false
	a.currentReplayID = ""
	a.replayStartedAt = time.Time{}
	a.replayDoneAt = time.Time{}
	a.replayReleaseAt = time.Time{}
	if a.replayTimer != nil {
		a.replayTimer.Stop()
		a.replayTimer = nil
	}
}

func (a *API) isReplayBlockingLocked(now time.Time, recordID string) (bool, time.Duration) {
	if !a.replayRunning {
		return false, 0
	}
	if recordID != "" && a.currentReplayID != "" && a.currentReplayID != recordID {
		return false, 0
	}
	if !a.replayReleaseAt.IsZero() && now.Before(a.replayReleaseAt) {
		return true, a.replayReleaseAt.Sub(now)
	}
	return false, 0
}

func (a *API) snapshotAlogState() gin.H {
	a.alogMu.Lock()
	defer a.alogMu.Unlock()
	return gin.H{
		"running":    a.alogRunning,
		"started":    a.alogStarted,
		"current_id": a.currentAlogID,
	}
}

func (a *API) snapshotReplayState() gin.H {
	now := time.Now()
	a.replayMu.Lock()
	defer a.replayMu.Unlock()
	blocking, retryAfter := a.isReplayBlockingLocked(now, "")
	state := gin.H{
		"running":           a.replayRunning,
		"current_id":        a.currentReplayID,
		"started_at":        zeroTimeToNil(a.replayStartedAt),
		"estimated_done_at": zeroTimeToNil(a.replayDoneAt),
		"release_after":     zeroTimeToNil(a.replayReleaseAt),
		"active":            blocking,
	}
	if retryAfter > 0 {
		state["retry_after_seconds"] = int(retryAfter.Seconds()) + 1
	}
	return state
}

func zeroTimeToNil(t time.Time) interface{} {
	if t.IsZero() {
		return nil
	}
	return t
}
