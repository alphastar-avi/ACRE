using System.Collections.Generic;
using Models;

namespace Services{

    public class RateLimiterService : IRateLimiterService
    {   
        //map to hold the relation - now stores RateLimiterBase instead of ApiRoute
        private readonly Dictionary<string, RateLimiterBase> _routes;
        public const string DefaultRouteName = "default";

        // Default constructor for _routes as dictonary
        public RateLimiterService() : this(new RateLimitSettings())
        {
        }

        public RateLimiterService(RateLimitSettings settings)
        {
            _routes = new Dictionary<string, RateLimiterBase>();

            AddTimeRoute(DefaultRouteName, settings.MaxRequestsPerMinute, settings.WindowSeconds);

            foreach (var route in settings.Routes)
            {
                AddTimeRoute(route.Name, route.MaxRequestsPerMinute, route.WindowSeconds);
            }
        }

        // To add a new route with counter strategy
        public bool AddRoute(string name, int limit)
        {
            if (_routes.ContainsKey(name)) return false; // route already present

            _routes[name] = new CounterLimiter(name, limit); // create new counter limiter
            return true;
        }

        // To add a new route with time strategy
        public bool AddTimeRoute(string name, int limit, int windowSeconds)
        {
            if (_routes.ContainsKey(name)) return false; // route already present

            _routes[name] = new TimeLimiter(name, limit, windowSeconds); // create new time limiter
            return true;
        }

        // Accessing route on request - polymorphism in action
        public bool TryRequest(string routeName)
        {
            if (!_routes.ContainsKey(routeName)) return false; // the route does not exist so return false 

            return _routes[routeName].AllowRequest(); // calls the correct override method
        }

        // get the route if it present else give null
        public RateLimiterBase? GetRoute(string routeName)
        {
            if (_routes.ContainsKey(routeName)) return _routes[routeName];

            //if not present return null (the '?' allows null)
            return null;
        }

        public bool ResetRoute(string routeName)
        {
            //if (_routes.ContainsKey(routeName))
            //{
            //    _routes[routeName].Reset();
            //    return true;
            //} Double lookup ... try get value is better i guess

            if (!_routes.TryGetValue(routeName, out var route))
                return false;

            route.Reset();   // resets CurrentCount of that route - calls virtual method
            return true;
        }

    }

}
