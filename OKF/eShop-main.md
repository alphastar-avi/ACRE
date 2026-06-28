# Open Knowledge Format (OKF) Codebase Index: eShop-main

## Overview
This is a standard .NET 10 multi-project web application (Microsoft eShopOnWeb reference architecture). It consists of several domain layers, a web presentation layer, and multiple test suites.

## Repository Architecture

### 1. Presentation Layer (`src/Web/`)
This is the main MVC and Razor Pages project. It handles routing, page rendering, controllers, and presentation model mappings.
* **Basket Page Entry Point**: `src/Web/Pages/Basket/Index.cshtml` and backend logic in `Index.cshtml.cs`.
* **Basket Mappings & Views**: `BasketViewModel.cs` and `BasketItemViewModel.cs` under `src/Web/Pages/Basket/`.
* **Basket ViewModel Service**: `src/Web/Services/BasketViewModelService.cs` (implements `IBasketViewModelService`). This service queries catalog and basket items and maps them into presentation models.
* **Controllers**: Controllers reside under `src/Web/Controllers/`. E.g., `OrderController.cs` handles order viewing.

### 2. Domain / Core Layer (`src/ApplicationCore/`)
Contains entities, domain services, specification patterns, and interfaces.
* **Basket Entities**: Domain models `Basket` and `BasketItem` under `src/ApplicationCore/Entities/BasketAggregate/`.
* **Basket Services**: `src/ApplicationCore/Services/BasketService.cs` (implements `IBasketService`) handles updating items, clearing carts, and setting quantities.
* **Order Services**: `src/ApplicationCore/Services/OrderService.cs` (implements `IOrderService`) manages order creation and checkout processing.
* **Uri Composer Service**: `src/ApplicationCore/Services/UriComposer.cs` (implements `IUriComposer`) replaces template URLs with active catalog image paths.
* **Specifications**: Core filters (e.g., `BasketWithItemsSpecification.cs`, `CatalogItemsSpecification.cs`) under `src/ApplicationCore/Specifications/`.

### 3. Infrastructure Layer (`src/Infrastructure/`)
Handles data persistence, entity framework configurations, identity seed database contexts, and concrete repository implementations.
* **EF Configurations**: Entity configurations (e.g., `BasketConfiguration.cs`, `CatalogItemConfiguration.cs`) under `src/Infrastructure/Data/Config/`.
* **Database Contexts**: `CatalogContext.cs` (app DB) and `AppIdentityDbContext.cs` (user DB) under `src/Infrastructure/Data/` and `src/Infrastructure/Identity/`.

---

## Test Suites (`tests/`)
The codebase has multiple xUnit test projects:
1. **Unit Tests** (`tests/UnitTests/`):
   * Focuses on domain objects and basic services (e.g. `tests/UnitTests/ApplicationCore/Entities/BasketTests/`).
2. **Functional Tests** (`tests/FunctionalTests/`):
   * Tests presentation and routing logic (e.g. `tests/FunctionalTests/Web/Pages/Basket/IndexTest.cs`).
3. **Integration Tests** (`tests/IntegrationTests/`):
   * Verifies repository queries and DB context integrations.
4. **Public API Tests** (`tests/PublicApiIntegrationTests/`):
   * Asserts correct behavior of API endpoints under `src/PublicApi/`.

---

## Solution Configuration
* **`eShopOnWeb.sln`**: The standard solution file, which includes the docker-compose project (`docker-compose.dcproj`).
* **`Everything.sln`**: A cleaner solution excluding docker-compose. **Always prefer building and testing against `Everything.sln`** to avoid failures related to missing Docker Desktop containers.
