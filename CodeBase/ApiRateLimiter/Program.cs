using Services;

using ApiRateLimiter.Cli;
using Tests;

class Program
{
    static void Main(string[] args)
    {
        if (args.Contains("--run-tests"))
        {
            RateLimiterRegressionTests.RunAll();
            return;
        }

        //var rateLimiter = new RateLimiterService();
        var settings = RateLimitConfiguration.Load();
        IRateLimiterService rateLimiter = new RateLimiterService(settings); // DI using interface

        ApiRateLimiterCli.Start(rateLimiter);
    }
}
