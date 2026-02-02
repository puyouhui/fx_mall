package api

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"go_backend/internal/model"
)

// prepayCacheEntry 预支付缓存条目（带过期时间）
type prepayCacheEntry struct {
	Data      *model.CachedPrepayEntry
	CreatedAt time.Time
}

var (
	prepayCache     = make(map[string]*prepayCacheEntry)
	prepayCacheLock sync.RWMutex
	prepayCacheTTL  = 30 * time.Minute
)

func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			prepayCacheLock.Lock()
			now := time.Now()
			for k, v := range prepayCache {
				if v != nil && now.Sub(v.CreatedAt) > prepayCacheTTL {
					delete(prepayCache, k)
				}
			}
			prepayCacheLock.Unlock()
		}
	}()
}

// GenerateOutTradeNo 生成预支付商户订单号（保证唯一性，支持高并发）
func GenerateOutTradeNo() string {
	now := time.Now()
	ts := now.Format("20060102150405")
	ns := now.Nanosecond() % 10000 // 纳秒后4位，降低同秒内碰撞
	r := rand.Intn(1000000)
	return fmt.Sprintf("P%s%04d%06d", ts, ns, r)
}

// SetPrepayCache 写入预支付缓存
func SetPrepayCache(outTradeNo string, data *model.CachedPrepayEntry) {
	prepayCacheLock.Lock()
	defer prepayCacheLock.Unlock()
	prepayCache[outTradeNo] = &prepayCacheEntry{Data: data, CreatedAt: time.Now()}
}

// GetPrepayCache 读取并删除预支付缓存（支付成功后一次性消费）
func GetPrepayCache(outTradeNo string) (*model.CachedPrepayEntry, error) {
	prepayCacheLock.Lock()
	defer prepayCacheLock.Unlock()
	entry, ok := prepayCache[outTradeNo]
	if !ok || entry == nil {
		return nil, fmt.Errorf("预支付缓存不存在或已过期")
	}
	if time.Since(entry.CreatedAt) > prepayCacheTTL {
		delete(prepayCache, outTradeNo)
		return nil, fmt.Errorf("预支付缓存已过期")
	}
	delete(prepayCache, outTradeNo)
	return entry.Data, nil
}

func logPrepayCache(prefix string, outTradeNo string, err error) {
	if err != nil {
		log.Printf("[PrepayCache] %s out_trade_no=%s err=%v", prefix, outTradeNo, err)
	}
}
