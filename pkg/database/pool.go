package database

import (
	"database/sql"
	"fmt"
	"sync/atomic"
	"time"
)

type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type PoolMetrics struct {
	OpenConnections int64
	InUse           int64
	WaitCount       int64
	WaitDuration    time.Duration
}

type OptimizedPool struct {
	db      *sql.DB
	config  PoolConfig
	metrics atomic.Value
}

func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 5 * time.Minute,
	}
}

func NewOptimizedPool(db *sql.DB, cfg PoolConfig) *OptimizedPool {
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	pool := &OptimizedPool{
		db:     db,
		config: cfg,
	}
	pool.updateMetrics()
	return pool
}

func (p *OptimizedPool) DB() *sql.DB {
	return p.db
}

func (p *OptimizedPool) updateMetrics() {
	stats := p.db.Stats()
	p.metrics.Store(PoolMetrics{
		OpenConnections: int64(stats.OpenConnections),
		InUse:           int64(stats.InUse),
		WaitCount:       int64(stats.WaitCount),
		WaitDuration:    stats.WaitDuration,
	})
}

func (p *OptimizedPool) GetMetrics() PoolMetrics {
	p.updateMetrics()
	val := p.metrics.Load()
	if val == nil {
		return PoolMetrics{}
	}
	return val.(PoolMetrics)
}

func (p *OptimizedPool) PromMetricsText() string {
	m := p.GetMetrics()
	return fmt.Sprintf(
		"# HELP db_pool_open Number of open connections\n"+
			"# TYPE db_pool_open gauge\n"+
			"db_pool_open %d\n"+
			"# HELP db_pool_in_use Number of connections in use\n"+
			"# TYPE db_pool_in_use gauge\n"+
			"db_pool_in_use %d\n"+
			"# HELP db_pool_wait_count Total number of connections waited for\n"+
			"# TYPE db_pool_wait_count counter\n"+
			"db_pool_wait_count %d\n",
		m.OpenConnections, m.InUse, m.WaitCount,
	)
}
