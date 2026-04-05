package main

import "github.com/dtrugman/qory/lib/session"

// cachingProvider is a look-through cache over historyProvider.
// HistorySession results are cached on first fetch and invalidated on HistoryDelete.
type cachingProvider struct {
	inner historyProvider
	cache map[string]session.Session
}

func newCachingProvider(inner historyProvider) *cachingProvider {
	return &cachingProvider{
		inner: inner,
		cache: make(map[string]session.Session),
	}
}

func (c *cachingProvider) HistoryAll(limit int) ([]session.SessionPreview, error) {
	return c.inner.HistoryAll(limit)
}

func (c *cachingProvider) HistorySession(id string) (session.Session, error) {
	if sess, ok := c.cache[id]; ok {
		return sess, nil
	}
	sess, err := c.inner.HistorySession(id)
	if err != nil {
		return session.Session{}, err
	}
	c.cache[id] = sess
	return sess, nil
}

func (c *cachingProvider) HistoryDelete(id string) error {
	if err := c.inner.HistoryDelete(id); err != nil {
		return err
	}
	delete(c.cache, id)
	return nil
}
