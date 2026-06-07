
-- Script: Create Routes Table
-- Description: Creates the Routes table with all constraints
-- Author: Avinash
-- Date: 2026-02-11


USE RateLimiterDB;
GO

-- Drop table if exists (for clean reinstall)
IF OBJECT_ID('dbo.Routes', 'U') IS NOT NULL
    DROP TABLE dbo.Routes;
GO

-- Create Routes table
CREATE TABLE dbo.Routes (
    Id INT PRIMARY KEY IDENTITY(1,1),
    Name NVARCHAR(100) NOT NULL UNIQUE,
    Limit INT NOT NULL CHECK (Limit > 0),
    CurrentCount INT NOT NULL DEFAULT 0 CHECK (CurrentCount >= 0),
    LimiterType NVARCHAR(20) NOT NULL CHECK (LimiterType IN ('Counter', 'Time')),
    WindowSeconds INT NULL CHECK (WindowSeconds > 0 OR WindowSeconds IS NULL),
    WindowStart DATETIME NULL,
    CreatedAt DATETIME NOT NULL DEFAULT GETDATE(),
    UpdatedAt DATETIME NULL
);
GO

-- Create index on Name for faster lookups
CREATE INDEX IX_Routes_Name ON dbo.Routes(Name);
GO

-- Create index on LimiterType for filtering
CREATE INDEX IX_Routes_LimiterType ON dbo.Routes(LimiterType);
GO

PRINT 'Routes table created successfully with indexes';
GO