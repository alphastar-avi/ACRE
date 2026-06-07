
-- Master Script: Run All Database Scripts
-- Description: Executes all scripts in order
-- Usage: Run this file to set up the entire database


PRINT 'Starting Database Setup';
GO

-- Script 1: Create Database
:r .\001_CreateDatabase.sql
GO

-- Script 2: Create Tables
:r .\002_CreateRoutesTable.sql
GO

-- Script 3: Create Stored Procedures
:r .\003_CreateStoredProcedures.sql
GO


PRINT 'Database Setup Complete!';
GO