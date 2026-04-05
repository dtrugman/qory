package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dtrugman/qory/lib/message"
	"github.com/dtrugman/qory/lib/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- test helpers ----

type mockProvider struct {
	previews   []session.SessionPreview
	msgs       []message.Message
	historyErr error
	sessErr    error
}

func (mp mockProvider) HistoryAll() ([]session.SessionPreview, error) {
	return mp.previews, mp.historyErr
}

func (mp mockProvider) HistorySession(id string) (session.Session, error) {
	return session.Session{Messages: mp.msgs}, mp.sessErr
}

func makePreview(name string) session.SessionPreview {
	return session.SessionPreview{
		Name:      name,
		UpdatedAt: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		Snippet:   "some snippet",
	}
}

func makePreviews(names ...string) []session.SessionPreview {
	p := make([]session.SessionPreview, len(names))
	for i, n := range names {
		p[i] = makePreview(n)
	}
	return p
}

// mustModel constructs a sessionModel from the given provider, failing the
// test immediately if construction returns an error.
func mustModel(t *testing.T, p historyProvider) sessionModel {
	t.Helper()
	m, err := newSessionModel(p)
	require.NoError(t, err)
	return m
}

func sendKey(m sessionModel, key string) sessionModel {
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return next.(sessionModel)
}

func sendSpecialKey(m sessionModel, keyType tea.KeyType) sessionModel {
	next, _ := m.Update(tea.KeyMsg{Type: keyType})
	return next.(sessionModel)
}

// ---- newSessionModel ----

func TestNewSessionModel_EmptyPreviews(t *testing.T) {
	m := mustModel(t, mockProvider{})
	assert.Equal(t, 0, m.list.cursor)
	assert.Empty(t, m.selected)
	assert.True(t, m.list.showPreview)
	assert.Equal(t, 0, m.list.previewOffset)
}

func TestNewSessionModel_PropagatesHistoryAllError(t *testing.T) {
	sentinel := fmt.Errorf("storage unavailable")
	_, err := newSessionModel(mockProvider{historyErr: sentinel})
	assert.ErrorIs(t, err, sentinel)
}

func TestNewSessionModel_PreloadsFirstSession(t *testing.T) {
	msgs := []message.Message{message.NewUserMessage("hello")}
	m := mustModel(t, mockProvider{previews: makePreviews("session-a", "session-b"), msgs: msgs})
	cached, ok := m.cache["session-a"]
	require.True(t, ok, "first session should be pre-loaded into cache")
	assert.Equal(t, msgs, cached.messages)
	assert.NoError(t, cached.err)
}

// ---- session navigation ----

func TestUpdate_NavigateDown(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b", "c")})
	m = sendKey(m, "j")
	assert.Equal(t, 1, m.list.cursor)
}

func TestUpdate_NavigateDownArrow(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b")})
	m = sendSpecialKey(m, tea.KeyDown)
	assert.Equal(t, 1, m.list.cursor)
}

func TestUpdate_NavigateUp(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b", "c")})
	m = sendKey(m, "j")
	m = sendKey(m, "k")
	assert.Equal(t, 0, m.list.cursor)
}

func TestUpdate_NavigateUpArrow(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b")})
	m = sendSpecialKey(m, tea.KeyDown)
	m = sendSpecialKey(m, tea.KeyUp)
	assert.Equal(t, 0, m.list.cursor)
}

func TestUpdate_CursorDoesNotGoAboveZero(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b")})
	m = sendKey(m, "k")
	assert.Equal(t, 0, m.list.cursor)
}

func TestUpdate_CursorDoesNotGoBeyondEnd(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b")})
	m = sendKey(m, "j")
	m = sendKey(m, "j")
	assert.Equal(t, 1, m.list.cursor)
}

func TestUpdate_NavigationCachesSession(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b")})
	assert.NotContains(t, m.cache, "b")
	m = sendKey(m, "j")
	assert.Contains(t, m.cache, "b")
}

func TestUpdate_NavigationResetsPreviewOffset(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b")})
	m = sendSpecialKey(m, tea.KeyCtrlJ)
	require.Equal(t, 1, m.list.previewOffset)
	m = sendKey(m, "j")
	assert.Equal(t, 0, m.list.previewOffset)
}

// ---- selection ----

func TestUpdate_EnterSelectsCurrentSession(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b")})
	m = sendKey(m, "j")
	m = sendSpecialKey(m, tea.KeyEnter)
	assert.Equal(t, "b", m.selected)
}

func TestUpdate_EnterOnEmptyDoesNothing(t *testing.T) {
	m := mustModel(t, mockProvider{})
	m = sendSpecialKey(m, tea.KeyEnter)
	assert.Empty(t, m.selected)
}

func TestUpdate_EnterCachesSession(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b")})
	m = sendKey(m, "j")
	// Simulate the session not yet being in the cache.
	delete(m.cache, "b")
	require.NotContains(t, m.cache, "b")
	m = sendSpecialKey(m, tea.KeyEnter)
	assert.Contains(t, m.cache, "b")
}

// ---- preview scroll ----

func TestUpdate_CtrlJScrollsPreviewDown(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendSpecialKey(m, tea.KeyCtrlJ)
	assert.Equal(t, 1, m.list.previewOffset)
}

func TestUpdate_CtrlDownScrollsPreviewDown(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendSpecialKey(m, tea.KeyCtrlDown)
	assert.Equal(t, 1, m.list.previewOffset)
}

func TestUpdate_CtrlKScrollsPreviewUp(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendSpecialKey(m, tea.KeyCtrlJ)
	m = sendSpecialKey(m, tea.KeyCtrlK)
	assert.Equal(t, 0, m.list.previewOffset)
}

func TestUpdate_CtrlUpScrollsPreviewUp(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendSpecialKey(m, tea.KeyCtrlDown)
	m = sendSpecialKey(m, tea.KeyCtrlUp)
	assert.Equal(t, 0, m.list.previewOffset)
}

func TestUpdate_PreviewOffsetDoesNotGoBelowZero(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendSpecialKey(m, tea.KeyCtrlK)
	assert.Equal(t, 0, m.list.previewOffset)
}

// ---- preview toggle ----

func TestUpdate_PTogglesPreviewOff(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendKey(m, "p")
	assert.False(t, m.list.showPreview)
}

func TestUpdate_PTogglesPreviewBackOn(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendKey(m, "p")
	m = sendKey(m, "p")
	assert.True(t, m.list.showPreview)
}

func TestUpdate_PreviewToggleResetsOffset(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendSpecialKey(m, tea.KeyCtrlJ)
	require.Equal(t, 1, m.list.previewOffset)
	m = sendKey(m, "p")
	assert.Equal(t, 0, m.list.previewOffset)
}

// ---- quit ----

func TestUpdate_QKeyQuits(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendKey(m, "q")
	assert.True(t, m.quitting)
	assert.Empty(t, m.selected)
	assert.Empty(t, m.View()) // quitting clears the screen on bubbletea's final render
}

func TestUpdate_EscQuits(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendSpecialKey(m, tea.KeyEsc)
	assert.True(t, m.quitting)
	assert.Empty(t, m.View())
}

func TestUpdate_CtrlCQuits(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendSpecialKey(m, tea.KeyCtrlC)
	assert.True(t, m.quitting)
	assert.Empty(t, m.View())
}

// ---- window resize ----

func TestUpdate_WindowSizeUpdated(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	result := next.(sessionModel)
	assert.Equal(t, 120, result.width)
	assert.Equal(t, 40, result.height)
}

// ---- navCount ----

func TestNavCount_PreviewVisible_AlwaysThree(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b", "c", "d", "e")})
	m.height = 40
	assert.Equal(t, 3, m.navCount())
}

func TestNavCount_PreviewHidden_FillsHeight(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b", "c", "d", "e")})
	m.list.showPreview = false
	m.height = 10 // 10 - 3 overhead = 7, but only 5 sessions exist
	assert.Equal(t, 5, m.navCount())
}

func TestNavCount_PreviewHidden_CappedBySessionCount(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b")})
	m.list.showPreview = false
	m.height = 40
	assert.Equal(t, 2, m.navCount())
}

func TestNavCount_PreviewHidden_MinimumThree(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b", "c", "d")})
	m.list.showPreview = false
	m.height = 4 // 4 - 3 = 1, below minimum; should clamp to 3
	assert.Equal(t, 3, m.navCount())
}

// ---- visibleItems ----

func TestVisibleItems_Empty(t *testing.T) {
	m := mustModel(t, mockProvider{})
	items, idx := m.visibleItems(3)
	assert.Empty(t, items)
	assert.Equal(t, 0, idx)
}

func TestVisibleItems_SingleSession(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("only")})
	items, idx := m.visibleItems(3)
	assert.Len(t, items, 1)
	assert.Equal(t, 0, idx)
}

func TestVisibleItems_AtFirst_ShowsUpToCount(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b", "c", "d")})
	items, idx := m.visibleItems(3)
	assert.Equal(t, makePreviews("a", "b", "c"), items)
	assert.Equal(t, 0, idx)
}

func TestVisibleItems_InMiddle_CursorCentred(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b", "c", "d")})
	m = sendKey(m, "j") // cursor=1
	m = sendKey(m, "j") // cursor=2
	items, idx := m.visibleItems(3)
	assert.Equal(t, makePreviews("b", "c", "d"), items)
	assert.Equal(t, 1, idx)
}

func TestVisibleItems_AtLast_WindowAnchors(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b", "c", "d")})
	m = sendKey(m, "j")
	m = sendKey(m, "j")
	m = sendKey(m, "j") // cursor=3 (last)
	items, idx := m.visibleItems(3)
	assert.Equal(t, makePreviews("b", "c", "d"), items)
	assert.Equal(t, 2, idx)
}

func TestVisibleItems_TwoSessions_AtFirst(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b")})
	items, idx := m.visibleItems(3)
	assert.Len(t, items, 2)
	assert.Equal(t, 0, idx)
}

func TestVisibleItems_TwoSessions_AtLast(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b")})
	m = sendKey(m, "j")
	items, idx := m.visibleItems(3)
	assert.Len(t, items, 2)
	assert.Equal(t, 1, idx)
}

func TestVisibleItems_LargeCount_ShowsAll(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b", "c", "d", "e")})
	m = sendKey(m, "j")
	m = sendKey(m, "j") // cursor=2
	items, idx := m.visibleItems(5)
	assert.Equal(t, makePreviews("a", "b", "c", "d", "e"), items)
	assert.Equal(t, 2, idx)
}

// ---- buildPreviewLines ----

func TestBuildPreviewLines_RendersRoles(t *testing.T) {
	msgs := []message.Message{
		message.NewUserMessage("hello"),
		message.NewAssistantMessage("world"),
	}
	m := mustModel(t, mockProvider{previews: makePreviews("s"), msgs: msgs})
	m.cacheSession("s")

	lines := m.buildPreviewLines(m.cache["s"])
	joined := strings.Join(lines, "\n")
	assert.Contains(t, joined, "USER")
	assert.Contains(t, joined, "ASSISTANT")
	assert.Contains(t, joined, "hello")
	assert.Contains(t, joined, "world")
}

func TestBuildPreviewLines_WrapsLongContent(t *testing.T) {
	longMsg := strings.Repeat("x", 200)
	msgs := []message.Message{message.NewUserMessage(longMsg)}
	m := mustModel(t, mockProvider{previews: makePreviews("s"), msgs: msgs})
	m.width = 80
	m.cacheSession("s")

	lines := m.buildPreviewLines(m.cache["s"])
	for _, l := range lines {
		assert.LessOrEqual(t, len(l), m.width*3, "line unexpectedly long: %q", l)
	}
}

// ---- renderPreview ----

func TestRenderPreview_ClampsOffsetBeyondContent(t *testing.T) {
	msgs := []message.Message{message.NewUserMessage("short")}
	m := mustModel(t, mockProvider{previews: makePreviews("s"), msgs: msgs})
	m.cacheSession("s")
	m.list.previewOffset = 9999

	out := m.renderPreview(10)
	assert.Contains(t, out, "short")
}

func TestRenderPreview_ShowsMoreIndicator(t *testing.T) {
	var msgs []message.Message
	for i := range 10 {
		msgs = append(msgs, message.NewUserMessage(fmt.Sprintf("message %d", i)))
	}
	m := mustModel(t, mockProvider{previews: makePreviews("s"), msgs: msgs})
	m.cacheSession("s")

	out := m.renderPreview(3)
	assert.Contains(t, out, "↓")
	assert.Contains(t, out, "more lines")
}

func TestRenderPreview_NoMoreIndicatorWhenAllFits(t *testing.T) {
	msgs := []message.Message{message.NewUserMessage("one line")}
	m := mustModel(t, mockProvider{previews: makePreviews("s"), msgs: msgs})
	m.cacheSession("s")

	out := m.renderPreview(50)
	assert.NotContains(t, out, "more lines")
}

func TestRenderPreview_ScrollRevealsNewContent(t *testing.T) {
	var msgs []message.Message
	for i := range 20 {
		msgs = append(msgs, message.NewUserMessage(fmt.Sprintf("msg-%d", i)))
	}
	m := mustModel(t, mockProvider{previews: makePreviews("s"), msgs: msgs})
	m.cacheSession("s")

	before := m.renderPreview(5)
	m.list.previewOffset = 10
	after := m.renderPreview(5)

	assert.NotEqual(t, before, after)
}

// ---- View ----

func TestView_HeaderShowsSessionCount(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b", "c")})
	assert.Contains(t, m.View(), "Session history (3 sessions)")
}

func TestView_NavigatorShowsAbsoluteNumbers(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a", "b", "c", "d", "e")})
	m = sendKey(m, "j")
	m = sendKey(m, "j") // cursor on "c" (index 2), window shows [b, c, d]
	view := m.View()
	assert.Contains(t, view, "2.")
	assert.Contains(t, view, "3.")
	assert.Contains(t, view, "4.")
}

func TestView_EmptyPreviewsMessage(t *testing.T) {
	m := mustModel(t, mockProvider{})
	assert.Contains(t, m.View(), "No sessions found")
}

func TestView_SessionNameAppearsInNavigator(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("my-session")})
	assert.Contains(t, m.View(), "my-session")
}

func TestView_DateAppearsInNavigator(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("s")})
	assert.Contains(t, m.View(), "Jan 15 2024 10:00")
}

func TestView_PreviewShowsMessages(t *testing.T) {
	msgs := []message.Message{
		message.NewUserMessage("what is the capital of France"),
		message.NewAssistantMessage("Paris"),
	}
	m := mustModel(t, mockProvider{previews: makePreviews("s1"), msgs: msgs})
	view := m.View()
	assert.Contains(t, view, "USER")
	assert.Contains(t, view, "ASSISTANT")
	assert.Contains(t, view, "what is the capital of France")
	assert.Contains(t, view, "Paris")
}

func TestView_PreviewHiddenWhenToggled(t *testing.T) {
	msgs := []message.Message{message.NewUserMessage("hello")}
	m := mustModel(t, mockProvider{previews: makePreviews("s1"), msgs: msgs})
	m = sendKey(m, "p")
	assert.NotContains(t, m.View(), "hello")
}

func TestView_PreviewShowsLoadError(t *testing.T) {
	sentinel := fmt.Errorf("disk error")
	m := mustModel(t, mockProvider{previews: makePreviews("bad"), sessErr: sentinel})
	assert.Contains(t, m.View(), "Error loading session")
}

func TestView_HelpShowsScrollHintWhenPreviewVisible(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	assert.Contains(t, m.View(), "ctrl+↑/↓ scroll preview")
}

func TestView_HelpHidesScrollHintWhenPreviewHidden(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendKey(m, "p")
	assert.NotContains(t, m.View(), "ctrl+↑/↓ scroll preview")
}

func TestView_HelpShowsHidePreviewWhenVisible(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	assert.Contains(t, m.View(), "p hide preview")
}

func TestView_HelpShowsShowPreviewWhenHidden(t *testing.T) {
	m := mustModel(t, mockProvider{previews: makePreviews("a")})
	m = sendKey(m, "p")
	assert.Contains(t, m.View(), "p show preview")
}

// ---- wordWrap ----

func TestWordWrap_ShortLineUnchanged(t *testing.T) {
	assert.Equal(t, "hello", wordWrap("hello", 20))
}

func TestWordWrap_LongLineBreaks(t *testing.T) {
	assert.Equal(t, "abcd\nefgh\nij", wordWrap("abcdefghij", 4))
}

func TestWordWrap_PreservesExistingNewlines(t *testing.T) {
	assert.Equal(t, "abc\ndef", wordWrap("abc\ndef", 10))
}

func TestWordWrap_ZeroWidthReturnsUnchanged(t *testing.T) {
	assert.Equal(t, "hello", wordWrap("hello", 0))
}

func TestWordWrap_EmptyString(t *testing.T) {
	assert.Equal(t, "", wordWrap("", 10))
}
