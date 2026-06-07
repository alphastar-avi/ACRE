-- Script: Create Stored Procedures
-- Description: Creates all stored procedures for Rate Limiter
-- Author: Your Name
-- Date: 2026-02-10


USE RateLimiterDB;
GO


-- Stored Procedure: AddCounterRoute
-- Description: Adds a new counter-based route

CREATE OR ALTER PROCEDURE dbo.sp_AddCounterRoute
    @Name NVARCHAR(100),
    @Limit INT
AS
BEGIN
    SET NOCOUNT ON;

    BEGIN TRY
        INSERT INTO dbo.Routes (Name, Limit, CurrentCount, LimiterType, WindowSeconds, WindowStart)
        VALUES (@Name, @Limit, 0, 'Counter', NULL, NULL);

        SELECT 1 AS Success, 'Route added successfully' AS Message;
    END TRY
    BEGIN CATCH
        IF ERROR_NUMBER() = 2627 -- Unique constraint violation
            SELECT 0 AS Success, 'Route already exists' AS Message;
        ELSE
            SELECT 0 AS Success, ERROR_MESSAGE() AS Message;
    END CATCH
END
GO


-- Stored Procedure: AddTimeRoute
-- Description: Adds a new time-based route

CREATE OR ALTER PROCEDURE dbo.sp_AddTimeRoute
    @Name NVARCHAR(100),
    @Limit INT,
    @WindowSeconds INT
AS
BEGIN
    SET NOCOUNT ON;

    BEGIN TRY
        INSERT INTO dbo.Routes (Name, Limit, CurrentCount, LimiterType, WindowSeconds, WindowStart)
        VALUES (@Name, @Limit, 0, 'Time', @WindowSeconds, GETDATE());

        SELECT 1 AS Success, 'Time route added successfully' AS Message;
    END TRY
    BEGIN CATCH
        IF ERROR_NUMBER() = 2627
            SELECT 0 AS Success, 'Route already exists' AS Message;
        ELSE
            SELECT 0 AS Success, ERROR_MESSAGE() AS Message;
    END CATCH
END
GO



-- Stored Procedure: GetRoute
-- Description: Gets route details by name


CREATE OR ALTER PROCEDURE dbo.sp_GetRoute
    @Name NVARCHAR(100)
AS
BEGIN
    SET NOCOUNT ON;

    SELECT 
        Id,
        Name,
        Limit,
        CurrentCount,
        LimiterType,
        WindowSeconds,
        WindowStart,
        CreatedAt,
        UpdatedAt
    FROM dbo.Routes
    WHERE Name = @Name;
END
GO



-- Stored Procedure: TryRequest
-- Description: Attempts to process a request
-- Returns: 1 if allowed, 0 if denied

CREATE OR ALTER PROCEDURE dbo.sp_TryRequest
    @Name NVARCHAR(100)
AS
BEGIN
    SET NOCOUNT ON;

    DECLARE @CurrentCount INT;
    DECLARE @Limit INT;
    DECLARE @LimiterType NVARCHAR(20);
    DECLARE @WindowSeconds INT;
    DECLARE @WindowStart DATETIME;

    -- Get route details
    SELECT 
        @CurrentCount = CurrentCount,
        @Limit = Limit,
        @LimiterType = LimiterType,
        @WindowSeconds = WindowSeconds,
        @WindowStart = WindowStart
    FROM dbo.Routes
    WHERE Name = @Name;

    -- Route not found
    IF @@ROWCOUNT = 0
    BEGIN
        SELECT 0 AS Allowed, 'Route not found' AS Message;
        RETURN;
    END

    -- Handle Time-based limiter
    IF @LimiterType = 'Time'
    BEGIN
        -- Check if window expired
        IF DATEDIFF(SECOND, @WindowStart, GETDATE()) >= @WindowSeconds
        BEGIN
            -- Reset window
            UPDATE dbo.Routes
            SET CurrentCount = 1, WindowStart = GETDATE(), UpdatedAt = GETDATE()
            WHERE Name = @Name;

            SELECT 1 AS Allowed, 'Request allowed (window reset)' AS Message;
            RETURN;
        END
    END

    -- Check limit
    IF @CurrentCount >= @Limit
    BEGIN
        SELECT 0 AS Allowed, 'Rate limit exceeded' AS Message;
        RETURN;
    END

    -- Increment counter
    UPDATE dbo.Routes
    SET CurrentCount = CurrentCount + 1, UpdatedAt = GETDATE()
    WHERE Name = @Name;

    SELECT 1 AS Allowed, 'Request allowed' AS Message;
END
GO



-- Stored Procedure: ResetRoute
-- Description: Resets a route's counter

CREATE OR ALTER PROCEDURE dbo.sp_ResetRoute
    @Name NVARCHAR(100)
AS
BEGIN
    SET NOCOUNT ON;

    UPDATE dbo.Routes
    SET CurrentCount = 0, WindowStart = GETDATE(), UpdatedAt = GETDATE()
    WHERE Name = @Name;

    IF @@ROWCOUNT > 0
        SELECT 1 AS Success, 'Route reset successfully' AS Message;
    ELSE
        SELECT 0 AS Success, 'Route not found' AS Message;
END
GO



-- Stored Procedure: GetAllRoutes
-- Description: Gets all routes

CREATE OR ALTER PROCEDURE dbo.sp_GetAllRoutes
AS
BEGIN
    SET NOCOUNT ON;

    SELECT 
        Id,
        Name,
        Limit,
        CurrentCount,
        LimiterType,
        WindowSeconds,
        WindowStart,
        CreatedAt,
        UpdatedAt
    FROM dbo.Routes
    ORDER BY CreatedAt DESC;
END
GO



-- Stored Procedure: DeleteRoute
-- Description: Deletes a route

CREATE OR ALTER PROCEDURE dbo.sp_DeleteRoute
    @Name NVARCHAR(100)
AS
BEGIN
    SET NOCOUNT ON;

    DELETE FROM dbo.Routes
    WHERE Name = @Name;

    IF @@ROWCOUNT > 0
        SELECT 1 AS Success, 'Route deleted successfully' AS Message;
    ELSE
        SELECT 0 AS Success, 'Route not found' AS Message;
END
GO

PRINT 'All stored procedures created successfully';
GO