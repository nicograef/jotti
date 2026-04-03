# Evaluation Criteria

Use this rubric during Step 2 (Audit) to tag each test as **Keep**,
**Refactor**, **Delete**, or **Merge**.

---

## Decision Tree

Work through these questions for each test:

```
1. Does the test make at least one meaningful assertion?
   └─ NO  → DELETE (Tests That Never Fail)

2. Does the test assert only on values the test itself created,
   with no non-trivial logic from the system?
   └─ YES → DELETE (Asserting on Setup Values)

3. Does the test access private/unexported methods or fields directly?
   └─ YES → DELETE (replace with a public-API test if the behavior matters)

4. Does the test assert on internal call counts, argument order,
   or mock invocations of code you own?
   └─ YES → DELETE (or REFACTOR if the same behavior can be verified via output)

5. Does the test mock internal collaborators (classes/modules you own)?
   └─ YES → REFACTOR (replace mocks with real or in-memory implementations)

6. Does the test verify state by bypassing the public interface
   (raw SQL, file reads, internal object inspection)?
   └─ YES → REFACTOR (verify through the public interface instead)

7. Is the test name phrased as HOW the system works
   ("calls X", "invokes Y", "sets flag Z") rather than WHAT it does?
   └─ YES → likely REFACTOR (rewrite intent + implementation)

8. Does an identical behavior already have 2+ other tests
   with only trivially different inputs?
   └─ YES → MERGE (keep one; fold the others into a table-driven test)

9. Does the test over-specify error messages (exact string match)?
   └─ YES → REFACTOR (assert on type / code instead)

All NO → KEEP
```

---

## Tag Definitions

### Keep

The test:
- Asserts on observable output or public state
- Uses the public API only
- Would survive a complete internal refactor of the unit under test
- Name describes WHAT the system does ("user can checkout with valid cart")

No changes needed.

### Refactor

The test has the right intent but is implemented poorly:
- Uses real collaborators but reaches through them to check internals
- Correct behavior assertion but test is structured around implementation
- Mocks internal collaborators instead of boundaries
- Bypasses the public interface for verification

Action: rewrite to assert on output / return value / public state only.

### Delete

The test one or more of:
- Tests a private method directly
- Only asserts on call counts / argument order for internal methods
- Has no assertions (never fails)
- Only asserts on values the test itself provided
- Is fully redundant — covered identically by another test

Action: remove the test. If behavior coverage is lost, note it in the step 3
report — the user may want to add a proper replacement via TDD.

### Merge

Two or more tests:
- Test the same behavior with trivially different inputs
- Could be expressed as one table-driven / parameterized test
- Are exact duplicates

Action: combine into a single test (table-driven in Go, `test.each` in Jest /
Vitest, `@pytest.mark.parametrize` in pytest).

---

## Coverage Loss Protocol

When a Delete tag removes the only test for a behavior:

1. Note it explicitly in the Step 3 report:
   ```
   ⚠ Delete "test name" removes the only coverage for [behavior X].
     Suggest: add a proper test via TDD skill after this review.
   ```
2. Do not add the replacement test during this skill — that is out of scope.
3. Confirm with the user before deleting.

---

## Mocking Boundary Reference

**Mock these** (system boundaries — real calls are slow, unreliable, or
destructive):

| Boundary | Examples |
|----------|---------|
| External HTTP APIs | Payment gateways, email providers, SMS |
| Database driver (optional) | Prefer a real test DB or in-memory DB |
| Time / randomness | `time.Now()`, `uuid.New()`, `Math.random()` |
| File system (sometimes) | When testing output format, not I/O mechanics |
| Message queues / event buses | Kafka, SQS, RabbitMQ |

**Do not mock these** (you own them):

| Code | What to do instead |
|------|--------------------|
| Your own services / use cases | Use the real implementation |
| Your own repositories | Use an in-memory fake or test DB |
| Internal helper functions | Call them through the public API |
| Value objects / domain models | Instantiate them directly |

---

## Quick Reference Card

| Signal | Tag |
|--------|-----|
| `assert.True(t, mock.called)` | Delete or Refactor |
| `expect(fn).toHaveBeenCalledWith(...)` on internal code | Delete or Refactor |
| `db.QueryRow(...)` in a test that uses a service | Refactor |
| Private method access (`as any`, `//nolint:unexported`) | Delete |
| Test name starts with "calls", "invokes", "sets" | Refactor |
| Two tests, same structure, only literals differ | Merge |
| `catch (_) {}` with no assertion | Delete |
| `_ = err` with no assertion | Delete |
| Asserts on `result.id` when the test itself provided `id` | Delete |
| Asserts on computed total, status, formatted output | Keep |
| Uses only public API; test name is behavior-first | Keep |
