# Test Anti-Patterns

A catalog of bad test patterns. Each section names the pattern, shows what it
looks like in Go and TypeScript, and explains why it should be deleted or
refactored.

---

## 1. Testing Internal Call Counts

Asserts that a specific internal method was called, rather than checking what
the system produced.

**Go:**
```go
// BAD
func TestCheckout_CallsPaymentGateway(t *testing.T) {
    mock := &mockPaymentGateway{}
    checkout(cart, mock)
    assert.True(t, mock.chargeCalled)
    assert.Equal(t, cart.Total, mock.chargeAmount)
}
```

**TypeScript:**
```typescript
// BAD
test("checkout calls gateway.charge", async () => {
  const gateway = { charge: jest.fn().mockResolvedValue({ id: "x" }) };
  await checkout(cart, gateway);
  expect(gateway.charge).toHaveBeenCalledWith(cart.total);
});
```

**Why it's bad**: The test breaks whenever `charge` is renamed, inlined, or
replaced — even if the user still gets charged correctly. It tests HOW, not
WHAT.

**Fix**: Assert on the observable outcome — the returned receipt, the updated
order status, the emitted event.

---

## 2. Mocking Internal Collaborators

Mocking a class or module that you own and control.

**Go:**
```go
// BAD — OrderRepository is internal, not a system boundary
func TestOrderService_Create(t *testing.T) {
    repo := &mockOrderRepository{}
    svc := NewOrderService(repo)
    svc.Create(ctx, orderInput)
    assert.True(t, repo.saveCalled)
}
```

**TypeScript:**
```typescript
// BAD — userService is internal
test("createOrder calls userService.get", async () => {
  jest.mock("../userService");
  const { get } = require("../userService");
  await createOrder(orderData);
  expect(get).toHaveBeenCalledWith(orderData.userId);
});
```

**Why it's bad**: Creates tight coupling to implementation decisions. Moving
logic between collaborators breaks tests without changing behavior.

**Fix**: Use a real implementation or an in-memory fake. Mock only at system
boundaries (HTTP clients, email senders, the database driver itself).

---

## 3. Testing Private Methods Directly

Reaches into unexported/private methods to test them in isolation.

**Go:**
```go
// BAD — tests unexported function directly
func Test_calculateLineItems(t *testing.T) {
    result := calculateLineItems(rawItems)
    assert.Equal(t, expected, result)
}
```

**TypeScript:**
```typescript
// BAD — accesses private method via type cast
test("_buildQuery constructs SQL correctly", () => {
  const repo = new UserRepository(db) as any;
  const sql = repo._buildQuery({ name: "Alice" });
  expect(sql).toContain("WHERE name =");
});
```

**Why it's bad**: Private methods are implementation details. They can be
merged, split, or renamed freely. Testing them directly locks the
implementation in place.

**Fix**: Delete the test. The behavior covered by the private method is already
exercised through the public API. If it isn't, write a public-API-level test
instead.

---

## 4. Verifying Through External Means

After calling the public interface, checks side effects by reaching outside the
interface (e.g. raw SQL queries, reading files directly, inspecting internal
state).

**Go:**
```go
// BAD — bypasses interface to check DB state
func TestCreateUser_PersistsData(t *testing.T) {
    createUser(ctx, db, User{Name: "Alice"})
    var count int
    db.QueryRow("SELECT COUNT(*) FROM users WHERE name = 'Alice'").Scan(&count)
    assert.Equal(t, 1, count)
}
```

**TypeScript:**
```typescript
// BAD — reads file directly instead of using the API
test("saveReport writes to disk", async () => {
  await saveReport(report, "/tmp/report.json");
  const content = fs.readFileSync("/tmp/report.json", "utf8");
  expect(JSON.parse(content)).toEqual(report);
});
```

**Why it's bad**: Breaks if the storage mechanism changes (different DB schema,
different file format) even though behavior is identical.

**Fix**: Verify through the same interface the caller would use.

```go
// GOOD
func TestCreateUser_IsRetrievable(t *testing.T) {
    user, err := createUser(ctx, store, User{Name: "Alice"})
    require.NoError(t, err)
    retrieved, err := getUser(ctx, store, user.ID)
    require.NoError(t, err)
    assert.Equal(t, "Alice", retrieved.Name)
}
```

---

## 5. Redundant Happy-Path Duplicates

Multiple tests for the same success path with trivially different inputs where
one would suffice.

**Go:**
```go
// BAD — both test the same behavior at the same level
func TestAdd_TwoPositiveNumbers(t *testing.T) {
    assert.Equal(t, 5, add(2, 3))
}
func TestAdd_TwoOtherPositiveNumbers(t *testing.T) {
    assert.Equal(t, 10, add(4, 6))
}
```

**TypeScript:**
```typescript
// BAD — identical structure, only input differs
test("formats name Alice", () => expect(formatName("Alice")).toBe("Alice"));
test("formats name Bob", () => expect(formatName("Bob")).toBe("Bob"));
```

**Why it's bad**: Adding more examples of the exact same behavior adds noise
without adding confidence. One table-driven test covers all cases with less
overhead.

**Fix**: Merge into a single table-driven / parameterized test, or delete all
but one.

---

## 6. Asserting on Test-Setup Values

The assertion only verifies data that was set up in the test itself — the
system under test made no meaningful contribution.

**TypeScript:**
```typescript
// BAD — asserts on the input it just created
test("cart stores items", () => {
  const cart = new Cart();
  cart.add({ id: 1, price: 10 });
  expect(cart.items[0].id).toBe(1);   // trivially true — we just put it there
  expect(cart.items[0].price).toBe(10);
});
```

**Go:**
```go
// BAD — asserts that a struct field was assigned
func TestNewOrder_HasCorrectID(t *testing.T) {
    order := NewOrder("order-123", items)
    assert.Equal(t, "order-123", order.ID)
}
```

**Why it's bad**: These tests cannot fail under any realistic failure mode in
the business logic. They are structural tests of the language, not behavioral
tests of the system.

**Fix**: Delete, or replace with a test that verifies a non-trivial computed
outcome (total price, discount applied, validation enforced).

---

## 7. Over-specified Error Messages

Asserts on the exact wording of an error string.

**Go:**
```go
// BAD
require.EqualError(t, err, "payment failed: card declined: insufficient funds on card ending 4242")
```

**TypeScript:**
```typescript
// BAD
expect(err.message).toBe("Payment failed: card declined: insufficient funds on card ending 4242");
```

**Why it's bad**: Error messages are UI concerns. Rephrasing for better UX
breaks the test.

**Fix**: Assert on error type, code, or a stable sentinel — not the full
message string.

```go
// GOOD
var paymentErr *PaymentError
require.ErrorAs(t, err, &paymentErr)
assert.Equal(t, "card_declined", paymentErr.Code)
```

---

## 8. Tests That Never Fail

A test with no assertions, or only `assert.True(t, true)`, or that catches all
errors silently.

**Go:**
```go
// BAD — no assertions
func TestProcessOrder(t *testing.T) {
    err := processOrder(ctx, order)
    _ = err
}
```

**TypeScript:**
```typescript
// BAD — catch swallows the error
test("handles errors", async () => {
  try {
    await riskyOperation();
  } catch (_) {}
});
```

**Why it's bad**: These tests always pass — they protect nothing.

**Fix**: Delete or add a meaningful assertion. If an error is expected, assert
`expect(fn).rejects.toThrow(...)` / `require.Error(t, err)`.
