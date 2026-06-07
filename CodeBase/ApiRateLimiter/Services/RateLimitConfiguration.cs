using System.Text.Json;

namespace Services
{
    public sealed class RateLimitSettings
    {
        public const int DefaultMaxRequestsPerMinute = 100;
        public const int DefaultWindowSeconds = 60;

        public int MaxRequestsPerMinute { get; init; } = DefaultMaxRequestsPerMinute;
        public int WindowSeconds { get; init; } = DefaultWindowSeconds;
        public IReadOnlyList<ConfiguredRoute> Routes { get; init; } = Array.Empty<ConfiguredRoute>();
    }

    public sealed class ConfiguredRoute
    {
        public required string Name { get; init; }
        public int MaxRequestsPerMinute { get; init; }
        public int WindowSeconds { get; init; } = RateLimitSettings.DefaultWindowSeconds;
    }

    public static class RateLimitConfiguration
    {
        public static RateLimitSettings Load(string path = "appsettings.json")
        {
            var resolvedPath = ResolvePath(path);
            if (resolvedPath == null)
                return new RateLimitSettings();

            using var stream = File.OpenRead(resolvedPath);
            using var document = JsonDocument.Parse(stream);
            var root = document.RootElement;

            var rateLimitSection = TryGetObject(root, "RateLimit")
                ?? TryGetObject(root, "RateLimiter")
                ?? root;

            var maxRequests = ReadPositiveInt(
                rateLimitSection,
                RateLimitSettings.DefaultMaxRequestsPerMinute,
                "MaxRequestsPerMinute",
                "RequestsPerMinute",
                "Limit");

            var windowSeconds = ReadPositiveInt(
                rateLimitSection,
                RateLimitSettings.DefaultWindowSeconds,
                "WindowSeconds",
                "WindowInSeconds");

            return new RateLimitSettings
            {
                MaxRequestsPerMinute = maxRequests,
                WindowSeconds = windowSeconds,
                Routes = ReadRoutes(root, rateLimitSection, maxRequests, windowSeconds)
            };
        }

        private static string? ResolvePath(string path)
        {
            if (File.Exists(path))
                return path;

            var baseDirectoryPath = Path.Combine(AppContext.BaseDirectory, path);
            return File.Exists(baseDirectoryPath) ? baseDirectoryPath : null;
        }

        private static IReadOnlyList<ConfiguredRoute> ReadRoutes(
            JsonElement root,
            JsonElement rateLimitSection,
            int defaultMaxRequests,
            int defaultWindowSeconds)
        {
            var routesElement = TryGetArray(rateLimitSection, "Routes")
                ?? TryGetArray(root, "Routes");

            if (routesElement == null)
                return Array.Empty<ConfiguredRoute>();

            var routes = new List<ConfiguredRoute>();

            foreach (var routeElement in routesElement.Value.EnumerateArray())
            {
                if (routeElement.ValueKind != JsonValueKind.Object)
                    continue;

                var name = ReadString(routeElement, "Name", "Route", "RouteName", "Path");
                if (string.IsNullOrWhiteSpace(name))
                    continue;

                routes.Add(new ConfiguredRoute
                {
                    Name = name,
                    MaxRequestsPerMinute = ReadPositiveInt(
                        routeElement,
                        defaultMaxRequests,
                        "MaxRequestsPerMinute",
                        "RequestsPerMinute",
                        "Limit"),
                    WindowSeconds = ReadPositiveInt(
                        routeElement,
                        defaultWindowSeconds,
                        "WindowSeconds",
                        "WindowInSeconds")
                });
            }

            return routes;
        }

        private static JsonElement? TryGetObject(JsonElement element, string propertyName)
        {
            if (element.ValueKind == JsonValueKind.Object
                && TryGetProperty(element, propertyName, out var property)
                && property.ValueKind == JsonValueKind.Object)
            {
                return property;
            }

            return null;
        }

        private static JsonElement? TryGetArray(JsonElement element, string propertyName)
        {
            if (element.ValueKind == JsonValueKind.Object
                && TryGetProperty(element, propertyName, out var property)
                && property.ValueKind == JsonValueKind.Array)
            {
                return property;
            }

            return null;
        }

        private static int ReadPositiveInt(JsonElement element, int defaultValue, params string[] propertyNames)
        {
            foreach (var propertyName in propertyNames)
            {
                if (!TryGetProperty(element, propertyName, out var property))
                    continue;

                if (property.ValueKind == JsonValueKind.Number
                    && property.TryGetInt32(out var numericValue)
                    && numericValue > 0)
                {
                    return numericValue;
                }

                if (property.ValueKind == JsonValueKind.String
                    && int.TryParse(property.GetString(), out var stringValue)
                    && stringValue > 0)
                {
                    return stringValue;
                }
            }

            return defaultValue;
        }

        private static string? ReadString(JsonElement element, params string[] propertyNames)
        {
            foreach (var propertyName in propertyNames)
            {
                if (TryGetProperty(element, propertyName, out var property)
                    && property.ValueKind == JsonValueKind.String)
                {
                    return property.GetString();
                }
            }

            return null;
        }

        private static bool TryGetProperty(JsonElement element, string propertyName, out JsonElement property)
        {
            if (element.TryGetProperty(propertyName, out property))
                return true;

            if (element.ValueKind == JsonValueKind.Object)
            {
                foreach (var candidate in element.EnumerateObject())
                {
                    if (string.Equals(candidate.Name, propertyName, StringComparison.OrdinalIgnoreCase))
                    {
                        property = candidate.Value;
                        return true;
                    }
                }
            }

            property = default;
            return false;
        }
    }
}
