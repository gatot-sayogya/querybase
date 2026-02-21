package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	CacheKeyStatsGlobal = "stats:dashboard:global"
	CacheKeyStatsUser   = "stats:dashboard:user:%s"
	CacheTTLExpiration  = 5 * time.Minute
)

// DashboardStats represents the aggregated statistics for the dashboard
type DashboardStats struct {
	MyQueriesToday   int `json:"my_queries_today"`
	PendingApprovals int `json:"pending_approvals,omitempty"`
	TotalQueries     int `json:"total_queries,omitempty"`
	DBAccessCount    int `json:"db_access_count"`
	TotalUsers       int `json:"total_users,omitempty"`
}

// StatsService handles dashboard statistics logic
type StatsService struct {
	db                   *gorm.DB
	redis                *redis.Client
	statsChangedCallback func()
}

// NewStatsService creates a new StatsService
func NewStatsService(db *gorm.DB, redisClient *redis.Client) *StatsService {
	return &StatsService{
		db:    db,
		redis: redisClient,
	}
}

// SetStatsChangedCallback sets a callback triggered when stats change
func (s *StatsService) SetStatsChangedCallback(callback func()) {
	s.statsChangedCallback = callback
}

// TriggerStatsChanged invalidates the cache and notifies listeners (e.g., WebSockets)
func (s *StatsService) TriggerStatsChanged() {
	// Flush the global and all user-scoped caches so the next fetch is fresh
	if s.redis != nil {
		ctx := context.Background()
		// Delete global cache
		s.redis.Del(ctx, CacheKeyStatsGlobal)
		// Use SCAN to delete all user-scoped keys
		iter := s.redis.Scan(ctx, 0, "stats:dashboard:user:*", 100).Iterator()
		for iter.Next(ctx) {
			s.redis.Del(ctx, iter.Val())
		}
	}

	if s.statsChangedCallback != nil {
		s.statsChangedCallback()
	}
}

// GetDashboardStats returns statistics for the specified user
func (s *StatsService) GetDashboardStats(ctx context.Context, userID string, isAdmin bool) (*DashboardStats, error) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf(CacheKeyStatsUser, userID)
	if isAdmin {
		cacheKey = CacheKeyStatsGlobal
	}

	cachedData, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && cachedData != "" {
		var stats DashboardStats
		if err := json.Unmarshal([]byte(cachedData), &stats); err == nil {
			return &stats, nil
		}
	}

	// Calculate and set cache if not found or expired
	stats, err := s.CalculateStats(userID, isAdmin)
	if err != nil {
		return nil, err
	}

	// Save to cache asynchronously or synchronously (we'll do sync for simplicity, async for speed usually)
	if statsBytes, err := json.Marshal(stats); err == nil {
		s.redis.Set(ctx, cacheKey, statsBytes, CacheTTLExpiration)
	}

	return stats, nil
}

// CalculateStats dynamically calculates the stats from the database
func (s *StatsService) CalculateStats(userID string, isAdmin bool) (*DashboardStats, error) {
	stats := &DashboardStats{}

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	// 1. My queries today
	var myQueries int64
	if err := s.db.Table("query_history").
		Where("user_id = ? AND executed_at >= ?", userID, startOfDay).
		Count(&myQueries).Error; err != nil {
		log.Printf("Error counting my queries: %v", err)
	}
	stats.MyQueriesToday = int(myQueries)

	// 2. DB Access Count (Active Data sources count)
	var dbCount int64
	if err := s.db.Table("data_sources").Count(&dbCount).Error; err != nil {
		log.Printf("Error counting databases: %v", err)
	}
	stats.DBAccessCount = int(dbCount)

	// Admin specific stats
	if isAdmin {
		// 3. Pending Approvals
		var pendingApprovals int64
		if err := s.db.Table("approval_requests").
			Where("status = ?", "pending").
			Count(&pendingApprovals).Error; err != nil {
			log.Printf("Error counting pending approvals: %v", err)
		}
		stats.PendingApprovals = int(pendingApprovals)

		// 4. Total Queries (All Users, all time)
		var totalQueries int64
		if err := s.db.Table("query_history").Count(&totalQueries).Error; err != nil {
			log.Printf("Error counting total queries: %v", err)
		}
		stats.TotalQueries = int(totalQueries)

		// 5. Total Users
		var totalUsers int64
		if err := s.db.Table("users").Count(&totalUsers).Error; err != nil {
			log.Printf("Error counting total users: %v", err)
		}
		stats.TotalUsers = int(totalUsers)
	}

	return stats, nil
}

// InvalidateUserCache invalidates the dashboard stats cache for a specific user
func (s *StatsService) InvalidateUserCache(ctx context.Context, userID string) error {
	return s.redis.Del(ctx, fmt.Sprintf(CacheKeyStatsUser, userID)).Err()
}

// InvalidateGlobalCache invalidates the global (admin) dashboard stats cache
func (s *StatsService) InvalidateGlobalCache(ctx context.Context) error {
	return s.redis.Del(ctx, CacheKeyStatsGlobal).Err()
}
