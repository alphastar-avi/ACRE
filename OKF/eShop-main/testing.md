---
type: Developer Guide
title: Compilation, Build, and Testing Guidelines
description: Commands and solutions to execute builds and regression tests.
resource: CodeBase/eShop-main/Everything.sln
tags: [build, test, validation]
timestamp: 2026-06-29T00:00:00Z
---

# Compilation and Testing Guide

Guides AI agents and developers on building the repository and running regression tests cleanly.

## Build Setup
This repository contains multiple solution files. Always specify the target solution to prevent compilation failures.
* **Preferred Solution**: **`Everything.sln`** (Excludes docker-compose projects entirely, avoiding dependencies on active Docker containers).
* **Command**: `dotnet build Everything.sln`

## Test Execution
The test suites are written using `xUnit` and can be executed via the `dotnet test` CLI:
* **Run all tests (excluding docker-compose)**: `dotnet test Everything.sln`
* **Run Unit Tests only**: `dotnet test tests/UnitTests/UnitTests.csproj`
* **Run Functional Tests only**: `dotnet test tests/FunctionalTests/FunctionalTests.csproj`
* **Filter to specific test suite**: `dotnet test --filter "FullyQualifiedName~Basket"`
