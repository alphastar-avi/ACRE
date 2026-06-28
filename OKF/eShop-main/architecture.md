---
type: Architecture Design
title: Codebase Layers and Design Patterns
description: Description of the presentation, application core, and infrastructure layers.
resource: CodeBase/eShop-main
tags: [architecture, design, patterns]
timestamp: 2026-06-29T00:00:00Z
---

# Architecture Layers

This project uses a clean architecture structure separating domain models and logic from infrastructure dependencies and delivery channels.

## Core Layers

### 1. Presentation Layer (`src/Web/`)
This project handles the user interface and coordinates interactions.
* Reusable view components live in `src/Web/Pages/Shared/Components/`.
* Mappings between Domain entities and ViewModels are performed in Services (e.g. `src/Web/Services/BasketViewModelService.cs`).

### 2. Core Domain Layer (`src/ApplicationCore/`)
This is the heart of the application. It contains all business entities, core interfaces, specifications, and domain services.
* It must not reference any database, framework, or third-party persistence SDKs.
* All persistence operations must be abstracted behind interfaces (e.g., `IRepository<T>`, `IReadRepository<T>`).

### 3. Infrastructure Layer (`src/Infrastructure/`)
This layer implements all core interfaces by referencing Entity Framework Core, SQL Server, and other external dependencies.
* Concrete repository classes reside under `src/Infrastructure/Data/`.
* Entity database configurations live under `src/Infrastructure/Data/Config/`.
