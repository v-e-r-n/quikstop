package quikstop

import (
	"github.com/v-e-r-n/quikstop/cors"
	"github.com/v-e-r-n/quikstop/db"
	"github.com/v-e-r-n/quikstop/events"
	"github.com/v-e-r-n/quikstop/httputil"
	"github.com/v-e-r-n/quikstop/jwt"
	"github.com/v-e-r-n/quikstop/limiter"
	"github.com/v-e-r-n/quikstop/mcfeely"
	"github.com/v-e-r-n/quikstop/otp"
)

// -----------------------------------------------------------------------------
// Database (quikstop/db)
// -----------------------------------------------------------------------------

type DB = db.DB
type Dialect = db.Dialect
type Binder = db.Binder

const (
	DialectSQLite3  = db.DialectSQLite3
	DialectPostgres = db.DialectPostgres
)

var ConnectDB = db.Connect

// -----------------------------------------------------------------------------
// CORS (quikstop/cors)
// -----------------------------------------------------------------------------

type CORSConfig = cors.Config
type CORSOption = cors.Option

var (
	CORS            = cors.Handler
	WithCORSOrigins = cors.WithOrigins
	WithCORSMethods = cors.WithMethods
	WithCORSHeaders = cors.WithHeaders
)

// -----------------------------------------------------------------------------
// HTTP Utilities (quikstop/httputil)
// -----------------------------------------------------------------------------

type ErrorPayload = httputil.ErrorPayload

var (
	JSON       = httputil.JSON
	Error      = httputil.Error
	DecodeJSON = httputil.DecodeJSON
)

// -----------------------------------------------------------------------------
// JWT Authentication (quikstop/jwt)
// -----------------------------------------------------------------------------

type JWTClaims = jwt.Claims
type JWTMiddlewareOption = jwt.MiddlewareOption

var (
	ErrInvalidJWTToken = jwt.ErrInvalidToken
	ErrMissingJWTToken = jwt.ErrMissingToken

	GenerateJWT         = jwt.Generate
	VerifyJWT           = jwt.Verify
	AuthMiddleware      = jwt.Middleware
	WithScopeHeaders    = jwt.WithScopeHeaders
	WithScopeQueryParams = jwt.WithScopeQueryParams
	WithoutQueryToken   = jwt.WithoutQueryToken
	UserIDFromContext   = jwt.UserIDFromContext
	ScopeIDFromContext  = jwt.ScopeIDFromContext
)

// -----------------------------------------------------------------------------
// Rate Limiter (quikstop/limiter)
// -----------------------------------------------------------------------------

type Limiter = limiter.Limiter
type LimiterOption = limiter.Option

var (
	NewLimiter          = limiter.New
	RateLimitMiddleware = limiter.Handler
	GetClientIP         = limiter.GetClientIP
	WithOnLimit         = limiter.WithOnLimit
	WithCleanupInterval = limiter.WithCleanupInterval
)

// -----------------------------------------------------------------------------
// Real-time Events & SSE (quikstop/events)
// -----------------------------------------------------------------------------

type Event = events.Event
type EventType = events.EventType
type EventPayload = events.EventPayload
type EventMeta = events.EventMeta
type ClientSession = events.ClientSession
type Connection = events.Connection
type RecipientResolver = events.RecipientResolver
type EventBus = events.EventBus
type EventBusOption = events.EventBusOption
type StreamManager = events.StreamManager
type Dispatcher = events.Dispatcher
type SSEConfig = events.SSEConfig

var (
	NewEvent         = events.NewEvent
	NewEventBus      = events.NewEventBus
	NewStreamManager = events.NewStreamManager
	NewDispatcher    = events.NewDispatcher
	NewSSEHandler    = events.NewSSEHandler
	WithOnDrop       = events.WithOnDrop
)

// -----------------------------------------------------------------------------
// OTP / Passwordless (quikstop/otp)
// -----------------------------------------------------------------------------

type OTPKeeper = otp.Keeper
type OTPStore = otp.Store
type OTPDeliverer = otp.Deliverer
type OTPConfig = otp.Config
type OTPTokenMetadata = otp.TokenMetadata
type OTPHandler = otp.Handler
type OTPSuccessCallback = otp.SuccessCallback

var (
	ErrMaxRetriesExceeded = otp.ErrMaxRetriesExceeded
	ErrInvalidOTPCode     = otp.ErrInvalidCode
	ErrOTPNotFound        = otp.ErrNotFound
	ErrOTPExpired         = otp.ErrExpired

	NewOTP               = otp.New
	NewMemoryOTPStore    = otp.NewMemoryStore
	NewConsoleDeliverer  = otp.NewConsoleDeliverer
	NewOTPHandler        = otp.NewHandler
	SecureNumericGenerator = otp.SecureNumericGenerator
)

// -----------------------------------------------------------------------------
// Email Delivery (quikstop/mcfeely)
// -----------------------------------------------------------------------------

type McFeely = mcfeely.McFeely
type ConsoleMcFeely = mcfeely.ConsoleMcFeely
type SmtpMcFeely = mcfeely.SmtpMcFeely

var (
	NewConsoleMcFeely = mcfeely.NewConsoleMcFeely
	NewSmtpMcFeely    = mcfeely.NewSmtpMcFeely
)
