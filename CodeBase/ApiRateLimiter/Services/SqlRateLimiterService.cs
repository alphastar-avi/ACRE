using System;
using System.Data;
using Microsoft.Data.SqlClient;
using Models;
using Data;

namespace Services
{
    public class SqlRateLimiterService : IRateLimiterService
    {
        private readonly DatabaseHelper _dbHelper;

        public SqlRateLimiterService()
        {
            _dbHelper = new DatabaseHelper();
        }

        // Add a counter route using stored procedure
        public bool AddRoute(string name, int limit)
        {
            try
            {
                string query = "EXEC dbo.sp_AddCounterRoute @Name, @Limit";

                var parameters = new SqlParameter[]
                {
                    new SqlParameter("@Name", name),
                    new SqlParameter("@Limit", limit)
                };

                var result = _dbHelper.ExecuteQuery(query, parameters);

                // Check Success column from stored procedure result
                if (result.Rows.Count > 0)
                {
                    int success = Convert.ToInt32(result.Rows[0]["Success"]);
                    return success == 1;
                }

                return false;
            }
            catch (SqlException)
            {
                return false;
            }
        }

        // Add a time-based route using stored procedure
        public bool AddTimeRoute(string name, int limit, int windowSeconds)
        {
            try
            {
                string query = "EXEC dbo.sp_AddTimeRoute @Name, @Limit, @WindowSeconds";

                var parameters = new SqlParameter[]
                {
                    new SqlParameter("@Name", name),
                    new SqlParameter("@Limit", limit),
                    new SqlParameter("@WindowSeconds", windowSeconds)
                };

                var result = _dbHelper.ExecuteQuery(query, parameters);

                // Check Success column from stored procedure result
                if (result.Rows.Count > 0)
                {
                    int success = Convert.ToInt32(result.Rows[0]["Success"]);
                    return success == 1;
                }

                return false;
            }
            catch (SqlException)
            {
                return false;
            }
        }

        // Try to make a request using stored procedure
        public bool TryRequest(string routeName)
        {
            try
            {
                string query = "EXEC dbo.sp_TryRequest @Name";

                var parameters = new SqlParameter[]
                {
                    new SqlParameter("@Name", routeName)
                };

                var result = _dbHelper.ExecuteQuery(query, parameters);

                // Check Allowed column from stored procedure result
                if (result.Rows.Count > 0)
                {
                    int allowed = Convert.ToInt32(result.Rows[0]["Allowed"]);
                    return allowed == 1;
                }

                return false;
            }
            catch (SqlException)
            {
                return false;
            }
        }

        // Get route details using stored procedure
        public RateLimiterBase? GetRoute(string routeName)
        {
            try
            {
                string query = "EXEC dbo.sp_GetRoute @Name";

                var parameters = new SqlParameter[]
                {
                    new SqlParameter("@Name", routeName)
                };

                var dataTable = _dbHelper.ExecuteQuery(query, parameters);

                if (dataTable.Rows.Count == 0)
                    return null;

                var row = dataTable.Rows[0];
                string limiterType = row["LimiterType"].ToString()!;
                string name = row["Name"].ToString()!;
                int limit = Convert.ToInt32(row["Limit"]);
                int currentCount = Convert.ToInt32(row["CurrentCount"]);

                if (limiterType == "Counter")
                {
                    var counterLimiter = new CounterLimiter(name, limit);
                    // Simulate current count by calling AllowRequest
                    for (int i = 0; i < currentCount; i++)
                    {
                        counterLimiter.AllowRequest();
                    }
                    return counterLimiter;
                }
                else if (limiterType == "Time")
                {
                    int windowSeconds = Convert.ToInt32(row["WindowSeconds"]);
                    var timeLimiter = new TimeLimiter(name, limit, windowSeconds);

                    // Simulate current count
                    for (int i = 0; i < currentCount; i++)
                    {
                        timeLimiter.AllowRequest();
                    }

                    return timeLimiter;
                }

                return null;
            }
            catch (SqlException)
            {
                return null;
            }
        }

        // Reset a route using stored procedure
        public bool ResetRoute(string routeName)
        {
            try
            {
                string query = "EXEC dbo.sp_ResetRoute @Name";

                var parameters = new SqlParameter[]
                {
                    new SqlParameter("@Name", routeName)
                };

                var result = _dbHelper.ExecuteQuery(query, parameters);

                // Check Success column from stored procedure result
                if (result.Rows.Count > 0)
                {
                    int success = Convert.ToInt32(result.Rows[0]["Success"]);
                    return success == 1;
                }

                return false;
            }
            catch (SqlException)
            {
                return false;
            }
        }
    }
}