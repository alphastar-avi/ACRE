---
type: Codebase Overview
title: eShop-main Reference Architecture
description: Microsoft eShopOnWeb reference ASP.NET Core e-commerce application.
resource: CodeBase/eShop-main
tags: [dotnet, eshop, core]
timestamp: 2026-06-29T00:00:00Z
---

# eShop-main Codebase Index

Welcome to the Open Knowledge Format (OKF) index for the `eShop-main` codebase. This mapping guides AI agents and developers in understanding the structural layout of the reference implementation.

## Navigation Graph

* **[Architecture Layers](architecture.md)**: Details the design patterns and segregation of responsibilities between presentation, core domain, and infrastructure code.
* **[Basket Flow Mapping](basket_flow.md)**: Explains the shopping cart/basket lifecycle and execution flow across ViewModels, Services, Entities, and UI Controllers/Pages.
* **[Build & Testing Guide](testing.md)**: Directs compilation commands and tells agents how to run regression tests against clean solution boundaries.

## Key Entry Points
* **Web UI**: `src/Web/` (ASP.NET Core Razor Pages & MVC)
* **API Endpoints**: `src/PublicApi/` (REST Endpoints for admin actions)
* **Database contexts**: `src/Infrastructure/` (EF Core CatalogContext & IdentityContext)
