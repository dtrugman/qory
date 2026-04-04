---
name: "review-unittest"
description: "Use when the user asks to write or review unit-tests"
---

# Golang unittest expert

**Tier:** POWERFUL
**Category:** Engineering
**Domain:** Code Review / Quality Assurance

---

## Overview

The perfect golang unittest for production grade code.
Unit-tests are the best tool we have to make sure our application keeps on working while changing rapidly.
Writing unit-tests is more than just getting them to pass:
- They should provide great coverage, but avoid covering the same functionality over and over again
- They should be simple to understand
- They should be easy to change

---

## When to use

- When writing unit-tests
- When asked to review a unit-test

---

## Workflow

### Step 1 — When reviewing, identify the files to review

You should be able to find those in the existing conversation.
If not, ask the user which files to cover.

### Step 2 — Make sure we use the right framework

All tests should be using testify assert/require and testify Mock-s.

Never implement mock objects manually.
Never use t.Fail or similar calls.

### Step 3 — Make sure we always call AssertExpectations on all mock objects

We should call AssertCalled or AssertNotCalled only if AssertExpectations doesn't cover that.

```
sm.AssertNotCalled(t, "Store", mock.Anything, mock.Anything) <-- NOT REQUIRED
sm.AssertExpectations(t)
```

### Step 4 — Happy flows should check arguments passed to mocks

Consider this mock expectation setup:
```
client.On("Query", "gpt-4o", mock.Anything).Return(...)
```

The second argument (`mock.Anything` here), is the actual payload we pass to the client.
It is critical to check that the correct payload is passed as an argument here.

However, if a following test is checking an error flow:
```
client.On("Query", "gpt-4o", mock.Anything).Return("", queryErr)
```

And we already tested the argument for the happy flow, we can use mock.Anything.

### Step 5 — Never inline calls

Each function call should get its own line:
```
err := q.QueryNew([]string{"hello"})
require.ErrorIs(t, err, queryErr)
```

Avoid nesting calls such as `require.Error(t, q.QueryNew(...))` as this is harder to read.

This is true when calling non-test methods as well.

Bad:
```
err := q.QueryNew(getPayload(messages))
```

Good:
```
payload := getPayload(messages)
err := q.QueryNew(payload)
```

### Step 6 — Avoid utility methods for creaeing pointers

We have a `util.Ptr` method for creating pointers. Use it instead of creating new methods.

### Step 7 — Avoid branches in the test flow as much as possible

If we have 10 tests for a method, some of which are expected to fail, and others to success,
Separate into two Test methods or wrap using two different t.Run()-s, so that the actual logic
within the test avoids 'if' conditions for setup and result testing.

### Step 8 — Consider the coverage of tests

Make sure tests do not overlap.
We should aim for maximal coverage with a minimal amount of code.

For example, e have logic that adds a system prompt to an AI inference request only when sending the first message.
Instead of writing a test solely for that, incorporate that logic into the main flow test:

- When querying with a new session - expect the prompt to be injected
- When querying with an existing session or the last session - prompt should not be injected

---

## Complete review checklist

```markdown
## Code Review Checklist
- [ ] Using testify
- [ ] Always AssertExpectations
- [ ] Avoid other Asserts, unless strictly required
- [ ] Happy flows check all arguments passed to mocks
- [ ] No inlined calls
- [ ] Minimize execution branches in the test flow
```
