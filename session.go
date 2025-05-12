package gooo

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"sync"
	"time"
)

const (
	SessionCookieName = "go_session"
	DefaultExpire     = 24 * time.Hour
	SameSiteMode      = http.SameSiteLaxMode // 防御CSRF攻击
	SessionIDLength   = 64                   // 增加SessionID长度
)

// 添加 错误响应
var ErrSessionNotFound = errors.New("session not found")

// 错误回调机制
var OnSessionError = func(c *Context, err error) {
	c.String(500, "Session Error: %v", err)
}

// 添加会话绑定中间件
func SessionMiddleware() HandlerFunc {
	return func(c *Context) {
		c.StartSession()
		defer func() {
			if mgr := c.engine.GetSessionManager(); mgr != nil {
				mgr.Store.Save(c.SessionID, c.Session)
			}
		}()
		c.Next()
	}
}

// SessionStore 是 session 存储后端的接口
type SessionStore interface {
	Get(id string) (map[string]any, error)
	Set(id string, data map[string]any) error
	Save(id string, data map[string]any) error
	Destroy(id string) error
	GC() error
}

// MemorySession 内存存储的 Session 数据结构
type MemorySession struct {
	Data      map[string]any
	ExpiresAt time.Time // 移除CreatedAt
}

type MemoryStore struct {
	store map[string]*MemorySession
	mu    sync.RWMutex
}

type SessionManager struct {
	Store      SessionStore
	CookieOpts CookieConfig // 新增配置结构
}

type CookieConfig struct {
	Secure   bool
	SameSite http.SameSite
	MaxAge   int
}

// NewMemoryStore 创建一个内存存储的 SessionStore
func NewMemoryStore(gcInterval time.Duration) *MemoryStore {
	store := &MemoryStore{
		store: make(map[string]*MemorySession),
	}

	// 启动后台 GC
	go func() {
		for range time.Tick(gcInterval) {
			store.GC()
		}
	}()

	return store
}

// GC 清理过期的 session
func (s *MemoryStore) GC() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, session := range s.store {
		if now.After(session.ExpiresAt) {
			delete(s.store, id)
		}
	}
	return nil
}

// Set 设置 session 数据
func (s *MemoryStore) Set(id string, data map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[id] = &MemorySession{
		Data:      data,
		ExpiresAt: time.Now().Add(DefaultExpire),
	}
	return nil
}

func (m *SessionManager) Create(w http.ResponseWriter, r *http.Request) (map[string]any, string) {
	cookie, err := r.Cookie(SessionCookieName)
	var sessionID string

	if err != nil || cookie.Value == "" {
		// 生成新 session ID
		sessionID, err = generateSessionID()
		if err != nil {
			OnSessionError(&Context{Writer: w, Req: r}, err)
			return nil, ""
		}
		// 使用 CookieOpts 配置
		http.SetCookie(w, &http.Cookie{
			Name:     SessionCookieName,
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
			Secure:   m.CookieOpts.Secure,
			SameSite: m.CookieOpts.SameSite,
			MaxAge:   m.CookieOpts.MaxAge,
		})
		m.Store.Set(sessionID, make(map[string]any))
	} else {
		sessionID = cookie.Value
	}

	data, err := m.Store.Get(sessionID)
	if err != nil {
		data = make(map[string]any)
		m.Store.Set(sessionID, data)
	}

	return data, sessionID
}

// Get 获取 session 数据
func (s *MemoryStore) Get(id string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if session, ok := s.store[id]; ok {
		if time.Now().Before(session.ExpiresAt) {
			// 自动续期
			session.ExpiresAt = time.Now().Add(DefaultExpire)
			return session.Data, nil
		}
		// 自动清理过期会话
		go s.Destroy(id)
	}
	return nil, ErrSessionNotFound
}

func (s *MemoryStore) Save(id string, data map[string]any) error {
	return s.Set(id, data)
}

func (s *MemoryStore) Destroy(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.store, id)
	return nil
}

func generateSessionID() (string, error) {
	b := make([]byte, SessionIDLength)
	if _, err := rand.Read(b); err != nil {
		return "", err // 不要panic
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func NewSessionManager(store SessionStore) *SessionManager {
	return &SessionManager{
		Store: store,
	}
}

func (m *SessionManager) DestroySession(w http.ResponseWriter, sessionID string) {
	m.Store.Destroy(sessionID)
	// 立即过期客户端cookie
	http.SetCookie(w, &http.Cookie{
		Name:   SessionCookieName,
		Value:  "",
		MaxAge: -1,
	})
}

func (m *SessionManager) Renew(sessionID string) {
	if data, err := m.Store.Get(sessionID); err == nil {
		m.Store.Set(sessionID, data) // 利用Set的自动续期
	}
}

// 新增会话重置方法
func (m *SessionManager) RegenerateID(w http.ResponseWriter, oldID string) string {
	newID, err := generateSessionID()
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		return oldID // 返回旧ID避免nil
	}

	// 获取旧会话数据
	data, err := m.Store.Get(oldID)
	if err != nil {
		return oldID // 如果旧会话不存在，直接返回旧ID
	}

	// 设置新会话并销毁旧会话
	if err := m.Store.Set(newID, data); err != nil {
		return oldID
	}
	if err := m.Store.Destroy(oldID); err != nil {
		return oldID
	}

	// 设置新 Cookie
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    newID,
		Path:     "/",
		HttpOnly: true,
		Secure:   m.CookieOpts.Secure,
		SameSite: m.CookieOpts.SameSite,
		MaxAge:   m.CookieOpts.MaxAge,
	})

	return newID
}
