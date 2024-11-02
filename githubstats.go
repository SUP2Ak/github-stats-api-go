package githubstats

import (
	"fmt"
	"sync"
	"time"

	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Types

type IncludeOptions struct {
	IncludeStars       bool // Include stars
	IncludeFollowers   bool // Include followers
	IncludeFollowing   bool // Include following
	IncludeRepos       bool // Include repositories
	IncludeFirstNRepos int  // Number of repositories to retrieve
	IncludeOrgs        bool // Include organizations
}

type Config struct {
	Path           string         // API path
	Token          string         // GitHub token
	IP             string         // IP address
	Port           string         // Port
	Scheme         string         // HTTP or HTTPS
	CertFile       string         // Certificate file
	KeyFile        string         // Key file
	IncludeOptions IncludeOptions // Include options
	CacheDuration  time.Duration  // Cache duration
	RateLimit      int            // Rate limit
}

type CacheEntry struct {
	Stats      GitHubStats
	Expiration time.Time
}

type Organizations struct {
	Organizations []string `json:"organizations"`
}

type GitHubStats struct {
	Username      string      `json:"username"`
	Followers     int         `json:"followers"`
	Following     int         `json:"following"`
	TotalStars    int         `json:"total_stars"`
	Repositories  []RepoStats `json:"repositories"`
	Organizations []string    `json:"organizations"`
}

type RepoStats struct {
	Name         string         `json:"name"`
	Stars        int            `json:"stars"`
	Forks        int            `json:"forks"`
	OpenIssues   int            `json:"open_issues"`
	Contributors map[string]int `json:"contributors"`
}

type GStats struct {
	client      *github.Client
	cache       *Cache
	rateLimiter *RateLimiter
}

type Cache struct {
	mu    sync.RWMutex
	store map[string]CacheEntry
}

type RateLimiter struct {
	mu          sync.Mutex
	requests    int
	lastRequest time.Time
	limit       int
	interval    time.Duration
}

// RateLimiter Fonctions

// NewRateLimiter Create a new rate limiter.
/*
 * @param limit int - The limit
 * @param interval time.Duration - The interval
 * @return *RateLimiter - The rate limiter
 */
func NewRateLimiter(limit int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:    limit,
		interval: interval,
	}
}

// Allow Allow a request.
/*
 * @return bool - The result
 */
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if time.Since(rl.lastRequest) > rl.interval {
		rl.requests = 0
	}

	if rl.requests < rl.limit {
		rl.requests++
		rl.lastRequest = time.Now()
		return true
	}
	return false
}

// Cache Fonctions

// NewCache Create a new cache.
/*
 * @return *Cache - The cache
 */
func NewCache() *Cache {
	return &Cache{
		store: make(map[string]CacheEntry),
	}
}

// Get Get the cache entry.
/*
 * @param key string - The key
 * @return GitHubStats, bool - The stats, found
 */
func (c *Cache) Get(key string) (GitHubStats, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, found := c.store[key]
	if !found || time.Now().After(entry.Expiration) {
		return GitHubStats{}, false
	}
	return entry.Stats, true
}

// Set Set the cache entry.
/*
 * @param key string - The key
 * @param stats GitHubStats - The stats
 * @param duration time.Duration - The duration
 * @return void
 */
func (c *Cache) Set(key string, stats GitHubStats, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store[key] = CacheEntry{
		Stats:      stats,
		Expiration: time.Now().Add(duration),
	}
}

// parseIncludeOptions Parse the include options.
/*
 * @param query url.Values - The query
 * @return IncludeOptions - The options
 */
func (g *GStats) parseIncludeOptions(query url.Values) IncludeOptions {
	opts := IncludeOptions{
		IncludeStars:       query.Get("include_stars") == "true",
		IncludeFollowers:   query.Get("include_followers") == "true",
		IncludeFollowing:   query.Get("include_following") == "true",
		IncludeRepos:       query.Get("include_repos") == "true",
		IncludeOrgs:        query.Get("include_orgs") == "true",
		IncludeFirstNRepos: 5, // Valeur par défaut
	}

	if firstN := query.Get("include_first_n_repos"); firstN != "" {
		if n, err := strconv.Atoi(firstN); err == nil {
			opts.IncludeFirstNRepos = n
		}
	}

	return opts
}

// githubStatsHandler Handle the requests to get the GitHub stats.
/*
 * @param w http.ResponseWriter - The response writer
 * @param r *http.Request - The request
 * @param config Config - The configuration
 * @return void
 */
func (g *GStats) githubStatsHandler(w http.ResponseWriter, r *http.Request, config Config) {
	query := r.URL.Query()
	username := query.Get("username")

	if username == "" {
		http.Error(w, "Le nom d'utilisateur est requis", http.StatusBadRequest)
		return
	}

	// Check the request limit
	if !g.rateLimiter.Allow() {
		http.Error(w, "Request limit exceeded", http.StatusTooManyRequests)
		return
	}

	// Check the cache
	if cachedStats, found := g.cache.Get(username); found {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(cachedStats)
		return
	}

	// Get the include options
	opts := g.parseIncludeOptions(query)

	stats, err := g.GetGitHubStats(username, opts)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des données", http.StatusInternalServerError)
		return
	}

	// Cache the stats
	g.cache.Set(username, stats, config.CacheDuration)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// GetGitHubStats Get the GitHub stats for a given user according to the specified options.
/*
 * @param username string - The username
 * @param opts IncludeOptions - The options
 * @return GitHubStats, error - The stats, the error
 */
func (g *GStats) GetGitHubStats(username string, opts IncludeOptions) (GitHubStats, error) {
	ctx := context.Background()

	user, _, err := g.client.Users.Get(ctx, username)
	if err != nil {
		return GitHubStats{}, err
	}

	repos, _, err := g.client.Repositories.List(ctx, username, nil)
	if err != nil {
		return GitHubStats{}, err
	}

	stats := GitHubStats{
		Username: username,
	}

	if opts.IncludeFollowers {
		stats.Followers = *user.Followers
	}
	if opts.IncludeFollowing {
		stats.Following = *user.Following
	}
	if opts.IncludeStars || opts.IncludeRepos {
		for i, repo := range repos {
			if opts.IncludeStars {
				stats.TotalStars += *repo.StargazersCount
			}
			if opts.IncludeRepos {
				if opts.IncludeFirstNRepos > 0 && i >= opts.IncludeFirstNRepos {
					break
				}
				stats.Repositories = append(stats.Repositories, RepoStats{
					Name:  *repo.Name,
					Stars: *repo.StargazersCount,
					Forks: *repo.ForksCount,
				})
			}
		}
	}

	if opts.IncludeOrgs {
		orgs, _, err := g.client.Organizations.List(ctx, username, nil)
		if err != nil {
			return GitHubStats{}, err
		}

		for _, org := range orgs {
			stats.Organizations = append(stats.Organizations, *org.Login)
		}
	}

	return stats, nil
}

// Connect initialise le client GitHub avec le token et configure le serveur.
/*
 * @param config Config - The configuration
 * @return error? - The error
 */
func (g *GStats) Connect(config Config) error {
	// Check if the token is defined
	if config.Token == "" {
		return fmt.Errorf("le token GitHub doit être défini")
	}
	if config.IP == "" {
		config.IP = "0.0.0.0" // Default value
	}
	if config.Port == "" {
		config.Port = "8080" // Default value
	}
	if config.Scheme == "" {
		config.Scheme = "http" // Default value
	}
	if config.Path == "" {
		config.Path = "/stats" // Default value
	}
	if config.RateLimit == 0 {
		config.RateLimit = 10 // Default value
	}
	if config.CacheDuration == 0 {
		config.CacheDuration = 1 * time.Hour // Default value
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	g.client = github.NewClient(tc)

	g.cache = NewCache()
	g.rateLimiter = NewRateLimiter(config.RateLimit, 1*time.Minute) // 10 requests per minute

	// Start the HTTP server
	http.HandleFunc(config.Path, func(w http.ResponseWriter, r *http.Request) {
		g.githubStatsHandler(w, r, config)
	})

	if config.Scheme == "https" {
		// Use ListenAndServeTLS for HTTPS
		return http.ListenAndServeTLS(config.IP+":"+config.Port, config.CertFile, config.KeyFile, nil)
	}

	// Use ListenAndServe for HTTP
	return http.ListenAndServe(config.IP+":"+config.Port, nil)
}
