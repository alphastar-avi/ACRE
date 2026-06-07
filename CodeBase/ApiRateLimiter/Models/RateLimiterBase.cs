using System;

namespace Models
{
    // abstract base for all rate limiting mode (since we have two now 1.counter and 2.timer)
    public abstract class RateLimiterBase
    {
        public string Name { get; protected set; }
        public int Limit { get; protected set; }
        protected int currentCount;

        // constructor,initilization of limiters with name and limit
        protected RateLimiterBase(string name, int limit)
        {
            Name = name;
            Limit = limit;
            currentCount = 0;
        }

        // abstract method, other class needs to use this
        public abstract bool AllowRequest();

        // for resetting current count
        public virtual void Reset()
        {
            currentCount = 0;
        }

        // 'get' shorthand for c# , to get the current counnt value
        public virtual int CurrentCount => currentCount;
    }
}