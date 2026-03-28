# Interface Design for Testability

Good interfaces make testing natural:

**1. Accept dependencies, don't create them**

**Go:**

```go
// Testable — dependency injected
func ProcessOrder(ctx context.Context, order Order, gw PaymentGateway) error {
    return gw.Charge(ctx, order.Total)
}

// Hard to test — dependency created internally
func ProcessOrder(ctx context.Context, order Order) error {
    gw := stripe.NewGateway(os.Getenv("STRIPE_KEY"))
    return gw.Charge(ctx, order.Total)
}
```

**TypeScript:**

```typescript
// Testable
function processOrder(order, paymentGateway) {}

// Hard to test
function processOrder(order) {
  const gateway = new StripeGateway();
}
```

**2. Return results, don't produce side effects**

**Go:**

```go
// Testable — returns a value
func CalculateDiscount(cart Cart) (Discount, error) { ... }

// Hard to test — mutates input
func ApplyDiscount(cart *Cart) { cart.Total -= discount }
```

**TypeScript:**

```typescript
// Testable
function calculateDiscount(cart): Discount {}

// Hard to test
function applyDiscount(cart): void {
  cart.total -= discount;
}
```

**3. Small surface area**

- Fewer methods = fewer tests needed
- Fewer params = simpler test setup
