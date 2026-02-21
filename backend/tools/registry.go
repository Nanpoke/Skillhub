package tools

import (
	"sync"
)

// Registry 工具适配器注册表
type Registry struct {
	adapters map[string]Adapter
	mu       sync.RWMutex
}

// NewRegistry 创建新的工具注册表
func NewRegistry() *Registry {
	r := &Registry{
		adapters: make(map[string]Adapter),
	}

	// 注册所有支持的工具
	r.Register(NewClaudeAdapter())
	r.Register(NewOpenCodeAdapter())
	r.Register(NewCursorAdapter())
	r.Register(NewCodeBuddyAdapter())
	r.Register(NewTraeAdapter())

	return r
}

// Register 注册工具适配器
func (r *Registry) Register(adapter Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.ID()] = adapter
}

// Get 获取指定工具的适配器
func (r *Registry) Get(id string) Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.adapters[id]
}

// GetAll 获取所有工具适配器
func (r *Registry) GetAll() []Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapters := make([]Adapter, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		adapters = append(adapters, adapter)
	}
	return adapters
}

// GetInstalled 获取所有已安装的工具
func (r *Registry) GetInstalled() []Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var installed []Adapter
	for _, adapter := range r.adapters {
		if adapter.IsInstalled() {
			installed = append(installed, adapter)
		}
	}
	return installed
}

// IDs 获取所有工具 ID
func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := make([]string, 0, len(r.adapters))
	for id := range r.adapters {
		ids = append(ids, id)
	}
	return ids
}
