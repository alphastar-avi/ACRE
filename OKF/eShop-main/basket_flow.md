---
type: Feature Flow
title: Shopping Cart / Basket Flow Mapping
description: Reference details on controllers, services, entities, and UI pages involved in basket management.
resource: CodeBase/eShop-main/src/Web/Pages/Basket/
tags: [basket, cart, flow]
timestamp: 2026-06-29T00:00:00Z
---

# Basket & Shopping Cart Flow

The basket management flow coordinates interactions across UI controllers, presentation mapping services, domain models, and database persistence.

## Key Files & Layers

### 1. The Presentation/UI Controller
* **File**: `src/Web/Pages/Basket/Index.cshtml.cs` & `Index.cshtml`
* **Purpose**: Coordinates incoming user page requests, handles updating item quantities, and maps the basket data model using the `BasketViewModelService`.

### 2. Presentation Model Mapping
* **File**: `src/Web/Services/BasketViewModelService.cs` (implements `IBasketViewModelService`)
* **Purpose**: Maps database domain entities (`Basket`, `BasketItem`, and `CatalogItem`) to frontend ViewModels (`BasketViewModel`, `BasketItemViewModel`).
* **Critical Picture Resolver**: Invokes `_uriComposer.ComposePicUri(catalogItem.PictureUri)` to compile final catalog image URLs.

### 3. Core Domain Logic
* **File**: `src/ApplicationCore/Services/BasketService.cs` (implements `IBasketService`)
* **Purpose**: Performs state modifications (adding items, updating quantities, deleting/clearing baskets) directly on domain aggregates.
* **Specification Filtering**: Leverages specifications (like `BasketWithItemsSpecification.cs`) to eagerly load basket relations.

### 4. Domain Aggregates
* **Files**: `src/ApplicationCore/Entities/BasketAggregate/Basket.cs` & `BasketItem.cs`
* **Purpose**: Domain models representing customer shopping carts.

## Diagnostic Flow Tips

When diagnosing a crash or NRE on the Basket Page:
1. Trace `BasketViewModelService.cs` to check how the catalog items are mapped into `BasketItemViewModel`. If catalog items are missing (e.g. deleted by an admin), `FirstOrDefault()` can return `null` and cause `NullReferenceException` on picture compositing.
2. Confirm null-safe access of properties like `catalogItem.PictureUri` and `catalogItem.Name`.
