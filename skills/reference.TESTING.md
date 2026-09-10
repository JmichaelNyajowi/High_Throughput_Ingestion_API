# Testing

Test types, seams, and the test-first rules. Used by `/foundation`, `/build`, `/review`, and `/fix`.

## The three types

**Integration — the default and the majority.**
Several pieces together through a real path, against a real test database. Slower than unit tests, but they survive refactors because they only care about observable behaviour.

**Unit — the exception.**
One function in isolation. Only for genuinely tricky pure logic: a pricing calculation, a date rule, a tax computation. Unit tests break when code is rearranged even though behaviour didn't change, which makes refactoring painful and makes tests untrustworthy.

**E2E — five to ten for a whole project.**
A browser driving the real app. Critical paths only: login, the main create flow, the main list view, anything involving money or permissions. A large E2E suite becomes a maintenance burden nobody wants to own.

## Never mock our own code

Mock only at true system boundaries:
- Third-party APIs (payment, email)
- Time and randomness
- Sometimes the filesystem

**Never mock:** your own modules, internal collaborators, the data layer.

Mocked data-access tests pass while the real query is wrong — the exact bug class most worth catching.

At real boundaries, design for mockability: pass dependencies in rather than constructing them inside, and prefer specific functions per operation over one generic fetcher.

## Seams

**A seam is where a test observes behaviour without reaching inside** — a module's interface.

**Test only at confirmed seams.** They're decided in Planning and listed in `docs/architecture.md`. **No test is written at a seam that isn't on that list.** If a new seam seems necessary, stop and ask.

Prefer existing seams. Prefer the highest seam available. Fewer is better.

## What a good test is

Tests verify behaviour through public interfaces, not implementation details. The code can change entirely; the test shouldn't. A good test reads like a specification — `"user can checkout with valid cart"` tells you what capability exists.

```ts
// GOOD — observable behaviour through the public interface
test("user can checkout with valid cart", async () => {
  const cart = createCart();
  cart.add(product);
  const result = await checkout(cart, paymentMethod);
  expect(result.status).toBe("confirmed");
});
```

```ts
// BAD — coupled to implementation
test("checkout calls paymentService.process", async () => {
  const mock = jest.mock(paymentService);
  await checkout(cart, payment);
  expect(mock.process).toHaveBeenCalledWith(cart.total);
});
```

```ts
// BAD — bypasses the interface to verify
test("createUser saves to database", async () => {
  await createUser({ name: "Alice" });
  const row = await db.query("SELECT * FROM users WHERE name = ?", ["Alice"]);
  expect(row).toBeDefined();
});

// GOOD — verifies through the interface
test("createUser makes user retrievable", async () => {
  const user = await createUser({ name: "Alice" });
  expect((await getUser(user.id)).name).toBe("Alice");
});
```

## Anti-patterns

**Implementation-coupled** — mocks internal collaborators, tests private functions, or verifies through a side channel. *The tell: it breaks when you refactor but behaviour hasn't changed.*

**Tautological** — the expected value is computed the way the code computes it, so it passes by construction and can never disagree.

```ts
// BAD — recomputes the implementation
const expected = items.reduce((sum, i) => sum + i.price, 0);
expect(calculateTotal(items)).toBe(expected);

// GOOD — an independent known literal
expect(calculateTotal([{ price: 10 }, { price: 5 }])).toBe(15);
```

**Horizontal slicing** — writing all tests first, then all implementation. Bulk tests verify *imagined* behaviour: you test the shape of things rather than what users need, and you commit to test structure before understanding the implementation.

## Test-first vs test-after

**The ticket declares which.** The judgment is made when the ticket is cut, not mid-build.

**Logic → test-first.** Anything with rules: pricing, permissions, state transitions, validation, edge cases.

**Plumbing → test-after or none.** CRUD, wiring, config, layout, styling.

### Why test-first for logic

It costs maybe 10–20% more within a ticket. It's cheaper overall for two reasons.

**First:** an agent that writes code and then tests writes a test that *passes*, because it already believes the code is correct. That's a description, not a check. The bug surfaces three weeks later as a client phone call — the genuinely expensive path.

**Second, and more important:** writing the test first forces the interface to be decided before the implementation. Code you must call before writing designs itself to be callable — clean inputs, clean outputs, no hidden dependencies. **It's a design tool disguised as a testing practice.**

### The loop

**Per behaviour, not per ticket:**

1. Write one failing test
2. Watch it fail — a test that passes immediately is testing nothing
3. Write the minimum implementation to pass it
4. Watch it pass
5. Next behaviour

**Never write all the tests up front.** One at a time lets each test learn from what the last one taught you.

**Red before green.** Don't anticipate future tests or add speculative features.

**Refactoring is not part of the loop.** It belongs to review.

## Cadence

- **Typecheck continuously** — the cheapest real feedback available
- **Run the tests you touched continuously**
- **Run the full suite once, at the end**
