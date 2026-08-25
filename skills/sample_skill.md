# Engineering & Domain Skill: Core Business & Codebase Rules

This skill document defines team-specific coding standards, business logic invariants, and architecture guidelines to be enforced during automated remediation.

---

## 1. Domain & Business Logic Invariants
* **Monetary & Pricing Calculations**: Always use high-precision decimals (`decimal` in C# / Go `shopspring/decimal` / Python `Decimal`). NEVER perform arithmetic on currency using floating-point types (`float`, `double`).
* **Tenant & Organization Isolation**: Every database query and cache lookup MUST include the `TenantId` / `OrgId` filter. Never perform cross-tenant data retrieval.
* **Idempotency**: All payment, refund, and webhook processing endpoints must enforce idempotent request keys (`Idempotency-Key` header) before state transitions.

---

## 2. Codebase Architecture & Coding Standards
* **Layering & Separation of Concerns**:
  - Keep controllers/handlers thin. Business logic must reside in Domain Services or Command Handlers.
  - Never execute direct SQL queries or database context calls inside API Controllers or Presentation layers.
* **Error Handling & Logging**:
  - Never swallow exceptions silently with empty `catch` / `except` blocks.
  - Use structured logging (e.g. `logger.LogError(ex, "Failed to process order {OrderId}", orderId)`) rather than raw string concatenation.
  - Do NOT log Sensitive Personal Identifiable Information (PII), Passwords, or API Tokens.
* **Null Safety & Defensive Checks**:
  - Always validate input arguments and preconditions at the beginning of public methods (Guard clauses).
  - Explicitly handle null/empty states for optional relationships and collections.

---

## 3. Testing Conventions
* **Unit Test Framework**:
  - Write test methods using the `MethodName_StateUnderTest_ExpectedBehavior` naming pattern.
  - Follow the **Arrange-Act-Assert (AAA)** pattern with clean separation.
* **Mocking & Test Isolation**:
  - Mock external network calls, third-party payment gateways, and email dispatchers.
  - Do not use `Thread.Sleep` / arbitrary timeouts in tests; use proper async/await signaling or fake clocks.

---

## 4. Prohibited Anti-Patterns (Do NOT Introduce)
* ❌ Direct string concatenation into SQL commands, shell executions, or file paths.
* ❌ Bypassing authorization / authentication middleware or permission checks.
* ❌ Hardcoding configuration values, secrets, connection strings, or base URLs in source code.
* ❌ Introducing breaking schema/API changes without backward compatibility support.
