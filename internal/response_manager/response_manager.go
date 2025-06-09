package response_manager

import (
	"context"
	"fmt"
	"github.com/MaiM-with-u/maibot-telegram-adapter/internal/logger"
	"sync"
	"time"
)

// Config 全局配置接口
type Config interface {
	GetHeartbeatInterval() time.Duration
}

// Response 响应结构体
type Response struct {
	Data      map[string]interface{}
	Timestamp time.Time
}

// ResponseManager 响应管理器
type ResponseManager struct {
	responses     sync.Map // map[string]*Response
	config        Config
	cleanupTicker *time.Ticker
	done          chan struct{}
	wg            sync.WaitGroup
	responseChans sync.Map // map[string]chan map[string]interface{}
}

// NewResponseManager 创建新的响应管理器
func NewResponseManager(config Config) *ResponseManager {
	rm := &ResponseManager{
		config: config,
		done:   make(chan struct{}),
	}

	// 启动超时清理协程
	rm.startTimeoutCleaner()

	return rm
}

// GetResponse 获取响应，基于 channel 的等待机制
func (rm *ResponseManager) GetResponse(ctx context.Context, requestID string) (map[string]interface{}, error) {
	ch := make(chan map[string]interface{}, 1)
	rm.responseChans.Store(requestID, ch)
	defer rm.responseChans.Delete(requestID)

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("请求超时或取消，request_id: %s", requestID)
	case resp := <-ch:
		return resp, nil
	}
}

// GetResponseAsync 异步获取响应，返回channel
func (rm *ResponseManager) GetResponseAsync(ctx context.Context, requestID string) <-chan Result {
	resultChan := make(chan Result, 1)

	rm.wg.Add(1)
	go func() {
		defer rm.wg.Done()
		defer close(resultChan)

		data, err := rm.GetResponse(ctx, requestID)
		resultChan <- Result{Data: data, Error: err}
	}()

	return resultChan
}

// Result 异步结果
type Result struct {
	Data  map[string]interface{}
	Error error
}

// PutResponse 存储响应，并通知等待的 channel
func (rm *ResponseManager) PutResponse(response map[string]interface{}) error {
	echoID, ok := response["echo"].(string)
	if !ok {
		return fmt.Errorf("响应中缺少echo字段或类型错误")
	}

	resp := &Response{
		Data:      response,
		Timestamp: time.Now(),
	}

	rm.responses.Store(echoID, resp)
	logger.Trace("响应信息id: %s 已存入响应字典", echoID)

	if ch, ok := rm.responseChans.LoadAndDelete(echoID); ok {
		ch.(chan map[string]interface{}) <- response
	}

	return nil
}

// startTimeoutCleaner 启动超时清理协程
func (rm *ResponseManager) startTimeoutCleaner() {
	interval := rm.config.GetHeartbeatInterval()
	rm.cleanupTicker = time.NewTicker(interval)

	rm.wg.Add(1)
	go func() {
		defer rm.wg.Done()

		for {
			select {
			case <-rm.done:
				return
			case <-rm.cleanupTicker.C:
				rm.cleanTimeoutResponses()
			}
		}
	}()
}

// cleanTimeoutResponses 清理超时响应
func (rm *ResponseManager) cleanTimeoutResponses() {
	cleanedCount := 0
	now := time.Now()
	heartbeatInterval := rm.config.GetHeartbeatInterval()

	rm.responses.Range(func(key, value interface{}) bool {
		echoID := key.(string)
		response := value.(*Response)

		if now.Sub(response.Timestamp) > heartbeatInterval {
			rm.responses.Delete(echoID)
			cleanedCount++
			logger.Warning("响应消息 %s 超时，已删除", echoID)
		}

		return true
	})

	logger.Info("已删除 %d 条超时响应消息", cleanedCount)
}

// Close 关闭响应管理器
func (rm *ResponseManager) Close() {
	if rm.cleanupTicker != nil {
		rm.cleanupTicker.Stop()
	}

	close(rm.done)
	rm.wg.Wait()
}

// Stats 获取统计信息
func (rm *ResponseManager) Stats() (count int) {
	rm.responses.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// 示例配置实现
type SimpleConfig struct {
	HeartbeatInterval time.Duration
}

func (c *SimpleConfig) GetHeartbeatInterval() time.Duration {
	return c.HeartbeatInterval
}

// 使用示例
func Example() {
	// 创建配置和日志
	config := &SimpleConfig{
		HeartbeatInterval: 30 * time.Second,
	}

	// 创建响应管理器
	rm := NewResponseManager(config)
	defer rm.Close()

	// 存储响应
	response := map[string]interface{}{
		"echo": "request123",
		"data": "some data",
	}
	_ = rm.PutResponse(response)

	// 获取响应
	ctx := context.Background()
	data, err := rm.GetResponse(ctx, "request123")
	if err != nil {
		logger.Warning("Error: %v\n", err)
		return
	}

	logger.Info("Received response: %v\n", data)
}
