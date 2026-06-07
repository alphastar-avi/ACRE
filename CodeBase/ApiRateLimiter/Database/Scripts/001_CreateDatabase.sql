
-- Script: Create RateLimiterDB Database
-- Description: Creates the main database for the Rate Limiter application
-- Author: Avinash 
-- Date: 2026-02-10

USE master;
GO

-- Drop database if it exists (for clean reinstall)
IF EXISTS (SELECT name FROM sys.databases WHERE name = 'RateLimiterDB')
BEGIN
    ALTER DATABASE RateLimiterDB SET SINGLE_USER WITH ROLLBACK IMMEDIATE;
    DROP DATABASE RateLimiterDB;
END
GO

-- Create the database
CREATE DATABASE RateLimiterDB;
GO

PRINT 'Database RateLimiterDB created successfully';
GO