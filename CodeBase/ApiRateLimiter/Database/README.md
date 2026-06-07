# Database Setup Instructions

## Prerequisites
- SQL Server LocalDB installed
- SQL Server Management Studio (SSMS)

## Setup Steps

### Option 1: Run Individual Scripts
1. Open SSMS
2. Connect to `(localdb)\MSSQLLocalDB`
3. Run scripts in order:
   - `001_CreateDatabase.sql`
   - `002_CreateRoutesTable.sql`
   - `003_CreateStoredProcedures.sql`

### Option 2: Run Master Script
1. Open SSMS
2. Connect to `(localdb)\MSSQLLocalDB`
3. Open `000_RunAll.sql`
4. Enable SQLCMD Mode: Query ? SQLCMD Mode
5. Execute

