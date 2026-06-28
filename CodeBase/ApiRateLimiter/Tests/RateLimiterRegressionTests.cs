using Services;
using ApiRateLimiter.Cli;

namespace Tests
{
    public static class RateLimiterRegressionTests
    {
        public static void RunAll()
        {
            ReadsMaxRequestsPerMinuteFromConfiguration();
            EnforcesConfiguredDefaultRouteLimit();
            CliReturnsToPromptWhenCounterLimitIsEmpty();
            Console.WriteLine("Rate limiter regression tests passed.");
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

        private static void CliReturnsToPromptWhenCounterLimitIsEmpty()
        {
            var originalIn = Console.In;
            var originalOut = Console.Out;
            var output = new StringWriter();

            try
            {
                Console.SetIn(new StringReader("add\nroute-a\n\nexit\n"));
                Console.SetOut(output);

                ApiRateLimiterCli.Start(new RateLimiterService());
            }
            finally
            {
                Console.SetIn(originalIn);
                Console.SetOut(originalOut);
            }

            var text = output.ToString();

            Assert(
                text.Contains("Error: Limit is required"),
                "Expected empty limit input to display an error.");
            Assert(
                text.Contains("Exiting application..."),
                "Expected CLI to return to the main prompt after empty limit input.");
        }

        private static void Assert(bool condition, string message)
        {
            if (!condition)
                throw new InvalidOperationException(message);
        }
    }
}
