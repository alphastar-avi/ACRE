using System;

namespace Models
{
    // time mode for api route
    public class TimeLimiter : RateLimiterBase // inherit parent
    {
        private DateTime windowStart;
        private int windowSeconds;

        public TimeLimiter(string name, int limit, int windowSeconds) : base(name, limit)
        {
            this.windowSeconds = windowSeconds;  // Store window duration
            windowStart = DateTime.Now; // save the time of created 
        }

        // check if window expired and reset if needed
        private void CheckAndResetWindow()
        {
            DateTime now = DateTime.Now;
            
            if ((now - windowStart).TotalSeconds >= windowSeconds) //converts that TimeSpan to seconds and check with the windowseconds
            {
                Reset();
                windowStart = now;
            }
        }

        // allow request for time mode
        public override bool AllowRequest()
        {
            CheckAndResetWindow();

            // check limit within current window
            if (currentCount >= Limit)
                return false;

            currentCount++;
            return true;
        }

        // override current count to check window first
        public override int CurrentCount 
        { 
            get 
            {
                CheckAndResetWindow();
                return currentCount;
            }
        }

        // show window start and seconds ( get )
        public int WindowSeconds => windowSeconds;
        public DateTime WindowStart => windowStart;
    }
}