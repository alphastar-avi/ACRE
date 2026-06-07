namespace Models
{
    // counter mode for apiroute
    public class CounterLimiter : RateLimiterBase // inherit rateLimiterBase
    {   
        //Call parent constructor
        public CounterLimiter(string name, int limit) : base(name, limit)
        {
        }

        // override, counter update per request
        public override bool AllowRequest()
        {
            if (currentCount >= Limit)
                return false;

            currentCount++;
            return true;
        }
    }
}