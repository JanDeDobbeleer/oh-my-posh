package daemon

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/config"
	"github.com/jandedobbeleer/oh-my-posh/src/daemon/ipc"
	"github.com/jandedobbeleer/oh-my-posh/src/log"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/prompt"
	"github.com/jandedobbeleer/oh-my-posh/src/shell"
	"github.com/jandedobbeleer/oh-my-posh/src/template"
	"github.com/jandedobbeleer/oh-my-posh/src/terminal"

	"google.golang.org/grpc"
)

// DefaultIdleTimeout is the default duration the daemon waits after all sessions end before shutting down.
// This allows users to close and reopen terminals without restarting the daemon.
// Can be overridden via the daemon_idle_timeout config option.
const DefaultIdleTimeout = 5 * time.Minute

// DefaultAsyncTimeout is how long to wait before returning partial results.
const DefaultAsyncTimeout = 100 * time.Millisecond

// sessionTextCache implements prompt.TextCache using the daemon's MemoryCache.
// It wraps the cache with a sessionID so each session has isolated cached values.
type sessionTextCache struct {
	cache     *MemoryCache
	sessionID string
}

func (s *sessionTextCache) Get(key string) (string, bool) {
	if cached, ok := s.cache.Get(s.sessionID, key); ok {
		if text, ok := cached.(string); ok {
			return text, true
		}
	}
	return "", false
}

func (s *sessionTextCache) Set(key, value string) {
	// Default: use AsyncRendering strategy with no expiration (session-scoped)
	s.cache.SetWithStrategy(s.sessionID, key, value, StrategyAsyncRendering, 0)
}

// GetWithAge retrieves cached text along with its age.
func (s *sessionTextCache) GetWithAge(key string) (string, time.Duration, bool) {
	entry, ok := s.cache.GetWithMetadata(s.sessionID, key)
	if !ok {
		return "", 0, false
	}
	if text, ok := entry.Value.(string); ok {
		return text, entry.Age(), true
	}
	return "", 0, false
}

// SetWithConfig stores text with explicit cache configuration.
func (s *sessionTextCache) SetWithConfig(key, value string, cacheConfig *config.Cache) {
	var strategy CacheStrategy
	var ttl time.Duration

	if cacheConfig == nil {
		// No config: use AsyncRendering (always recompute, cache for pending only)
		strategy = StrategyAsyncRendering
		ttl = 0 // No expiration within session
	} else {
		// Convert user strategy to daemon strategy
		strategy = ToDaemonStrategy(cacheConfig.Strategy)

		// Determine TTL based on strategy and duration
		switch {
		case strategy == StrategySession:
			// Session-scoped: no expiration (cleaned when session ends)
			ttl = 0
		case !cacheConfig.Duration.IsEmpty() && cacheConfig.Duration != cache.INFINITE:
			// Duration specified: use it
			ttl = time.Duration(cacheConfig.Duration.Seconds()) * time.Second
		default:
			// No duration or infinite: use daemon's default TTL
			ttl = s.cache.GetDefaultTTL()
		}
	}

	s.cache.SetWithStrategy(s.sessionID, key, value, strategy, ttl)
}

// ShouldRecompute determines if a segment should be recomputed based on its cache config.
func (s *sessionTextCache) ShouldRecompute(key string, cacheConfig *config.Cache) (recompute, useCacheForPending bool) {
	entry, found := s.cache.GetWithMetadata(s.sessionID, key)

	if !found {
		// No cached value: must recompute, nothing to show for pending
		return true, false
	}

	// Check if we have a valid cached text
	cachedText, isString := entry.Value.(string)
	hasValidCache := isString && cachedText != ""

	// No cache config: AsyncRendering behavior
	// Always recompute, use cache for pending display
	if cacheConfig == nil {
		return true, hasValidCache
	}

	// Check duration-based validation
	duration := cacheConfig.Duration

	// Empty duration or INFINITE with AsyncRendering strategy: always recompute
	if duration.IsEmpty() || duration == cache.INFINITE {
		// If original strategy was AsyncRendering, always recompute
		if entry.Strategy == StrategyAsyncRendering {
			return true, hasValidCache
		}
		// For user-configured INFINITE, don't recompute
		if duration == cache.INFINITE {
			return false, false
		}
		// Empty duration defaults to AsyncRendering behavior
		return true, hasValidCache
	}

	// Duration is specified: check if cache is still fresh
	maxAge := time.Duration(duration.Seconds()) * time.Second
	if entry.Age() <= maxAge {
		// Cache is fresh: don't recompute, use cached value directly
		return false, false
	}

	// Cache is stale: recompute, use old cache for pending display
	return true, hasValidCache
}

// Daemon is the background process that renders prompts.
type Daemon struct {
	ipc.UnimplementedDaemonServiceServer
	listener      net.Listener
	configWatcher *ConfigWatcher
	binaryWatcher *BinaryWatcher
	server        *grpc.Server
	lockFile      *LockFile
	cache         *MemoryCache
	configCache   *ConfigCache
	activeRenders *maps.Concurrent[context.CancelFunc]
	sessions      *SessionManager
	done          chan struct{}
	prototypePath string
	idleTimeout   time.Duration
	shutdownOnce  sync.Once
	mu            sync.Mutex
}

// New creates a new daemon instance.
// Acquires the lock file to ensure single instance.
func New(configPath string) (*Daemon, error) {
	// Acquire lock to ensure single instance
	lockFile, err := NewLockFile()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock: %w", err)
	}

	configCache := NewConfigCache()

	// Try to create file watcher - may fail on some systems
	configWatcher, err := NewConfigWatcher(configCache)
	if err != nil {
		log.Debugf("failed to create config watcher, file watching disabled: %v", err)
		configWatcher = nil
	}

	d := &Daemon{
		lockFile:      lockFile,
		cache:         NewMemoryCache(),
		configCache:   configCache,
		configWatcher: configWatcher,
		activeRenders: maps.NewConcurrent[context.CancelFunc](),
		done:          make(chan struct{}),
		prototypePath: configPath,
		idleTimeout:   DefaultIdleTimeout,
	}

	// Pre-load the prototype config and extract idle timeout
	cfg := d.getOrLoadConfig(configPath)
	d.idleTimeout = cfg.GetDaemonIdleTimeout()

	// Initialize session manager - starts idle timer when all sessions end
	d.sessions = NewSessionManager(func(pid int) {
		// Clean up memory cache for this session (PID)
		d.cache.CleanSession(fmt.Sprint(pid))
	}, d.startIdleTimer)

	// Start initial idle timer - will shut down if no sessions register
	d.startIdleTimer()

	// Try to create binary watcher for auto-restart on upgrade
	if binPath, err := os.Executable(); err == nil {
		bw, err := NewBinaryWatcher(binPath, func() {
			log.Debug("Binary changed on disk, shutting down for upgrade")
			d.shutdown()
		})
		if err != nil {
			log.Debugf("failed to create binary watcher, binary watching disabled: %v", err)
		} else {
			d.binaryWatcher = bw
		}
	}

	return d, nil
}

// Start begins serving gRPC requests.
// Blocks until shutdown is called or a signal is received.
func (d *Daemon) Start() error {
	// Create listener
	listener, err := ipc.Listen()
	if err != nil {
		_ = d.lockFile.Release()
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create gRPC server
	server := grpc.NewServer()
	ipc.RegisterDaemonServiceServer(server, d)

	// Store server and listener under lock
	d.mu.Lock()
	d.listener = listener
	d.server = server
	d.mu.Unlock()

	// Handle signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Debugf("Received signal: %v", sig)
		d.shutdown()
	}()

	// Start periodic cache cleanup
	go d.cacheCleanupLoop()

	log.Debug("Daemon starting on", ipc.SocketPath())

	// Serve blocks until Stop is called
	if err := server.Serve(listener); err != nil {
		// Server.Serve returns error when Stop/GracefulStop is called
		// This is expected behavior, not an error
		log.Debugf("Server stopped: %v", err)
	}

	return nil
}

// shutdown gracefully stops the daemon.
func (d *Daemon) shutdown() {
	d.shutdownOnce.Do(func() {
		log.Debug("Daemon shutting down")

		// Get server reference under lock, then stop outside lock
		// (GracefulStop may block waiting for RPCs to complete)
		d.mu.Lock()
		server := d.server
		d.mu.Unlock()

		// Stop accepting new connections and wait for existing RPCs to complete
		if server != nil {
			server.GracefulStop()
		}

		// Clean up socket
		if err := ipc.CleanupSocket(); err != nil {
			log.Debugf("Failed to cleanup socket: %v", err)
		}

		// Close config watcher
		if d.configWatcher != nil {
			if err := d.configWatcher.Close(); err != nil {
				log.Debugf("Failed to close config watcher: %v", err)
			}
		}

		// Close binary watcher
		if d.binaryWatcher != nil {
			if err := d.binaryWatcher.Close(); err != nil {
				log.Debugf("Failed to close binary watcher: %v", err)
			}
		}

		// Release lock file
		if d.lockFile != nil {
			if err := d.lockFile.Release(); err != nil {
				log.Debugf("Failed to release lock: %v", err)
			}
		}

		close(d.done)
	})
}

// Done returns a channel that is closed when the daemon has stopped.
func (d *Daemon) Done() <-chan struct{} {
	return d.done
}

// cacheCleanupLoop periodically evicts expired cache entries.
func (d *Daemon) cacheCleanupLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.cache.EvictExpired()
		case <-d.done:
			return
		}
	}
}

// startIdleTimer starts an idle shutdown timer.
// When it fires, it checks if there are active sessions.
// If not, it shuts down. If yes, it does nothing.
// If idleTimeout is 0, no timer is started (daemon never exits due to idle).
func (d *Daemon) startIdleTimer() {
	if d.idleTimeout == 0 {
		log.Debug("Idle timeout disabled, daemon will not auto-shutdown")
		return
	}

	time.AfterFunc(d.idleTimeout, func() {
		if d.sessions.Count() == 0 {
			log.Debug("Idle timeout reached with no active sessions, shutting down")
			d.shutdown()
		}
	})
}

// RenderPrompt handles prompt rendering requests.
// Streams updates as segments complete.
// A new request for the same session automatically cancels any in-progress render.
func (d *Daemon) RenderPrompt(req *ipc.PromptRequest, stream ipc.DaemonService_RenderPromptServer) error {
	log.Debugf("RenderPrompt: session=%s, requestId=%s", req.SessionId, req.RequestId)

	// Register session by PID for process exit tracking
	// We do this FIRST to ensure the session is registered before any async work or potential cancellation
	if req.Pid > 0 {
		// Convert flags early to get Shell
		flags := ipc.ProtoToFlags(req.Flags)
		// Extract UUID from Env if present
		uuid := req.Env["POSH_SESSION_ID"]
		d.sessions.Register(int(req.Pid), uuid, flags.Shell)
	}

	// Validate protocol version
	if req.Version != ipc.ProtocolVersion {
		return fmt.Errorf("protocol version mismatch: client=%d, server=%d", req.Version, ipc.ProtocolVersion)
	}

	// Cancel any existing render for this session
	if existingCancel, ok := d.activeRenders.Get(req.SessionId); ok {
		existingCancel()
	}

	// Create cancellable context for this request
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	// Register this render's cancel function
	d.activeRenders.Set(req.SessionId, cancel)
	defer d.activeRenders.Delete(req.SessionId)

	// Get or load config (Prototype) from the daemon's configured path
	// We ignore req.Config as the daemon operates on a single master config
	prototypeConfig := d.getOrLoadConfig(d.prototypePath)

	// CLONE the config for this request to ensure isolation
	cfg := prototypeConfig.Clone()

	// Convert proto flags to runtime flags
	flags := ipc.ProtoToFlags(req.Flags)

	// Create environment with request's env vars
	env := NewEnvironment(flags, req.Env)

	// Initialize template system (for non-cache globals like funcMap)
	template.Init(env, cfg.Var, cfg.Maps)

	// Create per-request template cache to avoid sharing segment data between requests
	templateCache := template.NewCache(env, cfg.Var, cfg.Maps)

	// Initialize terminal package-level state for functions that read Shell/Program/formats
	sh := env.Shell()
	if sh == shell.BASH && !flags.Escape {
		sh = shell.GENERIC
	}
	terminal.Init(sh)

	// Create per-request Writer — this is the key to concurrent rendering.
	// Each request gets its own builder, so no mutex is needed.
	writer := terminal.NewWriter(sh)
	writer.BackgroundColor = cfg.TerminalBackground.ResolveTemplate()
	writer.Colors = cfg.MakeColors(env)
	writer.Plain = flags.Plain

	// Create text cache wrapper for this session
	textCache := &sessionTextCache{
		cache:     d.cache,
		sessionID: req.SessionId,
	}

	// Create engine with per-request template cache, text cache, and writer
	eng := &prompt.Engine{
		Config:        cfg,
		Env:           env,
		Writer:        writer,
		Plain:         flags.Plain,
		TemplateCache: templateCache,
		TextCache:     textCache,
	}

	// Track pending segments for streaming updates
	var pendingRemaining atomic.Int32
	pendingDone := make(chan struct{})

	// Callback when a segment completes after timeout.
	// Called from background goroutines in PrimaryStreaming.
	// Safe without mutex: each request has its own Writer, and updates
	// for a single request are serialized by the results channel drain.
	updateCallback := func(segmentName string) {
		newPrompt := eng.ReRender()

		// Cache the segment's text
		for _, block := range cfg.Blocks {
			for _, segment := range block.Segments {
				if segment.Name() == segmentName {
					eng.CacheSegmentText(segment)
					break
				}
			}
		}

		prompts := d.buildPrompts(eng, newPrompt)

		// Send update
		select {
		case <-ctx.Done():
		default:
			response := &ipc.PromptResponse{
				Type:      "update",
				RequestId: req.RequestId,
				Prompts:   prompts,
			}
			if err := stream.Send(response); err != nil {
				log.Debugf("failed to send update: %v", err)
			}
		}

		if pendingRemaining.Add(-1) == 0 {
			close(pendingDone)
		}
	}

	// Render prompt with streaming
	asyncTimeout := DefaultAsyncTimeout
	if cfg.DaemonTimeout > 0 {
		asyncTimeout = time.Duration(cfg.DaemonTimeout) * time.Millisecond
	}
	log.Debugf("RenderPrompt: starting PrimaryStreaming with timeout=%v", asyncTimeout)
	promptText := eng.PrimaryStreaming(asyncTimeout, updateCallback)

	// Count pending segments
	eng.StreamingMu().Lock()
	pending := int32(len(eng.PendingSegments()))
	eng.StreamingMu().Unlock()
	pendingRemaining.Store(pending)
	log.Debugf("RenderPrompt: initial render done, pending=%d", pending)

	// Build initial response
	responseType := ResponseTypeComplete
	if pending > 0 {
		responseType = "update"
	}

	response := &ipc.PromptResponse{
		Type:      responseType,
		RequestId: req.RequestId,
		Prompts:   d.buildPrompts(eng, promptText),
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if err := stream.Send(response); err != nil {
			return err
		}
	}

	// If there are pending segments, wait for them to complete
	if pending > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-pendingDone:
			// All segments completed — send final response
			finalPrompt := eng.ReRender()
			finalPrompts := d.buildPrompts(eng, finalPrompt)

			finalResponse := &ipc.PromptResponse{
				Type:      ResponseTypeComplete,
				RequestId: req.RequestId,
				Prompts:   finalPrompts,
			}
			return stream.Send(finalResponse)
		}
	}

	return nil
}

// getOrLoadConfig retrieves a config from cache or loads and caches it.
// If configPath is empty, it returns the default config.
func (d *Daemon) getOrLoadConfig(configPath string) *config.Config {
	// Check cache first
	if cached, ok := d.configCache.Get(configPath); ok {
		log.Debugf("config cache hit: %s", configPath)
		return cached.Config
	}

	log.Debugf("config cache miss: %s", configPath)

	// Load config (falls back to default if path is empty or invalid)
	cfg := config.Load(configPath)

	// Only cache if loading actually succeeded (Source matches configPath)
	// or if we explicitly wanted the default config (empty path).
	if configPath == "" || cfg.Source == configPath {
		d.configCache.Set(configPath, cfg, cfg.FilePaths)

		// Setup file watching (for local files only)
		if d.configWatcher != nil && !strings.HasPrefix(configPath, "http") && len(cfg.FilePaths) > 0 {
			if err := d.configWatcher.Watch(configPath, cfg.FilePaths); err != nil {
				log.Debugf("failed to watch config files: %v", err)
			}
		}
	}

	return cfg
}

// buildPrompts creates the prompts map with all configured prompt types.
// Uses StreamingRPrompt() to avoid re-executing rprompt segments — they are
// already rendered during PrimaryStreaming/ReRender.
func (d *Daemon) buildPrompts(eng *prompt.Engine, primaryText string) map[string]*ipc.Prompt {
	prompts := map[string]*ipc.Prompt{
		"primary": {Text: primaryText},
		"right":   {Text: eng.StreamingRPrompt()},
	}

	// Add secondary prompt if configured
	if eng.Config.SecondaryPrompt != nil {
		prompts["secondary"] = &ipc.Prompt{Text: eng.ExtraPrompt(prompt.Secondary)}
	}

	// Add transient prompt if configured
	if eng.Config.TransientPrompt != nil {
		prompts["transient"] = &ipc.Prompt{Text: eng.ExtraPrompt(prompt.Transient)}
	}

	return prompts
}

// ToggleSegment toggles the visibility of segments for a session.
func (d *Daemon) ToggleSegment(_ context.Context, req *ipc.ToggleSegmentRequest) (*ipc.ToggleSegmentResponse, error) {
	if req.SessionId == "" {
		return &ipc.ToggleSegmentResponse{Success: false, Error: "missing session_id"}, nil
	}

	// Get current toggles
	var currentToggleSet map[string]bool
	if cached, ok := d.cache.Get(req.SessionId, cache.TOGGLECACHE); ok {
		if cm, ok := cached.(map[string]bool); ok {
			currentToggleSet = cm
		}
	}

	if currentToggleSet == nil {
		currentToggleSet = make(map[string]bool)
	}

	// Toggle segments: remove if present, add if not present
	for _, segment := range req.Segments {
		if currentToggleSet[segment] {
			delete(currentToggleSet, segment)
		} else {
			currentToggleSet[segment] = true
		}
	}

	d.cache.Set(req.SessionId, cache.TOGGLECACHE, currentToggleSet, 0)
	return &ipc.ToggleSegmentResponse{Success: true}, nil
}

// CacheClear clears all daemon cache entries.
func (d *Daemon) CacheClear(_ context.Context, _ *ipc.CacheClearRequest) (*ipc.CacheClearResponse, error) {
	d.cache.ClearAll()
	return &ipc.CacheClearResponse{Success: true}, nil
}

// CacheSetTTL sets the default cache TTL (in days).
func (d *Daemon) CacheSetTTL(_ context.Context, req *ipc.CacheSetTTLRequest) (*ipc.CacheSetTTLResponse, error) {
	ttl := time.Duration(req.Days) * 24 * time.Hour
	d.cache.SetDefaultTTL(ttl)
	return &ipc.CacheSetTTLResponse{Success: true}, nil
}

// CacheGetTTL gets the current default cache TTL (in days).
func (d *Daemon) CacheGetTTL(_ context.Context, _ *ipc.CacheGetTTLRequest) (*ipc.CacheGetTTLResponse, error) {
	days := int32(d.cache.GetDefaultTTL().Hours() / 24)
	return &ipc.CacheGetTTLResponse{Days: days}, nil
}

// SetLogging enables or disables file logging on the running daemon.
func (d *Daemon) SetLogging(_ context.Context, req *ipc.SetLoggingRequest) (*ipc.SetLoggingResponse, error) {
	if req.Path == "" {
		log.DisableFileLogging()
		return &ipc.SetLoggingResponse{Success: true}, nil
	}

	if err := log.EnableFileLogging(req.Path); err != nil {
		return &ipc.SetLoggingResponse{Success: false, Error: err.Error()}, nil
	}

	return &ipc.SetLoggingResponse{Success: true}, nil
}
