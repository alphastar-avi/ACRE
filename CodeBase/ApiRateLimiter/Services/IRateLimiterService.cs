using Models;

namespace Services
{
    // Interface for just defining a contract ( exposing the capabilities )
    public interface IRateLimiterService
    {  
        //add a new API route with counter strategy
        bool AddRoute(string name, int limit);
        
        //add a new API route with time strategy
        bool AddTimeRoute(string name, int limit, int windowSeconds);
        
        //process a request for a route.
        bool TryRequest(string routeName);
        
        // get route detail
        RateLimiterBase? GetRoute(string routeName);
        
        // to reset the current count for a route to 0
        bool ResetRoute(string routeName);
    }
}
