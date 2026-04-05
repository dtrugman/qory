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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---- mock ----

type MockHistoryProvider struct {
	mock.Mock
}

func (m *MockHistoryProvider) HistoryAll(_ ...int) ([]session.SessionPreview, error) {
	args := m.Called()
	return args.Get(0).([]session.SessionPreview), args.Error(1)
}

func (m *MockHistoryProvider) HistorySession(id string) (session.Session, error) {
	args := m.Called(id)
	return args.Get(0).(session.Session), args.Error(1)
}

func (m *MockHistoryProvider) HistoryDelete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// newMockProvider returns a mock pre-configured with a HistoryAll expectation.
func newMockProvider(previews []session.SessionPreview) *MockHistoryProvider {
	p := &MockHistoryProvider{}
	p.On("HistoryAll").Return(previews, nil)
	return p
}

// ---- test helpers ----

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
	p := newMockProvider(makePreviews())
	m := mustModel(t, p)
	assert.Equal(t, 0, m.list.cursor)
	assert.Empty(t, m.selected)
	assert.True(t, m.list.showPreview)
	assert.Equal(t, 0, m.list.previewOffset)
	p.AssertExpectations(t)
}

func TestNewSessionModel_PropagatesHistoryAllError(t *testing.T) {
	sentinel := fmt.Errorf("storage unavailable")
	p := &MockHistoryProvider{}
	p.On("HistoryAll").Return([]session.SessionPreview(nil), sentinel)

	_, err := newSessionModel(p)

	assert.ErrorIs(t, err, sentinel)
	p.AssertExpectations(t)
}

// ---- session navigation ----

func TestUpdate_NavigateDown(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c"))
	m := mustModel(t, p)
	m = sendKey(m, "j")
	assert.Equal(t, 1, m.list.cursor)
	p.AssertExpectations(t)
}

func TestUpdate_NavigateDownArrow(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b"))
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyDown)
	assert.Equal(t, 1, m.list.cursor)
	p.AssertExpectations(t)
}

func TestUpdate_NavigateUp(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c"))
	m := mustModel(t, p)
	m = sendKey(m, "j")
	m = sendKey(m, "k")
	assert.Equal(t, 0, m.list.cursor)
	p.AssertExpectations(t)
}

func TestUpdate_NavigateUpArrow(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b"))
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyDown)
	m = sendSpecialKey(m, tea.KeyUp)
	assert.Equal(t, 0, m.list.cursor)
	p.AssertExpectations(t)
}

func TestUpdate_CursorDoesNotGoAboveZero(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b"))
	m := mustModel(t, p)
	m = sendKey(m, "k")
	assert.Equal(t, 0, m.list.cursor)
	p.AssertExpectations(t)
}

func TestUpdate_CursorDoesNotGoBeyondEnd(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b"))
	m := mustModel(t, p)
	m = sendKey(m, "j")
	m = sendKey(m, "j")
	assert.Equal(t, 1, m.list.cursor)
	p.AssertExpectations(t)
}

func TestUpdate_NavigationResetsPreviewOffset(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b"))
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyCtrlJ)
	require.Equal(t, 1, m.list.previewOffset)
	m = sendKey(m, "j")
	assert.Equal(t, 0, m.list.previewOffset)
	p.AssertExpectations(t)
}

// ---- selection ----

func TestUpdate_EnterSelectsCurrentSession(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b"))
	m := mustModel(t, p)
	m = sendKey(m, "j")
	m = sendSpecialKey(m, tea.KeyEnter)
	assert.Equal(t, "b", m.selected)
	p.AssertExpectations(t)
}

func TestUpdate_EnterOnEmptyDoesNothing(t *testing.T) {
	p := newMockProvider(makePreviews())
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyEnter)
	assert.Empty(t, m.selected)
	p.AssertExpectations(t)
}

// ---- delete ----

func TestUpdate_DeleteRemovesSessionFromList(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c"))
	p.On("HistoryDelete", "b").Return(nil)
	m := mustModel(t, p)
	m = sendKey(m, "j") // cursor on "b"
	m = sendKey(m, "d")
	assert.Equal(t, makePreviews("a", "c"), m.previews)
	p.AssertExpectations(t)
}

func TestUpdate_DeleteKeepsCursorWhenNotAtEnd(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c"))
	p.On("HistoryDelete", "b").Return(nil)
	m := mustModel(t, p)
	m = sendKey(m, "j") // cursor=1 ("b")
	m = sendKey(m, "d")
	assert.Equal(t, 1, m.list.cursor) // cursor stays, now points at "c"
	p.AssertExpectations(t)
}

func TestUpdate_DeleteDecrementsCorsorWhenAtEnd(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c"))
	p.On("HistoryDelete", "c").Return(nil)
	m := mustModel(t, p)
	m = sendKey(m, "j")
	m = sendKey(m, "j") // cursor=2 ("c", the last item)
	m = sendKey(m, "d")
	assert.Equal(t, 1, m.list.cursor)
	p.AssertExpectations(t)
}

func TestUpdate_DeleteOnlyItem(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	p.On("HistoryDelete", "a").Return(nil)
	m := mustModel(t, p)
	m = sendKey(m, "d")
	assert.Empty(t, m.previews)
	assert.Equal(t, 0, m.list.cursor)
	p.AssertExpectations(t)
}

func TestUpdate_DeleteResetsPreviewOffset(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b"))
	p.On("HistoryDelete", "a").Return(nil)
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyCtrlJ)
	require.Equal(t, 1, m.list.previewOffset)
	m = sendKey(m, "d")
	assert.Equal(t, 0, m.list.previewOffset)
	p.AssertExpectations(t)
}

func TestUpdate_DeleteErrorLeavesListUnchanged(t *testing.T) {
	deleteErr := fmt.Errorf("permission denied")
	p := newMockProvider(makePreviews("a", "b"))
	p.On("HistoryDelete", "a").Return(deleteErr)
	m := mustModel(t, p)
	m = sendKey(m, "d")
	assert.Len(t, m.previews, 2)
	assert.Contains(t, m.statusMsg, "permission denied")
	p.AssertExpectations(t)
}

func TestUpdate_DeleteOnEmptyListDoesNothing(t *testing.T) {
	p := newMockProvider(makePreviews())
	m := mustModel(t, p)
	m = sendKey(m, "d")
	assert.Empty(t, m.previews)
	assert.Equal(t, 0, m.list.cursor)
	p.AssertExpectations(t)
}

// ---- preview scroll ----

func TestUpdate_CtrlJScrollsPreviewDown(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyCtrlJ)
	assert.Equal(t, 1, m.list.previewOffset)
	p.AssertExpectations(t)
}

func TestUpdate_CtrlDownScrollsPreviewDown(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyCtrlDown)
	assert.Equal(t, 1, m.list.previewOffset)
	p.AssertExpectations(t)
}

func TestUpdate_CtrlKScrollsPreviewUp(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyCtrlJ)
	m = sendSpecialKey(m, tea.KeyCtrlK)
	assert.Equal(t, 0, m.list.previewOffset)
	p.AssertExpectations(t)
}

func TestUpdate_CtrlUpScrollsPreviewUp(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyCtrlDown)
	m = sendSpecialKey(m, tea.KeyCtrlUp)
	assert.Equal(t, 0, m.list.previewOffset)
	p.AssertExpectations(t)
}

func TestUpdate_PreviewOffsetDoesNotGoBelowZero(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyCtrlK)
	assert.Equal(t, 0, m.list.previewOffset)
	p.AssertExpectations(t)
}

// ---- preview toggle ----

func TestUpdate_PTogglesPreviewOff(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendKey(m, "p")
	assert.False(t, m.list.showPreview)
	p.AssertExpectations(t)
}

func TestUpdate_PTogglesPreviewBackOn(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendKey(m, "p")
	m = sendKey(m, "p")
	assert.True(t, m.list.showPreview)
	p.AssertExpectations(t)
}

func TestUpdate_PreviewToggleResetsOffset(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyCtrlJ)
	require.Equal(t, 1, m.list.previewOffset)
	m = sendKey(m, "p")
	assert.Equal(t, 0, m.list.previewOffset)
	p.AssertExpectations(t)
}

// ---- quit ----

func TestUpdate_QKeyQuits(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendKey(m, "q")
	assert.True(t, m.quitting)
	assert.Empty(t, m.selected)
	assert.Empty(t, m.View()) // quitting clears the screen on bubbletea's final render
	p.AssertExpectations(t)
}

func TestUpdate_EscQuits(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyEsc)
	assert.True(t, m.quitting)
	assert.Empty(t, m.View())
	p.AssertExpectations(t)
}

func TestUpdate_CtrlCQuits(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendSpecialKey(m, tea.KeyCtrlC)
	assert.True(t, m.quitting)
	assert.Empty(t, m.View())
	p.AssertExpectations(t)
}

// ---- window resize ----

func TestUpdate_WindowSizeUpdated(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	next, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	result := next.(sessionModel)
	assert.Equal(t, 120, result.width)
	assert.Equal(t, 40, result.height)
	p.AssertExpectations(t)
}

// ---- navCount ----

func TestNavCount_PreviewVisible_AlwaysThree(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c", "d", "e"))
	m := mustModel(t, p)
	m.height = 40
	assert.Equal(t, 3, m.navCount())
}

func TestNavCount_PreviewHidden_FillsHeight(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c", "d", "e"))
	m := mustModel(t, p)
	m.list.showPreview = false
	m.height = 10 // 10 - 3 overhead = 7, but only 5 sessions exist
	assert.Equal(t, 5, m.navCount())
}

func TestNavCount_PreviewHidden_CappedBySessionCount(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b"))
	m := mustModel(t, p)
	m.list.showPreview = false
	m.height = 40
	assert.Equal(t, 2, m.navCount())
}

func TestNavCount_PreviewHidden_MinimumThree(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c", "d"))
	m := mustModel(t, p)
	m.list.showPreview = false
	m.height = 4 // 4 - 3 = 1, below minimum; should clamp to 3
	assert.Equal(t, 3, m.navCount())
}

// ---- visibleItems ----

func TestVisibleItems_Empty(t *testing.T) {
	p := newMockProvider(makePreviews())
	m := mustModel(t, p)
	items, idx := m.visibleItems(3)
	assert.Empty(t, items)
	assert.Equal(t, 0, idx)
}

func TestVisibleItems_SingleSession(t *testing.T) {
	p := newMockProvider(makePreviews("only"))
	m := mustModel(t, p)
	items, idx := m.visibleItems(3)
	assert.Len(t, items, 1)
	assert.Equal(t, 0, idx)
}

func TestVisibleItems_AtFirst_ShowsUpToCount(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c", "d"))
	m := mustModel(t, p)
	items, idx := m.visibleItems(3)
	assert.Equal(t, makePreviews("a", "b", "c"), items)
	assert.Equal(t, 0, idx)
}

func TestVisibleItems_InMiddle_CursorCentred(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c", "d"))
	m := mustModel(t, p)
	m = sendKey(m, "j") // cursor=1
	m = sendKey(m, "j") // cursor=2
	items, idx := m.visibleItems(3)
	assert.Equal(t, makePreviews("b", "c", "d"), items)
	assert.Equal(t, 1, idx)
}

func TestVisibleItems_AtLast_WindowAnchors(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c", "d"))
	m := mustModel(t, p)
	m = sendKey(m, "j")
	m = sendKey(m, "j")
	m = sendKey(m, "j") // cursor=3 (last)
	items, idx := m.visibleItems(3)
	assert.Equal(t, makePreviews("b", "c", "d"), items)
	assert.Equal(t, 2, idx)
}

func TestVisibleItems_TwoSessions_AtFirst(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b"))
	m := mustModel(t, p)
	items, idx := m.visibleItems(3)
	assert.Len(t, items, 2)
	assert.Equal(t, 0, idx)
}

func TestVisibleItems_TwoSessions_AtLast(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b"))
	m := mustModel(t, p)
	m = sendKey(m, "j")
	items, idx := m.visibleItems(3)
	assert.Len(t, items, 2)
	assert.Equal(t, 1, idx)
}

func TestVisibleItems_LargeCount_ShowsAll(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c", "d", "e"))
	m := mustModel(t, p)
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
	p := newMockProvider(makePreviews("s"))
	m := mustModel(t, p)

	lines := m.buildPreviewLines(msgs)
	joined := strings.Join(lines, "\n")
	assert.Contains(t, joined, "USER")
	assert.Contains(t, joined, "ASSISTANT")
	assert.Contains(t, joined, "hello")
	assert.Contains(t, joined, "world")
	p.AssertExpectations(t)
}

func TestBuildPreviewLines_WrapsLongContent(t *testing.T) {
	longMsg := strings.Repeat("x", 200)
	msgs := []message.Message{message.NewUserMessage(longMsg)}
	p := newMockProvider(makePreviews("s"))
	m := mustModel(t, p)
	m.width = 80

	lines := m.buildPreviewLines(msgs)
	for _, l := range lines {
		assert.LessOrEqual(t, len(l), m.width*3, "line unexpectedly long: %q", l)
	}
	p.AssertExpectations(t)
}

// ---- renderPreview ----

func TestRenderPreview_ClampsOffsetBeyondContent(t *testing.T) {
	msgs := []message.Message{message.NewUserMessage("short")}
	p := newMockProvider(makePreviews("s"))
	p.On("HistorySession", "s").Return(session.Session{Messages: msgs}, nil)
	m := mustModel(t, p)
	m.list.previewOffset = 9999

	out := m.renderPreview(10)
	assert.Contains(t, out, "short")
	p.AssertExpectations(t)
}

func TestRenderPreview_ShowsMoreIndicator(t *testing.T) {
	var msgs []message.Message
	for i := range 10 {
		msgs = append(msgs, message.NewUserMessage(fmt.Sprintf("message %d", i)))
	}
	p := newMockProvider(makePreviews("s"))
	p.On("HistorySession", "s").Return(session.Session{Messages: msgs}, nil)
	m := mustModel(t, p)

	out := m.renderPreview(3)
	assert.Contains(t, out, "↓")
	assert.Contains(t, out, "more lines")
	p.AssertExpectations(t)
}

func TestRenderPreview_NoMoreIndicatorWhenAllFits(t *testing.T) {
	msgs := []message.Message{message.NewUserMessage("one line")}
	p := newMockProvider(makePreviews("s"))
	p.On("HistorySession", "s").Return(session.Session{Messages: msgs}, nil)
	m := mustModel(t, p)

	out := m.renderPreview(50)
	assert.NotContains(t, out, "more lines")
	p.AssertExpectations(t)
}

func TestRenderPreview_ScrollRevealsNewContent(t *testing.T) {
	var msgs []message.Message
	for i := range 20 {
		msgs = append(msgs, message.NewUserMessage(fmt.Sprintf("msg-%d", i)))
	}
	p := newMockProvider(makePreviews("s"))
	p.On("HistorySession", "s").Return(session.Session{Messages: msgs}, nil)
	m := mustModel(t, p)

	before := m.renderPreview(5)
	m.list.previewOffset = 10
	after := m.renderPreview(5)

	assert.NotEqual(t, before, after)
	p.AssertExpectations(t)
}

func TestRenderPreview_ShowsErrorOnSessionFetchFailure(t *testing.T) {
	fetchErr := fmt.Errorf("disk error")
	p := newMockProvider(makePreviews("s"))
	p.On("HistorySession", "s").Return(session.Session{}, fetchErr)
	m := mustModel(t, p)

	out := m.renderPreview(10)
	assert.Contains(t, out, "Error loading session")
	p.AssertExpectations(t)
}

// ---- View ----

func TestView_HeaderShowsSessionCount(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c"))
	p.On("HistorySession", "a").Return(session.Session{}, nil)
	m := mustModel(t, p)
	assert.Contains(t, m.View(), "Session history (3 sessions)")
	p.AssertExpectations(t)
}

func TestView_NavigatorShowsAbsoluteNumbers(t *testing.T) {
	p := newMockProvider(makePreviews("a", "b", "c", "d", "e"))
	p.On("HistorySession", "c").Return(session.Session{}, nil)
	m := mustModel(t, p)
	m = sendKey(m, "j")
	m = sendKey(m, "j") // cursor on "c" (index 2), window shows [b, c, d]
	view := m.View()
	assert.Contains(t, view, "2.")
	assert.Contains(t, view, "3.")
	assert.Contains(t, view, "4.")
	p.AssertExpectations(t)
}

func TestView_EmptyPreviewsMessage(t *testing.T) {
	p := newMockProvider(makePreviews())
	m := mustModel(t, p)
	assert.Contains(t, m.View(), "No sessions found")
	p.AssertExpectations(t)
}

func TestView_SessionNameAppearsInNavigator(t *testing.T) {
	p := newMockProvider(makePreviews("my-session"))
	p.On("HistorySession", "my-session").Return(session.Session{}, nil)
	m := mustModel(t, p)
	assert.Contains(t, m.View(), "my-session")
	p.AssertExpectations(t)
}

func TestView_DateAppearsInNavigator(t *testing.T) {
	p := newMockProvider(makePreviews("s"))
	p.On("HistorySession", "s").Return(session.Session{}, nil)
	m := mustModel(t, p)
	assert.Contains(t, m.View(), "Jan 15 2024 10:00")
	p.AssertExpectations(t)
}

func TestView_PreviewShowsMessages(t *testing.T) {
	msgs := []message.Message{
		message.NewUserMessage("what is the capital of France"),
		message.NewAssistantMessage("Paris"),
	}
	p := newMockProvider(makePreviews("s1"))
	p.On("HistorySession", "s1").Return(session.Session{Messages: msgs}, nil)
	m := mustModel(t, p)
	view := m.View()
	assert.Contains(t, view, "USER")
	assert.Contains(t, view, "ASSISTANT")
	assert.Contains(t, view, "what is the capital of France")
	assert.Contains(t, view, "Paris")
	p.AssertExpectations(t)
}

func TestView_PreviewHiddenWhenToggled(t *testing.T) {
	msgs := []message.Message{message.NewUserMessage("hello")}
	p := newMockProvider(makePreviews("s1"))
	p.On("HistorySession", "s1").Return(session.Session{Messages: msgs}, nil).Once()
	m := mustModel(t, p)
	require.Contains(t, m.View(), "hello") // baseline: content visible before toggle
	m = sendKey(m, "p")                    // hide preview; HistorySession must not be called again
	assert.NotContains(t, m.View(), "hello")
	p.AssertExpectations(t)
}

func TestView_PreviewShowsLoadError(t *testing.T) {
	sentinel := fmt.Errorf("disk error")
	p := newMockProvider(makePreviews("bad"))
	p.On("HistorySession", "bad").Return(session.Session{}, sentinel)
	m := mustModel(t, p)
	assert.Contains(t, m.View(), "Error loading session")
	p.AssertExpectations(t)
}

func TestView_HelpShowsScrollHintWhenPreviewVisible(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	p.On("HistorySession", "a").Return(session.Session{}, nil)
	m := mustModel(t, p)
	assert.Contains(t, m.View(), "ctrl+↑/↓ scroll preview")
	p.AssertExpectations(t)
}

func TestView_HelpHidesScrollHintWhenPreviewHidden(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendKey(m, "p")
	assert.NotContains(t, m.View(), "ctrl+↑/↓ scroll preview")
	p.AssertExpectations(t)
}

func TestView_HelpShowsHidePreviewWhenVisible(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	p.On("HistorySession", "a").Return(session.Session{}, nil)
	m := mustModel(t, p)
	assert.Contains(t, m.View(), "p hide preview")
	p.AssertExpectations(t)
}

func TestView_HelpShowsShowPreviewWhenHidden(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	m := mustModel(t, p)
	m = sendKey(m, "p")
	assert.Contains(t, m.View(), "p show preview")
	p.AssertExpectations(t)
}

func TestView_HelpShowsDeleteHint(t *testing.T) {
	p := newMockProvider(makePreviews("a"))
	p.On("HistorySession", "a").Return(session.Session{}, nil)
	m := mustModel(t, p)
	assert.Contains(t, m.View(), "d delete")
	p.AssertExpectations(t)
}

func TestView_StatusMessageAppearsAfterDeleteError(t *testing.T) {
	deleteErr := fmt.Errorf("permission denied")
	p := newMockProvider(makePreviews("a"))
	p.On("HistoryDelete", "a").Return(deleteErr)
	p.On("HistorySession", "a").Return(session.Session{}, nil)
	m := mustModel(t, p)
	m = sendKey(m, "d")
	assert.Contains(t, m.View(), "permission denied")
	p.AssertExpectations(t)
}

// ---- wordWrap ----

func TestWordWrap_ShortLineUnchanged(t *testing.T) {
	result := wordWrap("hello", 20)
	assert.Equal(t, "hello", result)
}

func TestWordWrap_LongLineBreaks(t *testing.T) {
	result := wordWrap("abcdefghij", 4)
	assert.Equal(t, "abcd\nefgh\nij", result)
}

func TestWordWrap_PreservesExistingNewlines(t *testing.T) {
	result := wordWrap("abc\ndef", 10)
	assert.Equal(t, "abc\ndef", result)
}

func TestWordWrap_ZeroWidthReturnsUnchanged(t *testing.T) {
	result := wordWrap("hello", 0)
	assert.Equal(t, "hello", result)
}

func TestWordWrap_EmptyString(t *testing.T) {
	result := wordWrap("", 10)
	assert.Equal(t, "", result)
}
