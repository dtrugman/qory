package main

import (
	"fmt"
	"testing"

	"github.com/dtrugman/qory/lib/message"
	"github.com/dtrugman/qory/lib/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCachingProvider_HistoryAllDelegatesToInner(t *testing.T) {
	previews := makePreviews("a", "b")
	inner := &MockHistoryProvider{}
	inner.On("HistoryAll").Return(previews, nil)

	p := newCachingProvider(inner)
	result, err := p.HistoryAll()

	require.NoError(t, err)
	assert.Equal(t, previews, result)
	inner.AssertExpectations(t)
}

func TestCachingProvider_HistorySessionFetchesFromInnerOnFirstCall(t *testing.T) {
	msgs := []message.Message{message.NewUserMessage("hello")}
	inner := &MockHistoryProvider{}
	inner.On("HistorySession", "a").Return(session.Session{Messages: msgs}, nil)

	p := newCachingProvider(inner)
	sess, err := p.HistorySession("a")

	require.NoError(t, err)
	assert.Equal(t, msgs, sess.Messages)
	inner.AssertExpectations(t)
}

func TestCachingProvider_HistorySessionReturnsCachedResultOnSubsequentCalls(t *testing.T) {
	msgs := []message.Message{message.NewUserMessage("hello")}
	inner := &MockHistoryProvider{}
	inner.On("HistorySession", "a").Return(session.Session{Messages: msgs}, nil).Once()

	p := newCachingProvider(inner)
	_, _ = p.HistorySession("a")
	sess, err := p.HistorySession("a")

	require.NoError(t, err)
	assert.Equal(t, msgs, sess.Messages)
	inner.AssertExpectations(t) // called exactly once
}

func TestCachingProvider_HistorySessionDoesNotCacheErrors(t *testing.T) {
	fetchErr := fmt.Errorf("disk error")
	inner := &MockHistoryProvider{}
	inner.On("HistorySession", "a").Return(session.Session{}, fetchErr).Once()
	inner.On("HistorySession", "a").Return(session.Session{Messages: []message.Message{message.NewUserMessage("retry ok")}}, nil).Once()

	p := newCachingProvider(inner)
	_, err := p.HistorySession("a")
	require.ErrorIs(t, err, fetchErr)

	sess, err := p.HistorySession("a")
	require.NoError(t, err)
	assert.Len(t, sess.Messages, 1)
	inner.AssertExpectations(t)
}

func TestCachingProvider_HistoryDeleteInvalidatesCache(t *testing.T) {
	msgs := []message.Message{message.NewUserMessage("hello")}
	inner := &MockHistoryProvider{}
	inner.On("HistorySession", "a").Return(session.Session{Messages: msgs}, nil).Once()
	inner.On("HistoryDelete", "a").Return(nil)
	inner.On("HistorySession", "a").Return(session.Session{Messages: msgs}, nil).Once()

	p := newCachingProvider(inner)
	_, _ = p.HistorySession("a") // warms cache
	err := p.HistoryDelete("a")
	require.NoError(t, err)
	_, _ = p.HistorySession("a") // must hit inner again

	inner.AssertExpectations(t)
}

func TestCachingProvider_HistoryDeleteErrorPreservesCache(t *testing.T) {
	msgs := []message.Message{message.NewUserMessage("hello")}
	deleteErr := fmt.Errorf("permission denied")
	inner := &MockHistoryProvider{}
	inner.On("HistorySession", "a").Return(session.Session{Messages: msgs}, nil).Once()
	inner.On("HistoryDelete", "a").Return(deleteErr)

	p := newCachingProvider(inner)
	_, _ = p.HistorySession("a") // warms cache
	err := p.HistoryDelete("a")
	require.ErrorIs(t, err, deleteErr)

	// cache still valid: inner should NOT be called again
	sess, err := p.HistorySession("a")
	require.NoError(t, err)
	assert.Equal(t, msgs, sess.Messages)
	inner.AssertExpectations(t)
}
