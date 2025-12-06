<?php
/**
 * Lightspeed CMS (Velocity)
 *
 * Read-only client for the Velocity headless CMS
 * Supports APCu caching with stale-while-revalidate and ETag validation
 */

require_once __DIR__ . '/site.php';

class CMS {
    private const CACHE_PREFIX = 'lightspeed_cms_';
    private const DEFAULT_SOFT_TTL = 300;   // 5 minutes - serve stale, revalidate in background
    private const DEFAULT_HARD_TTL = 3600;  // 1 hour - absolute expiration

    private string $endpoint;
    private string $tenant;
    private int $softTtl;
    private int $hardTtl;
    private bool $cacheEnabled;

    /**
     * Create a new CMS client
     *
     * @param string|null $endpoint The Velocity API endpoint (defaults to velocity.ee)
     * @param string|null $tenant The tenant identifier (defaults to site name)
     * @param int|null $softTtl Soft TTL in seconds (triggers background revalidation)
     * @param int|null $hardTtl Hard TTL in seconds (absolute cache expiration)
     */
    public function __construct(
        ?string $endpoint = null,
        ?string $tenant = null,
        ?int $softTtl = null,
        ?int $hardTtl = null
    ) {
        $site = site();

        $this->endpoint = $endpoint
            ?? $site->get('velocity.endpoint')
            ?: 'https://velocity.ee';

        $this->tenant = $tenant
            ?? $site->get('velocity.tenant')
            ?: $site->name();

        $this->endpoint = rtrim($this->endpoint, '/');

        $this->softTtl = $softTtl ?? $site->getInt('velocity.cache.soft_ttl', self::DEFAULT_SOFT_TTL);
        $this->hardTtl = $hardTtl ?? $site->getInt('velocity.cache.hard_ttl', self::DEFAULT_HARD_TTL);

        // Check if APCu is available
        $apcuAvailable = function_exists('apcu_fetch') && apcu_enabled();

        // Cache enabled if: APCu available AND velocity.cache is true (default true)
        $this->cacheEnabled = $apcuAvailable && $site->getBoolean('velocity.cache', true);

        // Setup webhook on first run if caching is enabled
        if ($this->cacheEnabled) {
            $this->ensureWebhookSetup();
        }
    }

    /**
     * Get the configured endpoint
     */
    public function endpoint(): string {
        return $this->endpoint;
    }

    /**
     * Get the configured tenant
     */
    public function tenant(): string {
        return $this->tenant;
    }

    /**
     * Get content by type and id/key
     *
     * @param string $id The content id/key (name of the content)
     * @param string $type The content type (defaults to 'content')
     * @param string $contentType The expected content type (defaults to 'html')
     * @return string|null The content or null if not found
     */
    public function get(string $id, string $type = 'content', string $contentType = 'html'): ?string {
        $cacheKey = $this->cacheKey($type, $id, $contentType);

        // Try cache first
        if ($this->cacheEnabled) {
            $cached = apcu_fetch($cacheKey, $success);
            if ($success && $cached) {
                $isStale = time() > $cached['stale_at'];

                // If stale, trigger background revalidation
                if ($isStale) {
                    $this->revalidateInBackground($cacheKey, $cached, $type, $id, $contentType);
                }

                return $cached['content'];
            }
        }

        // Cache miss - fetch fresh
        $result = $this->fetchWithEtag($type, $id, $contentType);
        if ($result === null) {
            return null;
        }

        // Store in cache
        if ($this->cacheEnabled) {
            $this->storeInCache($cacheKey, $result['content'], $result['etag']);
        }

        return $result['content'];
    }

    /**
     * Get content with metadata (returns array with 'content', 'etag', 'lastModified', etc.)
     *
     * @param string $id The content id/key
     * @param string $type The content type (defaults to 'content')
     * @param string $contentType The expected content type (defaults to 'html')
     * @return array|null Content with metadata or null if not found
     */
    public function getWithMetadata(string $id, string $type = 'content', string $contentType = 'html'): ?array {
        $url = sprintf('%s/api/content/%s/%s', $this->endpoint, urlencode($type), urlencode($id));

        $headers = [
            'X-Tenant: ' . $this->tenant,
            'Accept: ' . $this->mimeType($contentType),
        ];

        $result = $this->fetchWithHeaders($url, $headers);
        if ($result === null) {
            return null;
        }

        return [
            'content' => $result['body'],
            'etag' => $result['headers']['etag'] ?? null,
            'lastModified' => $result['headers']['last-modified'] ?? null,
            'versionId' => $result['headers']['x-version-id'] ?? null,
            'contentType' => $result['headers']['content-type'] ?? null,
        ];
    }

    /**
     * List all content items of a given type
     *
     * @param string $type The content type (defaults to 'content')
     * @return array List of content items with id, last_modified, and size
     */
    public function list(string $type = 'content'): array {
        $url = sprintf('%s/api/content/%s', $this->endpoint, urlencode($type));

        $headers = [
            'X-Tenant: ' . $this->tenant,
            'Accept: application/json',
        ];

        $response = $this->fetch($url, $headers);
        if ($response === false) {
            return [];
        }

        $data = json_decode($response, true);
        return $data['items'] ?? [];
    }

    /**
     * Get all available content types
     *
     * @return array List of content type names
     */
    public function types(): array {
        $url = sprintf('%s/api/types', $this->endpoint);

        $headers = [
            'X-Tenant: ' . $this->tenant,
            'Accept: application/json',
        ];

        $response = $this->fetch($url, $headers);
        if ($response === false) {
            return [];
        }

        $data = json_decode($response, true);
        return $data['types'] ?? [];
    }

    /**
     * Get all versions of a content item
     *
     * @param string $id The content id/key
     * @param string $type The content type (defaults to 'content')
     * @return array List of versions
     */
    public function versions(string $id, string $type = 'content'): array {
        $url = sprintf('%s/api/content/%s/%s/versions', $this->endpoint, urlencode($type), urlencode($id));

        $headers = [
            'X-Tenant: ' . $this->tenant,
            'Accept: application/json',
        ];

        $response = $this->fetch($url, $headers);
        if ($response === false) {
            return [];
        }

        $data = json_decode($response, true);
        return $data['versions'] ?? [];
    }

    /**
     * Get a specific version of content
     *
     * @param string $id The content id/key
     * @param string $versionId The version id
     * @param string $type The content type (defaults to 'content')
     * @param string $contentType The expected content type (defaults to 'html')
     * @return string|null The content or null if not found
     */
    public function version(string $id, string $versionId, string $type = 'content', string $contentType = 'html'): ?string {
        $url = sprintf(
            '%s/api/content/%s/%s/versions/%s',
            $this->endpoint,
            urlencode($type),
            urlencode($id),
            urlencode($versionId)
        );

        $headers = [
            'X-Tenant: ' . $this->tenant,
            'Accept: ' . $this->mimeType($contentType),
        ];

        $content = $this->fetch($url, $headers);
        return $content !== false ? $content : null;
    }

    /**
     * Check if content exists (uses conditional request)
     *
     * @param string $id The content id/key
     * @param string $type The content type (defaults to 'content')
     * @return bool True if content exists
     */
    public function exists(string $id, string $type = 'content'): bool {
        $url = sprintf('%s/api/content/%s/%s', $this->endpoint, urlencode($type), urlencode($id));

        $headers = [
            'X-Tenant: ' . $this->tenant,
        ];

        $context = stream_context_create([
            'http' => [
                'method' => 'HEAD',
                'header' => implode("\r\n", $headers),
                'ignore_errors' => true,
            ],
        ]);

        $response = @file_get_contents($url, false, $context);
        if (isset($http_response_header)) {
            $statusLine = $http_response_header[0] ?? '';
            if (preg_match('/HTTP\/[\d.]+\s+(\d+)/', $statusLine, $matches)) {
                return (int)$matches[1] === 200;
            }
        }

        return false;
    }

    /**
     * Convert content type shorthand to MIME type
     */
    private function mimeType(string $contentType): string {
        return match (strtolower($contentType)) {
            'html' => 'text/html',
            'json' => 'application/json',
            'xml' => 'application/xml',
            'text', 'txt' => 'text/plain',
            'php' => 'text/php',
            'png' => 'image/png',
            'jpg', 'jpeg' => 'image/jpeg',
            'gif' => 'image/gif',
            'webp' => 'image/webp',
            'svg' => 'image/svg+xml',
            'pdf' => 'application/pdf',
            default => $contentType,
        };
    }

    /**
     * Fetch content from URL
     */
    private function fetch(string $url, array $headers): string|false {
        $context = stream_context_create([
            'http' => [
                'method' => 'GET',
                'header' => implode("\r\n", $headers),
                'ignore_errors' => true,
            ],
        ]);

        $response = @file_get_contents($url, false, $context);
        if ($response === false) {
            return false;
        }

        if (isset($http_response_header)) {
            $statusLine = $http_response_header[0] ?? '';
            if (preg_match('/HTTP\/[\d.]+\s+(\d+)/', $statusLine, $matches)) {
                $statusCode = (int)$matches[1];
                if ($statusCode >= 400) {
                    return false;
                }
            }
        }

        return $response;
    }

    /**
     * Fetch content with response headers
     */
    private function fetchWithHeaders(string $url, array $headers): ?array {
        $context = stream_context_create([
            'http' => [
                'method' => 'GET',
                'header' => implode("\r\n", $headers),
                'ignore_errors' => true,
            ],
        ]);

        $response = @file_get_contents($url, false, $context);
        if ($response === false) {
            return null;
        }

        $responseHeaders = [];
        if (isset($http_response_header)) {
            $statusLine = $http_response_header[0] ?? '';
            if (preg_match('/HTTP\/[\d.]+\s+(\d+)/', $statusLine, $matches)) {
                $statusCode = (int)$matches[1];
                if ($statusCode >= 400) {
                    return null;
                }
            }

            foreach ($http_response_header as $header) {
                if (str_contains($header, ':')) {
                    [$name, $value] = explode(':', $header, 2);
                    $responseHeaders[strtolower(trim($name))] = trim($value);
                }
            }
        }

        return [
            'body' => $response,
            'headers' => $responseHeaders,
        ];
    }

    /**
     * Generate cache key for content
     */
    private function cacheKey(string $type, string $id, string $contentType): string {
        return self::CACHE_PREFIX . md5($this->endpoint . ':' . $this->tenant . ':' . $type . ':' . $id . ':' . $contentType);
    }

    /**
     * Fetch content with ETag support
     */
    private function fetchWithEtag(string $type, string $id, string $contentType, ?string $etag = null): ?array {
        $url = sprintf('%s/api/content/%s/%s', $this->endpoint, urlencode($type), urlencode($id));

        $headers = [
            'X-Tenant: ' . $this->tenant,
            'Accept: ' . $this->mimeType($contentType),
        ];

        if ($etag !== null) {
            $headers[] = 'If-None-Match: ' . $etag;
        }

        $context = stream_context_create([
            'http' => [
                'method' => 'GET',
                'header' => implode("\r\n", $headers),
                'ignore_errors' => true,
            ],
        ]);

        $response = @file_get_contents($url, false, $context);

        $statusCode = 0;
        $responseEtag = null;

        if (isset($http_response_header)) {
            $statusLine = $http_response_header[0] ?? '';
            if (preg_match('/HTTP\/[\d.]+\s+(\d+)/', $statusLine, $matches)) {
                $statusCode = (int)$matches[1];
            }

            foreach ($http_response_header as $header) {
                if (stripos($header, 'etag:') === 0) {
                    $responseEtag = trim(substr($header, 5));
                    break;
                }
            }
        }

        // 304 Not Modified
        if ($statusCode === 304) {
            return ['status' => 304, 'content' => null, 'etag' => $etag];
        }

        // Error
        if ($statusCode >= 400 || $response === false) {
            return null;
        }

        return [
            'status' => $statusCode,
            'content' => $response,
            'etag' => $responseEtag,
        ];
    }

    /**
     * Store content in cache
     */
    private function storeInCache(string $cacheKey, string $content, ?string $etag): void {
        $data = [
            'content' => $content,
            'etag' => $etag,
            'cached_at' => time(),
            'stale_at' => time() + $this->softTtl,
        ];

        apcu_store($cacheKey, $data, $this->hardTtl);
    }

    /**
     * Revalidate content in background using ETag
     * Uses a lock to prevent multiple simultaneous revalidations
     */
    private function revalidateInBackground(string $cacheKey, array $cached, string $type, string $id, string $contentType): void {
        $lockKey = $cacheKey . '_lock';

        // Try to acquire lock (expires in 30 seconds)
        if (!apcu_add($lockKey, true, 30)) {
            return; // Another process is already revalidating
        }

        // Perform conditional request with ETag
        $result = $this->fetchWithEtag($type, $id, $contentType, $cached['etag']);

        if ($result === null) {
            // Fetch failed, keep stale content but extend TTL slightly
            $cached['stale_at'] = time() + 60; // Retry in 1 minute
            apcu_store($cacheKey, $cached, $this->hardTtl);
        } elseif ($result['status'] === 304) {
            // Not modified, refresh TTLs
            $cached['stale_at'] = time() + $this->softTtl;
            apcu_store($cacheKey, $cached, $this->hardTtl);
        } else {
            // Content changed, update cache
            $this->storeInCache($cacheKey, $result['content'], $result['etag']);
        }

        apcu_delete($lockKey);
    }

    /**
     * Clear all cached content for this tenant
     */
    public function clearCache(): void {
        if (!$this->cacheEnabled || !class_exists('APCUIterator')) {
            return;
        }

        $prefix = self::CACHE_PREFIX;
        foreach (new \APCUIterator('/^' . preg_quote($prefix, '/') . '/') as $item) {
            apcu_delete($item['key']);
        }
    }

    /**
     * Clear cached content for a specific item
     */
    public function clearCacheFor(string $id, string $type = 'content', string $contentType = 'html'): void {
        if (!$this->cacheEnabled) {
            return;
        }

        $cacheKey = $this->cacheKey($type, $id, $contentType);
        apcu_delete($cacheKey);
    }

    /**
     * Clear cached content for an item across all content types
     */
    public function clearCacheForAll(string $id, string $type = 'content'): void {
        if (!$this->cacheEnabled || !class_exists('APCUIterator')) {
            return;
        }

        // Match cache keys for this endpoint/tenant/type/id with any content type
        $prefix = self::CACHE_PREFIX . md5($this->endpoint . ':' . $this->tenant . ':' . $type . ':' . $id . ':');

        // Since we hash the full key, we need to try common content types
        $contentTypes = ['html', 'json', 'xml', 'text', 'txt', 'php', 'png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'pdf'];
        foreach ($contentTypes as $contentType) {
            $cacheKey = $this->cacheKey($type, $id, $contentType);
            apcu_delete($cacheKey);
        }
    }

    /**
     * Ensure webhook handler file exists and webhook is registered
     */
    private function ensureWebhookSetup(): void {
        $setupKey = self::CACHE_PREFIX . 'webhook';

        // Check if already setup in this cache lifetime
        if (apcu_fetch($setupKey)) {
            return;
        }

        // Use /tmp/lightspeed/velocity for generated files (nginx routes /_/ to /tmp/lightspeed/)
        $velocityDir = '/tmp/lightspeed/velocity';
        if (!is_dir($velocityDir)) {
            @mkdir($velocityDir, 0777, true);
        }

        // Write cache handler if it doesn't exist
        $cacheFile = $velocityDir . '/cache.php';
        if (!file_exists($cacheFile)) {
            $this->writeCacheHandler($cacheFile);
        }

        // TODO: Register webhook with Velocity when API is available
        // $this->registerWebhook();

        // Mark as setup (cache for 24 hours)
        apcu_store($setupKey, true, 86400);
    }

    /**
     * Write the cache management PHP file
     */
    private function writeCacheHandler(string $path): void {
        $code = <<<'PHP'
<?php
/**
 * Velocity CMS Cache Handler
 * Auto-generated by Lightspeed CMS client
 */

require_once 'lightspeed/cms.php';

header('Content-Type: application/json');

$operation = $_GET['operation'] ?? 'list';
$type = $_GET['type'] ?? null;
$id = $_GET['id'] ?? null;

if ($operation === 'clear') {
    if ($id) {
        cms()->clearCacheForAll($id, $type ?? 'content');
        echo json_encode(['status' => 'ok', 'operation' => 'clear', 'cleared' => ['type' => $type ?? 'content', 'id' => $id]]);
    } else {
        cms()->clearCache();
        echo json_encode(['status' => 'ok', 'operation' => 'clear', 'cleared' => 'all']);
    }
    exit;
}

// List cache contents
$entries = [];
$prefix = 'lightspeed_cms_';

if (function_exists('apcu_cache_info') && class_exists('APCUIterator')) {
    foreach (new APCUIterator('/^' . preg_quote($prefix, '/') . '/') as $item) {
        $data = $item['value'];
        $entries[] = [
            'key' => $item['key'],
            'cached_at' => date('c', $data['cached_at'] ?? 0),
            'stale_at' => date('c', $data['stale_at'] ?? 0),
            'is_stale' => time() > ($data['stale_at'] ?? 0),
            'etag' => $data['etag'] ?? null,
            'content_length' => strlen($data['content'] ?? ''),
            'ttl' => $item['ttl'],
            'hits' => $item['num_hits'],
        ];
    }
}

echo json_encode([
    'status' => 'ok',
    'enabled' => function_exists('apcu_fetch') && apcu_enabled(),
    'entries' => $entries,
    'count' => count($entries),
]);
PHP;

        file_put_contents($path, $code);
    }

    /**
     * Register webhook with Velocity
     */
    private function registerWebhook(): void {
        $site = site();

        $domain = $site->domain();
        if ($domain === '') {
            return; // Can't register without a domain
        }

        $protocol = $site->getBoolean('ssl', true) ? 'https' : 'http';
        $webhookUrl = $protocol . '://' . $domain . '/_/velocity/cache.php?operation=clear';

        $url = $this->endpoint . '/api/webhooks';
        $payload = json_encode([
            'tenant' => $this->tenant,
            'url' => $webhookUrl,
            'events' => ['content.published', 'content.deleted'],
        ]);

        $context = stream_context_create([
            'http' => [
                'method' => 'POST',
                'header' => "Content-Type: application/json\r\nX-Tenant: " . $this->tenant,
                'content' => $payload,
                'ignore_errors' => true,
            ],
        ]);

        @file_get_contents($url, false, $context);
    }
}

/**
 * Helper function to get a CMS instance with default configuration
 */
function cms(?string $endpoint = null, ?string $tenant = null): CMS {
    static $instance = null;
    if ($instance === null || $endpoint !== null || $tenant !== null) {
        $instance = new CMS($endpoint, $tenant);
    }
    return $instance;
}

/**
 * Helper function to get content directly
 */
function content(string $id, string $type = 'content', string $contentType = 'html'): ?string {
    return cms()->get($id, $type, $contentType);
}
