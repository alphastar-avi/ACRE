using Services;

namespace Tests
{
    public static class RateLimiterRegressionTests
    {
        public static void RunAll()
        {
            ReadsMaxRequestsPerMinuteFromConfiguration();
            EnforcesConfiguredDefaultRouteLimit();
            LoadsRateLimitFromConfigurationFile();
            EnforcesRateLimitFromConfiguration();
            Console.WriteLine("Rate limiter regression tests passed.");
        }

        private static void LoadsRateLimitFromConfigurationFile()
        {
            var settings = RateLimitConfiguration.Load();

            Assert(settings.MaxRequestsPerMinute == 50,
                $"Expected MaxRequestsPerMinute to be 50 from appsettings.json, got {settings.MaxRequestsPerMinute}.");
            Assert(settings.WindowSeconds == 60,
                $"Expected WindowSeconds to be 60 from appsettings.json, got {settings.WindowSeconds}.");
        }

        private static void EnforcesRateLimitFromConfiguration()
        {
            var settings = RateLimitConfiguration.Load();
            var service = new RateLimiterService(settings);

            for (var i = 0; i < 50; i++)
            {
                Assert(
                    service.TryRequest(RateLimiterService.DefaultRouteName),
                    $"Expected request {i + 1} to be allowed.");
            }

            for (var i = 0; i < 10; i++)
            {
                Assert(
                    !service.TryRequest(RateLimiterService.DefaultRouteName),
                    $"Expected request after limit to be throttled.");
            }
        }

        private static void ReadsMaxRequestsPerMinuteFromConfiguration()
        {
            var path = Path.Combine(Path.GetTempPath(), $"rate-limit-{Guid.NewGuid():N}.json");
            File.WriteAllText(path, """
            {
              "RateLimit": {
                "MaxRequestsPerMinute": 50,
                "WindowSeconds": 60
              }
            }
            """);

            try
            {
                var settings = RateLimitConfiguration.Load(path);

                Assert(settings.MaxRequestsPerMinute == 50, "Expected MaxRequestsPerMinute to be read as 50.");
                Assert(settings.WindowSeconds == 60, "Expected WindowSeconds to be read as 60.");
            }
            finally
            {
                File.Delete(path);
            }
        }

        private static void EnforcesConfiguredDefaultRouteLimit()
        {
            var service = new RateLimiterService(new RateLimitSettings
            {
                MaxRequestsPerMinute = 50,
                WindowSeconds = 60
            });

            for (var i = 0; i < 50; i++)
            {
                Assert(
                    service.TryRequest(RateLimiterService.DefaultRouteName),
                    $"Expected request {i + 1} to be allowed.");
            }

            Assert(
                !service.TryRequest(RateLimiterService.DefaultRouteName),
                "Expected request 51 to be throttled.");
        }

        private static void Assert(bool condition, string message)
        {
            if (!condition)
                throw new InvalidOperationException(message);
        }
    }
}
