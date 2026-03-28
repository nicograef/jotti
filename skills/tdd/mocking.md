# When to Mock

Mock at **system boundaries** only:

- External APIs (payment, email, etc.)
- Databases (sometimes — prefer test DB)
- Time/randomness
- File system (sometimes)

Don't mock:

- Your own classes/modules
- Internal collaborators
- Anything you control

## Designing for Mockability

At system boundaries, design interfaces that are easy to mock:

**1. Use dependency injection**

Pass external dependencies in rather than creating them internally.

**Go:**

```go
// Easy to mock — accepts interface
func ProcessPayment(order Order, client PaymentClient) (Receipt, error) {
    return client.Charge(order.Total)
}

// Hard to mock — creates dependency internally
func ProcessPayment(order Order) (Receipt, error) {
    client := stripe.NewClient(os.Getenv("STRIPE_KEY"))
    return client.Charge(order.Total)
}
```

**TypeScript:**

```typescript
// Easy to mock
function processPayment(order, paymentClient) {
  return paymentClient.charge(order.total);
}

// Hard to mock
function processPayment(order) {
  const client = new StripeClient(process.env.STRIPE_KEY);
  return client.charge(order.total);
}
```

**2. Prefer specific interfaces over generic ones**

Create specific functions/methods for each external operation instead of one
generic function with conditional logic.

**Go:**

```go
// GOOD: Each method is independently mockable
type UserAPI interface {
    GetUser(ctx context.Context, id string) (User, error)
    ListOrders(ctx context.Context, userID string) ([]Order, error)
    CreateOrder(ctx context.Context, data OrderInput) (Order, error)
}

// BAD: Mocking requires conditional logic
type API interface {
    Do(ctx context.Context, method, path string, body any) (any, error)
}
```

**TypeScript:**

```typescript
// GOOD: Each function is independently mockable
const api = {
  getUser: (id) => fetch(`/users/${id}`),
  getOrders: (userId) => fetch(`/users/${userId}/orders`),
  createOrder: (data) => fetch("/orders", { method: "POST", body: data }),
};

// BAD: Mocking requires conditional logic inside the mock
const api = {
  fetch: (endpoint, options) => fetch(endpoint, options),
};
```

The specific-interface approach means:

- Each mock returns one specific shape
- No conditional logic in test setup
- Easier to see which operations a test exercises
- Type safety per operation
