package config

import (
	"log"
	"net/http"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Env       string // development | release | production
	Server    *ServerConfig
	Auth      *AuthConfig
	Database  *DatabaseConfig
	Judge     *JudgeConfig
	Docker    *DockerConfig
	Judgehost *JudgehostConfig
	Sonar     *SonarConfig
	Test      *TestConfig
	Log       *LogConfig
}

func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

type LogConfig struct {
	Level string // debug | info | warn | error  (default: info)
}

type ServerConfig struct {
	Port           int
	AppKey         string
	JwtKey         string
	StoreKey       string
	CsrfKey        string
	RdbDriver      string
	StoragePath    string
	AllowedOrigins []string
	// MaxUploadSizeMB caps a single request body (default 100). An upload is
	// spooled to a temp file and then read whole into memory, so without a
	// cap one request can exhaust either.
	MaxUploadSizeMB int
	// MaxUnzipSizeMB caps the total bytes one archive may extract to
	// (default 1024). Counted as bytes are written, never taken from the zip
	// header, which a bomb understates freely.
	MaxUnzipSizeMB int
	// MaxUnzipEntries caps how many entries one archive may contain
	// (default 50000).
	MaxUnzipEntries int
	// BundleImportMaxUploadSizeMB caps the two bundle-import endpoints
	// (default 2048). They get their own cap because an archive-mode image
	// bundle carries docker save output — hundreds of megabytes is normal —
	// and raising the global cap to fit would hand the same allowance to
	// every upload endpoint at once, each of which reads the file whole into
	// memory.
	BundleImportMaxUploadSizeMB int
	// MetricsHost is the interface the Prometheus metrics server (:9091) binds
	// to. Defaults to 127.0.0.1 so /metrics — and, outside production, the
	// pprof endpoints — are not reachable over the network by default. A
	// deployment that scrapes metrics from another host or container sets this
	// to 0.0.0.0 (the monitoring compose does).
	MetricsHost string
	// TrustedProxies is the CIDR/IP allowlist of reverse proxies whose
	// X-Forwarded-For header may be believed. EMPTY (the default) means trust
	// none, so the login rate limiter keys on the real TCP peer (RemoteAddr) and
	// a client cannot spoof a fresh IP per request to sidestep it. Set this to
	// your proxy's address ONLY when the API is actually behind one, so the real
	// client IP is recovered instead of every request looking like the proxy.
	TrustedProxies []string
	// LoginRateLimitBurst / LoginRateLimitPerMinute size the per-(client IP,
	// username) failed-login bucket. Only FAILED logins spend
	// tokens, so a room logging in successfully at once behind one NAT costs
	// nothing, and the long-term failed-guess rate equals perMinute regardless of
	// how many requests run in parallel. Because failures are cheap to bound
	// without hurting anyone, the default is deliberately strict (burst 30 for
	// finger-slips, then 6/min ≈ one guess every 10s — ~28h to exhaust a 4-digit
	// PIN).
	//
	// Pointers for the same three-state reason the container caps are: absent
	// (nil) applies the default; an explicit 0 (or any non-positive value) is
	// "off", disabling the limiter — the off/unlimited spelling shared with the
	// pids/memory caps and the repair intervals (the upload caps differ: there,
	// 0 means "use the default" and only a negative value disables). With a plain int, 0 would be indistinguishable
	// from an absent key, so an operator who wrote 0 to turn the limit OFF would
	// silently get the default limit instead.
	LoginRateLimitBurst     *int
	LoginRateLimitPerMinute *int
	// LoginRateLimitIpBurst / LoginRateLimitIpPerMinute size the looser per-IP
	// backstop that catches a single source spraying many usernames (which the
	// per-username buckets alone would not, since each fresh username starts a
	// fresh bucket). It is shared by everyone behind that IP, so its burst must
	// absorb a whole room's finger-slips at contest start (default 300) while its
	// long-term rate stays low enough to stop a spray (default 60/min). Size the
	// burst by your largest single source's head count. Same three-state pointer
	// semantics as above; non-positive disables the backstop.
	LoginRateLimitIpBurst     *int
	LoginRateLimitIpPerMinute *int
	// LoginRateLimitExemptCIDRs are IP ranges whose clients skip the login
	// limiter entirely — an explicit escape hatch for a trusted contest LAN.
	// Empty by default. It is a convenience, not a defence: it should never be
	// the only thing standing between the login endpoint and the internet.
	LoginRateLimitExemptCIDRs []string
}

// applyDefaults fills the upload and extraction limits.
//
// config.yaml is gitignored, so a deployment whose config predates these
// keys never receives them. An absent key therefore has to mean "use the
// default" — the alternative silently leaves production unlimited, which is
// the whole thing these limits exist to prevent. A NEGATIVE value is the
// explicit opt-out for turning a limit off in a hurry.
func (c *ServerConfig) applyDefaults() {
	if c.MaxUploadSizeMB == 0 {
		c.MaxUploadSizeMB = 100
	}
	if c.MaxUnzipSizeMB == 0 {
		c.MaxUnzipSizeMB = 1024
	}
	if c.MaxUnzipEntries == 0 {
		c.MaxUnzipEntries = 50000
	}
	if c.BundleImportMaxUploadSizeMB == 0 {
		c.BundleImportMaxUploadSizeMB = 2048
	}
	if c.MetricsHost == "" {
		c.MetricsHost = "127.0.0.1"
	}
	// Only an ABSENT (nil) value gets the fallback; an explicit 0 or negative is
	// the operator disabling the login limiter and must survive.
	if c.LoginRateLimitBurst == nil {
		defaultBurst := 30
		c.LoginRateLimitBurst = &defaultBurst
	}
	if c.LoginRateLimitPerMinute == nil {
		defaultPerMinute := 6
		c.LoginRateLimitPerMinute = &defaultPerMinute
	}
	if c.LoginRateLimitIpBurst == nil {
		defaultIpBurst := 300
		c.LoginRateLimitIpBurst = &defaultIpBurst
	}
	if c.LoginRateLimitIpPerMinute == nil {
		defaultIpPerMinute := 60
		c.LoginRateLimitIpPerMinute = &defaultIpPerMinute
	}
}

// MaxUploadBytes is MaxUploadSizeMB in bytes. A non-positive result means
// the limit is disabled.
func (c *ServerConfig) MaxUploadBytes() int64 {
	if c.MaxUploadSizeMB <= 0 {
		return -1
	}
	return int64(c.MaxUploadSizeMB) * 1024 * 1024
}

// BundleImportMaxUploadBytes is BundleImportMaxUploadSizeMB in bytes. A
// non-positive result means the limit is disabled.
func (c *ServerConfig) BundleImportMaxUploadBytes() int64 {
	if c.BundleImportMaxUploadSizeMB <= 0 {
		return -1
	}
	return int64(c.BundleImportMaxUploadSizeMB) * 1024 * 1024
}

// MaxUnzipBytes is MaxUnzipSizeMB in bytes. A non-positive result means the
// limit is disabled.
func (c *ServerConfig) MaxUnzipBytes() int64 {
	if c.MaxUnzipSizeMB <= 0 {
		return -1
	}
	return int64(c.MaxUnzipSizeMB) * 1024 * 1024
}

type AuthConfig struct {
	AccessTokenTTLSeconds  int
	RefreshTokenTTLSeconds int
	Cookie                 *AuthCookieConfig
}

// applyDefaults fills the token lifetimes and cookie attributes.
//
// These keys existed unread for a long time while the lifetimes were
// hardcoded to an hour and a week. config.yaml is gitignored, so a
// deployment predating them never receives them — an absent key therefore
// has to mean "keep what was in force", not "zero-second tokens".
// hardenCookies enforces the two auth-cookie invariants that are unsafe to
// leave to the operator's config file: a SameSite=None cookie is silently
// dropped by browsers unless it is also Secure, and a production deployment
// must never transmit session cookies without Secure.
func (c *Config) hardenCookies() {
	if c.Auth == nil || c.Auth.Cookie == nil {
		return
	}
	if strings.EqualFold(c.Auth.Cookie.SameSite, "None") && !c.Auth.Cookie.Secure {
		c.Auth.Cookie.Secure = true
	}
	if c.IsProduction() && !c.Auth.Cookie.Secure {
		log.Println("[config] production requires Secure auth cookies; forcing cookie.secure=true")
		c.Auth.Cookie.Secure = true
	}
}

func (c *AuthConfig) applyDefaults() {
	if c.AccessTokenTTLSeconds <= 0 {
		c.AccessTokenTTLSeconds = 3600
	}
	if c.RefreshTokenTTLSeconds <= 0 {
		c.RefreshTokenTTLSeconds = 604800
	}

	// Every SetCookie call reads this block, so a config omitting it would
	// panic on the first login rather than fall back.
	if c.Cookie == nil {
		c.Cookie = &AuthCookieConfig{}
	}
	if c.Cookie.Path == "" {
		c.Cookie.Path = "/"
	}
	if c.Cookie.SameSite == "" {
		c.Cookie.SameSite = "Lax"
	}
}

type AuthCookieConfig struct {
	Domain string
	Path   string
	Secure bool
	// SameSite is "Lax", "None" or "Strict". Note "None" only takes effect
	// in browsers when Secure is also true.
	SameSite string
}

// SameSiteMode maps the configured string onto the net/http mode gin's
// SetSameSite expects. Anything unrecognised falls back to Lax — the
// narrower choice, since guessing None would widen the cookie's exposure
// on the strength of a typo.
func (c *AuthCookieConfig) SameSiteMode() http.SameSite {
	switch strings.ToLower(c.SameSite) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

// JudgehostConfig is NOT consumed by any code in this repository — it is a
// leftover shape kept only so existing config files with a judgehost section
// keep parsing. The example config no longer documents it; do not add keys
// here expecting behavior.
type JudgehostConfig struct {
	Port          int
	Image         string
	Network       string
	ProjectVolume string
	DockerVolume  string
	StorageVolume string
	EnvFile       string
	Replicas      int
	RegisterURL   string
	Label         string
}

type MysqlConfig struct {
	Host         string
	Port         int
	Username     string
	Password     string
	DBName       string
	MaxOpenConns int
	MaxIdleConns int
	// ConnMaxLifetimeMinutes bounds how long a pooled connection may be
	// reused. MySQL's wait_timeout (commonly 8 hours) closes idle
	// connections server-side; without a lifetime cap the pool can hand out
	// an already-dead connection after a quiet period (e.g. the night
	// before a contest) and fail with "invalid connection". A nil pointer
	// means the key is absent and takes the default (5 minutes); an
	// explicit value <= 0 disables the cap; a positive value is minutes.
	// Same three-state convention as PidsLimit / MemoryLimitMB.
	ConnMaxLifetimeMinutes *int
}

// applyDefaults fills the connection-pool knobs that are unsafe to leave
// unbounded. Only nil (key absent) falls back — an explicit value, including
// zero, survives.
func (c *MysqlConfig) applyDefaults() {
	if c.ConnMaxLifetimeMinutes == nil {
		minutes := 5
		c.ConnMaxLifetimeMinutes = &minutes
	}
}

type MongoConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type JudgeConfig struct {
	Host     string
	Port     int
	Replicas int
	// OrphanSubmissionRepairInterval (seconds) paces the background sweep
	// that attaches a terminal ERROR judgment to submissions whose judgment
	// was never created. Three states, same convention as PidsLimit: nil
	// (key absent) → 300; an explicit value <= 0 disables the periodic
	// sweep (the startup sweep still runs); > 0 → that many seconds.
	OrphanSubmissionRepairInterval *int
	// OrphanJudgmentRepairInterval (seconds) paces the background sweep that
	// fails judgments stuck in Judging with no live worker (server restart
	// or worker crash mid-judgment). It used to run only at startup, which
	// meant a judgment orphaned MID-CONTEST stayed stuck until someone
	// restarted the server. Same three states as above. Only safe with a
	// SINGLE API instance: the sweep can only see the workers of its own
	// process, so a second instance's in-flight judgments look orphaned from
	// there and would be wrongly failed.
	OrphanJudgmentRepairInterval *int
	// RejudgeCommitRepairInterval (seconds) paces the background replay of
	// approvals whose judgments flip was lost mid-commit (rejudge APPROVED,
	// judgments still Rejudged). Same three states, same mid-contest
	// rationale as the sweeps above.
	RejudgeCommitRepairInterval *int
	ContainerLabel               string
	TestcaseTimeoutSeconds         int // per-testcase exec timeout (default 60)
	LifecycleTimeoutSeconds        int // per-lifecycle exec timeout (default 120)
	HealthcheckTimeoutSeconds      int // wait for project container healthy (default 300)
	MaxJudgeDurationMinutes        int // worker emergency timeout per judgment (default 30)
	WorkerCount                    int // number of judge workers in the pool (default 8)
	// SubmissionCooldownSeconds is the default minimum gap between one
	// contestant's submissions, used when a contest project does not set its
	// own. 0 disables the wait. It is a contest-fairness knob as much as a
	// load control: without it a contestant can use the judge as a
	// brute-force oracle, and each attempt costs two containers.
	SubmissionCooldownSeconds int
	// PidsLimit and MemoryLimitMB are the default container caps, used when a
	// contest project does not set its own. Both read the same way, and the
	// same way as the contest project columns they back:
	//
	//	absent/null  use the built-in default (1024 for pids, off for memory)
	//	0            no cap
	//	> 0          the cap
	//
	// They are pointers so that 0 can mean "no cap" here exactly as it does in
	// the API and the admin UI. With a plain int the two spellings would have
	// to diverge — 0 would be indistinguishable from an absent key, and an
	// absent key cannot mean "no cap" for PidsLimit: config.yaml is gitignored,
	// so no existing deployment's file contains these keys, and reading their
	// absence as "off" would leave every upgraded deployment uncapped, which is
	// the state the caps exist to fix. A negative value is also accepted as
	// "no cap", which is what an earlier revision of this comment told
	// operators to write.
	//
	// PidsLimit maps to the cgroup pids controller, which counts TASKS —
	// threads included, not just processes — so the realistic ceiling for a
	// web toolchain is in the hundreds rather than the dozens a process count
	// would suggest: a chromium run is a handful of processes but a few
	// hundred threads, and a sonar-scanner JVM is one process with dozens.
	// 1024 still clears that, and a fork bomb passes it in milliseconds.
	// Raising it is bounded from above by WorkerCount * 2 containers * this
	// value against the host's threads-max: at the default 8 workers, 1024
	// tops out at ~16k tasks, which is the point of the cap.
	PidsLimit *int
	// MemoryLimitMB defaults to off, unlike PidsLimit, because it is the one
	// with a realistic false positive — a sonar-scanner JVM or a chromium run
	// will take a gigabyte without misbehaving. Turning it on is a decision an
	// operator makes after checking their heaviest problem, not one taken for
	// them by an upgrade.
	MemoryLimitMB *int
	// MaxOutputBytesPerExec caps how many bytes of stdout+stderr are captured
	// from a single command run in the judge. Beyond it the output is truncated
	// with a marker rather than buffered in full, so a submission that prints
	// without bound cannot exhaust the worker's memory or the judgment document
	// it is persisted into. Absent applies the built-in default (10 MiB); <= 0
	// disables the cap (unbounded, not recommended).
	MaxOutputBytesPerExec *int
	// CpuLimit is the default whole-CPU cap per container, read the same way as
	// the memory limit: absent/null or <= 0 means no cap. It maps to the
	// container's NanoCPUs. Left off by default because a fair value depends on
	// how many workers share the host; an operator sets it after sizing.
	CpuLimit *int
	// IsolateProjectNetwork controls whether the contestant's project container
	// is denied external network access. When true (the default) the project
	// container is created already attached to an INTERNAL worker bridge (no
	// off-host route), and to nothing else, so contestant code cannot reach the
	// internet, the host gateway, or other services — while the test/grading
	// container keeps its normal bridge so it can still reach a Sonar host and
	// the like. It is attached at creation on purpose: Docker's "none" mode is
	// exclusive, so a container started with none can never be connected to
	// another network afterwards. A pointer so an absent key defaults to true
	// (secure) rather than to a bare bool's false; set it false only for a
	// deployment that genuinely needs run-time network inside the project
	// container.
	IsolateProjectNetwork *bool
	// ProvisionMaxAttempts is how many times a judgment's environment
	// provisioning is attempted before the judgment errors out (default 3).
	// A loaded host can make the project container miss its HEALTHCHECK
	// window on the first try; provisioning cleans up after itself, so
	// retrying is safe.
	ProvisionMaxAttempts int
}

func (c *JudgeConfig) applyDefaults() {
	if c.TestcaseTimeoutSeconds <= 0 {
		c.TestcaseTimeoutSeconds = 60
	}
	if c.LifecycleTimeoutSeconds <= 0 {
		c.LifecycleTimeoutSeconds = 120
	}
	if c.HealthcheckTimeoutSeconds <= 0 {
		c.HealthcheckTimeoutSeconds = 300
	}
	if c.MaxJudgeDurationMinutes <= 0 {
		c.MaxJudgeDurationMinutes = 30
	}
	if c.WorkerCount <= 0 {
		c.WorkerCount = 8
	}
	if c.ProvisionMaxAttempts <= 0 {
		c.ProvisionMaxAttempts = 3
	}
	// Only an ABSENT pids limit gets the fallback. An explicit 0 is the
	// operator saying "no cap" and must survive — that is the whole reason
	// this field is a pointer. The memory limit has no fallback at all: off is
	// its intended shipped state.
	if c.PidsLimit == nil {
		defaultPidsLimit := 1024
		c.PidsLimit = &defaultPidsLimit
	}
	// Absent output cap lands on the 10 MiB default. Like the pids limit this
	// ships on rather than off: an uncapped default would leave every upgraded
	// deployment exposed to the output-flood OOM this exists to prevent.
	if c.MaxOutputBytesPerExec == nil {
		defaultMaxOutputBytes := 10 * 1024 * 1024
		c.MaxOutputBytesPerExec = &defaultMaxOutputBytes
	}
	// Only an absent isolation flag gets the fallback. An explicit false is the
	// operator opting the project container back onto the network and must
	// survive; a missing key means "secure by default".
	if c.IsolateProjectNetwork == nil {
		isolate := true
		c.IsolateProjectNetwork = &isolate
	}
	// The repair intervals share the pointer three-state: only an absent key
	// falls back to 5 minutes; an explicit <= 0 is the off switch for the
	// periodic sweep and must survive.
	if c.OrphanSubmissionRepairInterval == nil {
		seconds := 300
		c.OrphanSubmissionRepairInterval = &seconds
	}
	if c.OrphanJudgmentRepairInterval == nil {
		seconds := 300
		c.OrphanJudgmentRepairInterval = &seconds
	}
	if c.RejudgeCommitRepairInterval == nil {
		seconds := 300
		c.RejudgeCommitRepairInterval = &seconds
	}
}

type DatabaseConfig struct {
	Mysql *MysqlConfig
	Mongo *MongoConfig
	Redis *RedisConfig
}

type DockerConfig struct {
	Mode     string
	Protocol string
	Host     string
	Port     int
	// TLSCACert, TLSCert and TLSKey are the mutual-TLS material for talking to a
	// REMOTE docker daemon. The docker API is root-equivalent, so a remote TCP
	// connection without client certificates is unacceptable; when Mode is
	// "remote" these must all be set (the provider fails closed otherwise)
	// unless AllowInsecureRemote is explicitly turned on.
	TLSCACert string
	TLSCert   string
	TLSKey    string
	// AllowInsecureRemote opts out of the mutual-TLS requirement for a REMOTE
	// daemon, permitting a plain-TCP connection with no client certificates.
	// It defaults to false (secure): only turn it on for a daemon reachable
	// solely over a trusted, isolated private network, and never where the
	// daemon's port could be reached from an untrusted host — anyone who can
	// reach it gets root on the judge host.
	AllowInsecureRemote bool
	// HostGatewayIP is the default-bridge (docker0) gateway address of the
	// docker host the judge spawns containers on. Judge-spawned containers
	// reach services published on that host (e.g. SonarQube) via this IP —
	// container IPs and cross-bridge traffic are not usable for that. It is
	// host-specific: fill it with scripts/detect-docker-gateway.sh, which
	// queries the same daemon the judge uses.
	HostGatewayIP string
}

// SonarConfig describes the SonarQube server that grading containers call
// for code-quality items. The full URL is composed by consumers as
// http://<Docker.HostGatewayIP>:<Port>.
type SonarConfig struct {
	Port  int    // published port on the judge docker host (default 9000)
	Token string // SonarQube user token (squ_...); keep out of git
}

func (c *SonarConfig) applyDefaults() {
	if c.Port <= 0 {
		c.Port = 9000
	}
}

type TestConfig struct {
	Server *ServerConfig
}

func rootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Join(filepath.Dir(d), "..")
}

// for test
var configRootDir = rootDir()
var configName = "config"

func NewConfig() *Config {
	viper.AddConfigPath(configRootDir)
	viper.SetConfigName(configName)

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Panicf("read config failed, err:%v\n", err)
	}

	var config Config

	if err := viper.Unmarshal(&config); err != nil {
		log.Panicf("unmarshal config failed, err:%v\n", err)
	}

	if config.Server.AppKey == "" {
		log.Panicf("app key is required")
	}
	// The encrypt service uses the first 16 bytes of this key as its AES-128
	// key, so a shorter key would panic at the first encrypt (a 500 in the
	// middle of a request) rather than fail loudly at startup. Catch it here.
	if len(config.Server.AppKey) < 16 {
		log.Panicf("app key must be at least 16 characters, got %d", len(config.Server.AppKey))
	}

	if config.Server.JwtKey == "" {
		log.Panicf("jwt key is required")
	}

	if config.Server.StoreKey == "" {
		log.Panicf("store key is required")
	}

	if config.Server.CsrfKey == "" {
		log.Panicf("csrf key is required")
	}

	config.Server.applyDefaults()

	if config.Auth == nil {
		config.Auth = &AuthConfig{}
	}
	config.Auth.applyDefaults()
	config.hardenCookies()

	if config.Judge == nil {
		config.Judge = &JudgeConfig{}
	}
	config.Judge.applyDefaults()

	if config.Sonar == nil {
		config.Sonar = &SonarConfig{}
	}
	config.Sonar.applyDefaults()

	// No fabrication here: a missing mysql section cannot boot anyway (no
	// host/credentials), so defaults only apply when the section exists.
	if config.Database != nil && config.Database.Mysql != nil {
		config.Database.Mysql.applyDefaults()
	}

	return &config
}
