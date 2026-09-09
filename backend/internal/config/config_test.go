package config

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
)

func GetExpectedConfig() *Config {
	return &Config{
		Server: &ServerConfig{
			Port:        8080,
			AppKey:      "test-app-key-secret",
			JwtKey:      "test-jwt-key",
			StoreKey:    "test-store-key",
			CsrfKey:     "test-csrf-key",
			RdbDriver:   "sqlite",
			StoragePath: "./storage",
			// The fixture omits these keys on purpose: config.yaml is
			// gitignored, so a deployment predating them must still come up
			// limited rather than unlimited.
			MaxUploadSizeMB: 100,
			MaxUnzipSizeMB:  1024,
			MaxUnzipEntries: 50000,
			// Absent from the fixture, so these are the default landings.
			BundleImportMaxUploadSizeMB: 2048,
			MetricsHost:                 "127.0.0.1",
			LoginRateLimitBurst:         intPtr(30),
			LoginRateLimitPerMinute:     intPtr(6),
			LoginRateLimitIpBurst:       intPtr(300),
			LoginRateLimitIpPerMinute:   intPtr(60),
		},
		Auth: &AuthConfig{
			// Present in the fixture, so these assert the values are read
			// rather than the defaults being applied.
			AccessTokenTTLSeconds:  3600,
			RefreshTokenTTLSeconds: 604800,
			Cookie: &AuthCookieConfig{
				Domain:   "",
				Path:     "/",
				Secure:   false,
				SameSite: "Lax",
			},
		},
		Database: &DatabaseConfig{
			Mysql: &MysqlConfig{
				Host:         "test-mysql-host",
				Port:         3306,
				Username:     "root",
				Password:     "1234",
				DBName:       "test-db",
				MaxOpenConns: 20,
				MaxIdleConns: 5,
				// The fixture yaml deliberately omits the key: this pins
				// "absent key means the 5-minute default".
				ConnMaxLifetimeMinutes: intPtr(5),
			},
			Redis: &RedisConfig{
				Host:     "redis",
				Port:     6379,
				Password: "",
				DB:       0,
			},
			Mongo: &MongoConfig{
				Host:     "mongo",
				Port:     27017,
				Username: "root",
				Password: "example",
			},
		},
		Judge: &JudgeConfig{
			TestcaseTimeoutSeconds:    60,
			LifecycleTimeoutSeconds:   120,
			HealthcheckTimeoutSeconds: 300,
			MaxJudgeDurationMinutes:   30,
			WorkerCount:               8,
			ProvisionMaxAttempts:      3,
			// Absent from the fixture, so these are the default landings.
			PidsLimit:                      intPtr(1024),
			MaxOutputBytesPerExec:          intPtr(10 * 1024 * 1024),
			IsolateProjectNetwork:          boolPtr(true),
			OrphanSubmissionRepairInterval: intPtr(300),
			OrphanJudgmentRepairInterval:   intPtr(300),
			RejudgeCommitRepairInterval:    intPtr(300),
		},
		Judgehost: &JudgehostConfig{
			Port:          8080,
			Image:         "clean-backend-judge",
			Network:       "go-app-network",
			ProjectVolume: "./judge/:/app/",
			DockerVolume:  "/var/run/docker.sock:/var/run/docker.sock",
			StorageVolume: "./storage/submissions/:/app/storage/submissions/",
			EnvFile:       ".env",
			Replicas:      3,
			RegisterURL:   "http://api:8080",
			Label:         "judgehost=true",
		},
		Docker: &DockerConfig{
			Mode:     "remote",
			Protocol: "tcp",
			Host:     "192.0.2.10",
			Port:     2376,
		},
		Sonar: &SonarConfig{
			Port: 9000,
		},
		Test: &TestConfig{
			Server: &ServerConfig{
				Port: 8081,
			},
		},
	}
}

// Token lifetimes were hardcoded to an hour and a week for a long time while
// these keys sat unread, so a config that predates them — config.yaml is
// gitignored and never receives new keys on deploy — must still come up with
// the lifetimes that were in force, not with zero-second tokens.
func TestAuthConfigApplyDefaults(t *testing.T) {
	t.Run("fills unset lifetimes", func(t *testing.T) {
		c := &AuthConfig{}
		c.applyDefaults()

		if c.AccessTokenTTLSeconds != 3600 {
			t.Errorf("expected access TTL 3600, but got %d", c.AccessTokenTTLSeconds)
		}
		if c.RefreshTokenTTLSeconds != 604800 {
			t.Errorf("expected refresh TTL 604800, but got %d", c.RefreshTokenTTLSeconds)
		}
	})

	t.Run("keeps configured lifetimes", func(t *testing.T) {
		c := &AuthConfig{AccessTokenTTLSeconds: 2, RefreshTokenTTLSeconds: 60}
		c.applyDefaults()

		if c.AccessTokenTTLSeconds != 2 {
			t.Errorf("expected access TTL 2, but got %d", c.AccessTokenTTLSeconds)
		}
		if c.RefreshTokenTTLSeconds != 60 {
			t.Errorf("expected refresh TTL 60, but got %d", c.RefreshTokenTTLSeconds)
		}
	})

	// The cookie block is optional in yaml, and every SetCookie call now
	// reads it — a nil block would panic the first login.
	t.Run("fills a missing cookie block", func(t *testing.T) {
		c := &AuthConfig{}
		c.applyDefaults()

		if c.Cookie == nil {
			t.Fatal("expected a cookie config, but got nil")
		}
		if c.Cookie.Path != "/" {
			t.Errorf("expected path /, but got %q", c.Cookie.Path)
		}
		if c.Cookie.SameSite != "Lax" {
			t.Errorf("expected SameSite Lax, but got %q", c.Cookie.SameSite)
		}
	})

	t.Run("fills only a missing cookie path", func(t *testing.T) {
		c := &AuthConfig{Cookie: &AuthCookieConfig{Domain: "example.com", SameSite: "None", Secure: true}}
		c.applyDefaults()

		if c.Cookie.Path != "/" {
			t.Errorf("expected path /, but got %q", c.Cookie.Path)
		}
		if c.Cookie.Domain != "example.com" || c.Cookie.SameSite != "None" || !c.Cookie.Secure {
			t.Errorf("expected configured cookie values to survive, but got %+v", c.Cookie)
		}
	})
}

func intPtr(v int) *int { return &v }

func boolPtr(v bool) *bool { return &v }

// The login rate limit is a *int for the same three-state reason as the
// container caps: an explicit 0 (the off switch) must not collapse into "unset"
// and be handed the default limit the operator was trying to remove.
func TestServerConfigApplyDefaultsLoginRateLimit(t *testing.T) {
	t.Run("absent applies the defaults", func(t *testing.T) {
		c := &ServerConfig{}
		c.applyDefaults()
		if c.LoginRateLimitBurst == nil || *c.LoginRateLimitBurst != 30 {
			t.Errorf("expected default burst 30, got %v", c.LoginRateLimitBurst)
		}
		if c.LoginRateLimitPerMinute == nil || *c.LoginRateLimitPerMinute != 6 {
			t.Errorf("expected default per-minute 6, got %v", c.LoginRateLimitPerMinute)
		}
		if c.LoginRateLimitIpBurst == nil || *c.LoginRateLimitIpBurst != 300 {
			t.Errorf("expected default IP burst 300, got %v", c.LoginRateLimitIpBurst)
		}
		if c.LoginRateLimitIpPerMinute == nil || *c.LoginRateLimitIpPerMinute != 60 {
			t.Errorf("expected default IP per-minute 60, got %v", c.LoginRateLimitIpPerMinute)
		}
	})

	t.Run("an explicit zero survives as the off switch", func(t *testing.T) {
		c := &ServerConfig{LoginRateLimitBurst: intPtr(0), LoginRateLimitPerMinute: intPtr(0)}
		c.applyDefaults()
		if *c.LoginRateLimitBurst != 0 || *c.LoginRateLimitPerMinute != 0 {
			t.Errorf("explicit 0 must survive, got burst=%d perMinute=%d", *c.LoginRateLimitBurst, *c.LoginRateLimitPerMinute)
		}
	})

	t.Run("a configured value survives", func(t *testing.T) {
		c := &ServerConfig{LoginRateLimitBurst: intPtr(5), LoginRateLimitPerMinute: intPtr(10)}
		c.applyDefaults()
		if *c.LoginRateLimitBurst != 5 || *c.LoginRateLimitPerMinute != 10 {
			t.Errorf("configured values must survive, got burst=%d perMinute=%d", *c.LoginRateLimitBurst, *c.LoginRateLimitPerMinute)
		}
	})
}

// Same three-state convention as the container caps: absent means the safe
// default, an explicit non-positive value is the off switch, a positive value
// survives untouched.
func TestMysqlConfigApplyDefaultsConnMaxLifetime(t *testing.T) {
	t.Run("fills an absent lifetime with 5 minutes", func(t *testing.T) {
		c := &MysqlConfig{}
		c.applyDefaults()
		if c.ConnMaxLifetimeMinutes == nil || *c.ConnMaxLifetimeMinutes != 5 {
			t.Errorf("absent key must default to 5, got %v", c.ConnMaxLifetimeMinutes)
		}
	})

	t.Run("an explicit zero survives as the off switch", func(t *testing.T) {
		c := &MysqlConfig{ConnMaxLifetimeMinutes: intPtr(0)}
		c.applyDefaults()
		if *c.ConnMaxLifetimeMinutes != 0 {
			t.Errorf("explicit 0 must survive, got %d", *c.ConnMaxLifetimeMinutes)
		}
	})

	t.Run("a configured value survives", func(t *testing.T) {
		c := &MysqlConfig{ConnMaxLifetimeMinutes: intPtr(30)}
		c.applyDefaults()
		if *c.ConnMaxLifetimeMinutes != 30 {
			t.Errorf("configured value must survive, got %d", *c.ConnMaxLifetimeMinutes)
		}
	})
}

// The repair intervals follow the same three-state convention; the explicit
// zero used to silently fall back to the default, so "0 disables the sweep"
// only became true when the fields moved to pointers — pin it.
func TestJudgeConfigApplyDefaultsRepairIntervals(t *testing.T) {
	t.Run("fills absent intervals with 300 seconds", func(t *testing.T) {
		c := &JudgeConfig{}
		c.applyDefaults()
		if c.OrphanSubmissionRepairInterval == nil || *c.OrphanSubmissionRepairInterval != 300 {
			t.Errorf("absent submission interval must default to 300, got %v", c.OrphanSubmissionRepairInterval)
		}
		if c.OrphanJudgmentRepairInterval == nil || *c.OrphanJudgmentRepairInterval != 300 {
			t.Errorf("absent judgment interval must default to 300, got %v", c.OrphanJudgmentRepairInterval)
		}
		if c.RejudgeCommitRepairInterval == nil || *c.RejudgeCommitRepairInterval != 300 {
			t.Errorf("absent rejudge-commit interval must default to 300, got %v", c.RejudgeCommitRepairInterval)
		}
	})

	t.Run("an explicit zero survives as the off switch", func(t *testing.T) {
		c := &JudgeConfig{
			OrphanSubmissionRepairInterval: intPtr(0),
			OrphanJudgmentRepairInterval:   intPtr(0),
			RejudgeCommitRepairInterval:    intPtr(0),
		}
		c.applyDefaults()
		if *c.OrphanSubmissionRepairInterval != 0 || *c.OrphanJudgmentRepairInterval != 0 || *c.RejudgeCommitRepairInterval != 0 {
			t.Errorf("explicit 0 must survive, got %d / %d / %d", *c.OrphanSubmissionRepairInterval, *c.OrphanJudgmentRepairInterval, *c.RejudgeCommitRepairInterval)
		}
	})

	t.Run("a configured value survives", func(t *testing.T) {
		c := &JudgeConfig{
			OrphanSubmissionRepairInterval: intPtr(60),
			OrphanJudgmentRepairInterval:   intPtr(90),
			RejudgeCommitRepairInterval:    intPtr(120),
		}
		c.applyDefaults()
		if *c.OrphanSubmissionRepairInterval != 60 || *c.OrphanJudgmentRepairInterval != 90 || *c.RejudgeCommitRepairInterval != 120 {
			t.Errorf("configured values must survive, got %d / %d / %d", *c.OrphanSubmissionRepairInterval, *c.OrphanJudgmentRepairInterval, *c.RejudgeCommitRepairInterval)
		}
	})
}

// The container caps have three states and only two of them are visible in a
// yaml file, so every one of these is a way to run without the limit the
// deployment believes it has.
func TestJudgeConfigApplyDefaultsContainerLimits(t *testing.T) {
	// config.yaml is gitignored, so an existing deployment's file predates
	// this key and will never contain it. Absent has to mean 1024, or the
	// upgrade that added the cap leaves everyone uncapped.
	t.Run("fills an absent pids limit", func(t *testing.T) {
		c := &JudgeConfig{}
		c.applyDefaults()

		if c.PidsLimit == nil {
			t.Fatal("expected pids limit 1024, but got nil")
		}
		if *c.PidsLimit != 1024 {
			t.Errorf("expected pids limit 1024, but got %d", *c.PidsLimit)
		}
	})

	// The reason this field is a pointer. An explicit 0 is the operator
	// switching the cap off, and it must not be mistaken for an absent key —
	// which is what happened before, silently handing them 1024 instead.
	t.Run("keeps an explicit zero pids limit as no cap", func(t *testing.T) {
		c := &JudgeConfig{PidsLimit: intPtr(0)}
		c.applyDefaults()

		if c.PidsLimit == nil {
			t.Fatal("expected the explicit 0 to survive, but got nil")
		}
		if *c.PidsLimit != 0 {
			t.Errorf("expected the explicit 0 to survive, but got %d", *c.PidsLimit)
		}
	})

	t.Run("keeps a configured pids limit", func(t *testing.T) {
		c := &JudgeConfig{PidsLimit: intPtr(256)}
		c.applyDefaults()

		if c.PidsLimit == nil {
			t.Fatal("expected pids limit 256, but got nil")
		}
		if *c.PidsLimit != 256 {
			t.Errorf("expected pids limit 256, but got %d", *c.PidsLimit)
		}
	})

	// Still accepted: an earlier revision of the config documentation told
	// operators that a negative value was how to switch the cap off.
	t.Run("keeps a negative pids limit", func(t *testing.T) {
		c := &JudgeConfig{PidsLimit: intPtr(-1)}
		c.applyDefaults()

		if c.PidsLimit == nil {
			t.Fatal("expected the opt-out to survive, but got nil")
		}
		if *c.PidsLimit != -1 {
			t.Errorf("expected the opt-out to survive, but got %d", *c.PidsLimit)
		}
	})

	// The opposite call, for the opposite reason: a memory cap has real false
	// positives (a sonar-scanner JVM legitimately takes a gigabyte), so
	// turning it on is an operator's decision. A default here would fail
	// heavy-but-honest submissions on upgrade.
	t.Run("leaves the memory limit off", func(t *testing.T) {
		c := &JudgeConfig{}
		c.applyDefaults()

		if c.MemoryLimitMB != nil {
			t.Errorf("expected the memory limit to stay unset, but got %v", *c.MemoryLimitMB)
		}
	})

	t.Run("keeps a configured memory limit", func(t *testing.T) {
		c := &JudgeConfig{MemoryLimitMB: intPtr(2048)}
		c.applyDefaults()

		if c.MemoryLimitMB == nil {
			t.Fatal("expected memory limit 2048, but got nil")
		}
		if *c.MemoryLimitMB != 2048 {
			t.Errorf("expected memory limit 2048, but got %d", *c.MemoryLimitMB)
		}
	})
}

// SameSite arrives from yaml as a string and has to become a gin/net-http
// mode. An unrecognised value must not silently become None, which would
// widen the cookie's exposure rather than narrow it.
func TestAuthCookieConfigSameSiteMode(t *testing.T) {
	cases := map[string]http.SameSite{
		"Lax":      http.SameSiteLaxMode,
		"lax":      http.SameSiteLaxMode,
		"None":     http.SameSiteNoneMode,
		"none":     http.SameSiteNoneMode,
		"Strict":   http.SameSiteStrictMode,
		"strict":   http.SameSiteStrictMode,
		"":         http.SameSiteLaxMode,
		"nonsense": http.SameSiteLaxMode,
	}

	for value, want := range cases {
		t.Run(value, func(t *testing.T) {
			c := &AuthCookieConfig{SameSite: value}
			if got := c.SameSiteMode(); got != want {
				t.Errorf("expected %v for %q, but got %v", want, value, got)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	configRootDir = "./config/" // updated directory
	configName = "config_test"
	config := NewConfig()

	expected := GetExpectedConfig()
	if diff := cmp.Diff(expected, config); diff != "" {
		t.Errorf("Config mismatch (-expected +got):\n%s", diff)
	}
}

func TestNewConfigWithEmptyAppKey(t *testing.T) {
	configRootDir = "./config/" // updated directory
	configName = "config_test_empty_app_key"
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewConfig() did not panic")
		}
	}()

	NewConfig()
}

func TestNewConfigWithEmptyJwtKey(t *testing.T) {
	configRootDir = "./config/" // updated directory
	configName = "config_test_empty_jwt_key"
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewConfig() did not panic")
		}
	}()

	NewConfig()
}

func TestNewConfigWithEmptyStoreKey(t *testing.T) {
	configRootDir = "./config/" // updated directory
	configName = "config_test_empty_store_key"
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewConfig() did not panic")
		}
	}()

	NewConfig()
}

func TestNewConfigWithEmptyCsrfKey(t *testing.T) {
	configRootDir = "./config/" // updated directory
	configName = "config_test_empty_csrf_key"
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("NewConfig() did not panic")
		}
	}()

	NewConfig()
}

// The pointer fields only work because viper tells an absent key apart from an
// explicit 0. Every test above builds JudgeConfig directly and so assumes that
// rather than showing it; this one goes through the yaml, which is the only
// form an operator ever writes, and would catch the assumption breaking under
// a viper upgrade.
func TestContainerLimitsSurviveUnmarshalFromYAML(t *testing.T) {
	cases := []struct {
		label string
		yaml  string
		want  int
	}{
		// The case the whole pointer change exists for: written as 0, stays 0.
		{"an explicit zero", "judge:\n  pidsLimit: 0\n", 0},
		{"an absent key", "judge:\n  workerCount: 8\n", 1024},
		{"a configured value", "judge:\n  pidsLimit: 256\n", 256},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			v := viper.New()
			v.SetConfigType("yaml")
			if err := v.ReadConfig(strings.NewReader(tc.yaml)); err != nil {
				t.Fatalf("read config: %v", err)
			}

			var c Config
			if err := v.Unmarshal(&c); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			c.Judge.applyDefaults()

			if c.Judge.PidsLimit == nil {
				t.Fatalf("expected pids limit %d, but got nil", tc.want)
			}
			if *c.Judge.PidsLimit != tc.want {
				t.Errorf("expected pids limit %d, but got %d", tc.want, *c.Judge.PidsLimit)
			}
		})
	}
}

// config.example.yaml is the file every new deployment copies, and nothing
// else checks it. A typo in a key, or a field added to Config whose example
// spelling does not match, breaks the very first step of the quick start with
// a panic — and it breaks it only for new users, who are the least equipped to
// tell whether they made the mistake or the project did.
func TestExampleConfigIsUsable(t *testing.T) {
	f, err := os.Open("../../config.example.yaml")
	if err != nil {
		t.Fatalf("open config.example.yaml: %v", err)
	}
	defer f.Close()

	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(f); err != nil {
		t.Fatalf("config.example.yaml is not valid yaml: %v", err)
	}

	var c Config
	if err := v.Unmarshal(&c); err != nil {
		t.Fatalf("config.example.yaml does not unmarshal into Config: %v", err)
	}
	// Checked BEFORE applyDefaults, deliberately. Afterwards a misspelled key
	// is indistinguishable from a correct one — the fallback fills in the same
	// 1024 either way — so asserting on the resolved value would pass on a
	// typo, which is the single most likely way this file breaks.
	if c.Judge.PidsLimit == nil {
		t.Fatal("config.example.yaml does not set judge.pidsLimit — check the key spelling")
	}
	if *c.Judge.PidsLimit != 1024 {
		t.Errorf("expected the example to ship pidsLimit 1024, but got %d", *c.Judge.PidsLimit)
	}
	if c.Judge.MemoryLimitMB == nil {
		t.Fatal("config.example.yaml does not set judge.memoryLimitMB — check the key spelling")
	}
	// The shipped posture, asserted rather than described: process cap on,
	// memory cap off. Flipping either changes what every fresh install runs
	// under and should not pass unnoticed.
	if *c.Judge.MemoryLimitMB != 0 {
		t.Errorf("expected the example to ship the memory cap off, but got %d", *c.Judge.MemoryLimitMB)
	}
}
