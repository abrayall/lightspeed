<?php
/**
 * Lightspeed CMS (Velocity)
 *
 * Read-only client for the Velocity headless CMS
 * Supports APCu caching with stale-while-revalidate and ETag validation
 */

require_once __DIR__ . '/site.php';

class CMS {
    private const CACHE_PREFIX = 'velocity/cms/';
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
        $cacheKey = $this->cacheKey($type, $id, 'unknown', 'content');

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
     * Get a URL for content (instead of the content itself)
     * Useful for images, CSS, and other assets that can be loaded via src/href
     *
     * @param string $id The content id/key
     * @param string $type The content type (defaults to 'content')
     * @return string|null The URL or null if not found
     */
    public function getUrl(string $id, string $type = 'content'): ?string {
        $result = $this->getAll([
            ['type' => $type, 'id' => $id, 'attribute' => 'url']
        ]);

        $key = "$type/$id";
        return $result[$key]['url'] ?? null;
    }

    /**
     * Get metadata for content (alt text, dimensions, etc.)
     *
     * @param string $id The content id/key
     * @param string $type The content type (defaults to 'content')
     * @return array|null The metadata or null if not found
     */
    public function getMetadata(string $id, string $type = 'content'): ?array {
        $result = $this->getAll([
            ['type' => $type, 'id' => $id, 'attribute' => 'metadata']
        ]);

        $key = "$type/$id";
        return $result[$key]['metadata'] ?? null;
    }

    /**
     * Get URL and metadata for content in one request
     *
     * @param string $id The content id/key
     * @param string $type The content type (defaults to 'content')
     * @return array|null Array with 'url' and 'metadata' keys, or null if not found
     */
    public function getAsset(string $id, string $type = 'content'): ?array {
        $result = $this->getAll([
            ['type' => $type, 'id' => $id, 'attributes' => ['url', 'metadata']]
        ]);

        $key = "$type/$id";
        return $result[$key] ?? null;
    }

    /**
     * Get multiple content items in a single request
     *
     * @param array $items Array of items to fetch, each with 'type', 'id', and optional 'attribute'/'attributes'/'content-type' keys
     *                     Example: [
     *                         ['type' => 'pages', 'id' => 'bio'],
     *                         ['type' => 'images', 'id' => 'hero', 'attributes' => ['url', 'metadata']],
     *                         ['type' => 'images', 'id' => 'logo', 'attribute' => 'url', 'content-type' => 'image/png'],
     *                     ]
     *                     Use attribute: 'url' or 'metadata' for single attribute
     *                     Use attributes: ['url', 'metadata'] for multiple in one request
     *                     Use content-type for MIME type hints (e.g., 'image/png')
     *                     Results are merged per type/id
     * @return array Associative array keyed by "type/id" with content or error info
     */
    public function getAll(array $items): array {
        if (empty($items)) {
            return [];
        }

        $results = [];
        $toFetch = [];

        // Check cache and determine what needs fetching
        foreach ($items as $item) {
            $type = $item['type'] ?? 'content';
            $id = $item['id'] ?? '';
            $key = "$type/$id";

            // Support both 'attribute' (single) and 'attributes' (array)
            $requestedAttrs = $item['attributes'] ?? [$item['attribute'] ?? 'content'];
            if (!is_array($requestedAttrs)) {
                $requestedAttrs = [$requestedAttrs];
            }

            $missingAttrs = [];
            $cachedData = ['type' => $type, 'id' => $id];
            $anyStale = false;
            $contentType = $item['content-type'] ?? 'unknown';

            if ($this->cacheEnabled) {
                // Check cache for each requested attribute
                foreach ($requestedAttrs as $attr) {
                    $cacheKey = $this->cacheKey($type, $id, $contentType, $attr);
                    $cached = apcu_fetch($cacheKey, $success);

                    if ($success && $cached) {
                        $isStale = time() > ($cached['stale_at'] ?? 0);
                        if ($isStale) {
                            $anyStale = true;
                        }
                        $cachedData[$attr] = $cached['content'];
                    } else {
                        $missingAttrs[] = $attr;
                    }
                }

                if (count($cachedData) > 2) {
                    $cachedData['cached'] = true;
                    $cachedData['stale'] = $anyStale;
                }

                // If stale, refetch all requested attributes
                if ($anyStale && empty($missingAttrs)) {
                    $missingAttrs = $requestedAttrs;
                }
            } else {
                $missingAttrs = $requestedAttrs;
            }

            // Merge any cached data into results
            if (count($cachedData) > 2) {
                $results[$key] = array_merge($results[$key] ?? [], $cachedData);
            }

            // Queue missing attributes for fetch
            if (!empty($missingAttrs)) {
                $fetchItem = ['type' => $type, 'id' => $id];
                if (count($missingAttrs) === 1) {
                    $fetchItem['attribute'] = $missingAttrs[0];
                } else {
                    $fetchItem['attributes'] = $missingAttrs;
                }
                if (isset($item['content-type'])) {
                    $fetchItem['content-type'] = $item['content-type'];
                }
                $toFetch[] = $fetchItem;
            }
        }

        // Fetch missing attributes from server
        if (!empty($toFetch)) {
            $fetched = $this->fetchBulk($toFetch);

            foreach ($fetched as $key => $data) {
                if (isset($data['error'])) {
                    $results[$key] = array_merge($results[$key] ?? [], $data);
                    continue;
                }

                $resultData = [
                    'type' => $data['type'],
                    'id' => $data['id'],
                    'version' => $data['version'] ?? null,
                    'last_modified' => $data['last_modified'] ?? null,
                ];

                $responseContentType = $data['content-type'] ?? 'unknown';

                // URL
                if (isset($data['url'])) {
                    $resultData['url'] = $data['url'];
                    if ($this->cacheEnabled) {
                        $cacheKey = $this->cacheKey($data['type'], $data['id'], $responseContentType, 'url');
                        $this->storeInCache($cacheKey, $data['url'], $data['version'] ?? null);
                    }
                }

                // Metadata
                if (isset($data['metadata'])) {
                    $resultData['metadata'] = $data['metadata'];
                    if ($this->cacheEnabled) {
                        $cacheKey = $this->cacheKey($data['type'], $data['id'], $responseContentType, 'metadata');
                        $this->storeInCache($cacheKey, $data['metadata'], $data['version'] ?? null);
                    }
                }

                // Content
                if (isset($data['content'])) {
                    $content = $this->decodeContent($data);
                    $resultData['content'] = $content;
                    $resultData['content-type'] = $responseContentType;
                    if ($this->cacheEnabled) {
                        $cacheKey = $this->cacheKey($data['type'], $data['id'], $responseContentType, 'content');
                        $this->storeInCache($cacheKey, $content, $data['version'] ?? null);
                    }
                }

                $results[$key] = array_merge($results[$key] ?? [], $resultData);
            }
        }

        return $results;
    }

    /**
     * Fetch multiple items from the server in a single request
     */
    private function fetchBulk(array $items): array {
        $url = $this->endpoint . '/api/content';

        $payload = json_encode([
            'items' => array_map(function($item) {
                $mapped = [
                    'type' => $item['type'] ?? 'content',
                    'id' => $item['id'] ?? '',
                ];
                // Support both 'attribute' (single) and 'attributes' (array)
                if (isset($item['attributes'])) {
                    $mapped['attributes'] = $item['attributes'];
                } elseif (isset($item['attribute'])) {
                    $mapped['attribute'] = $item['attribute'];
                }
                if (isset($item['content-type'])) {
                    $mapped['content-type'] = $item['content-type'];
                }
                return $mapped;
            }, $items),
        ]);

        $context = stream_context_create([
            'http' => [
                'method' => 'POST',
                'header' => "Content-Type: application/json\r\nX-Tenant: " . $this->tenant,
                'content' => $payload,
                'ignore_errors' => true,
            ],
        ]);

        $response = @file_get_contents($url, false, $context);
        if ($response === false) {
            return [];
        }

        $statusCode = 0;
        if (isset($http_response_header)) {
            $statusLine = $http_response_header[0] ?? '';
            if (preg_match('/HTTP\/[\d.]+\s+(\d+)/', $statusLine, $matches)) {
                $statusCode = (int)$matches[1];
            }
        }

        if ($statusCode >= 400) {
            return [];
        }

        $data = json_decode($response, true);
        return $data['items'] ?? [];
    }

    /**
     * Decode content from bulk response (handles base64 binary and JSON)
     */
    private function decodeContent(array $data): ?string {
        $content = $data['content'] ?? null;
        if ($content === null) {
            return null;
        }

        // Handle base64-encoded binary content
        if (is_string($content) && str_starts_with($content, 'base64:')) {
            return base64_decode(substr($content, 7));
        }

        // Handle JSON content (returned as parsed object)
        if (is_array($content)) {
            return json_encode($content);
        }

        return $content;
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
     * Format: velocity/cms/{type}/{id}/{content-type}/{attribute}
     */
    private function cacheKey(string $type, string $id, string $contentType = 'unknown', string $attribute = 'content'): string {
        return self::CACHE_PREFIX . $type . '/' . $id . '/' . $contentType . '/' . $attribute;
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
    private function storeInCache(string $cacheKey, mixed $content, ?string $version): void {
        $data = [
            'content' => $content,
            'version' => $version,
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
     * Clear cached content for a specific attribute
     */
    public function clearCacheFor(string $id, string $type = 'content', string $contentType = 'unknown', string $attribute = 'content'): void {
        if (!$this->cacheEnabled) {
            return;
        }

        $cacheKey = $this->cacheKey($type, $id, $contentType, $attribute);
        apcu_delete($cacheKey);
    }

    /**
     * Clear cached content for an item (all attributes)
     * Clears both the specified content-type and 'unknown' variations
     */
    public function clearCacheForAll(string $id, string $type = 'content', string $contentType = 'unknown'): void {
        if (!$this->cacheEnabled) {
            return;
        }

        $contentTypes = [$contentType];
        if ($contentType !== 'unknown') {
            $contentTypes[] = 'unknown';
        }

        foreach ($contentTypes as $ct) {
            foreach (['content', 'url', 'metadata'] as $attr) {
                apcu_delete($this->cacheKey($type, $id, $ct, $attr));
            }
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

// Support both query params and JSON body
$params = $_GET;
if (in_array($_SERVER['REQUEST_METHOD'], ['POST', 'PUT'])) {
    $body = json_decode(file_get_contents('php://input'), true);
    if (is_array($body)) {
        $params = array_merge($params, $body);
    }
}

$event = $params['event'] ?? $params['operation'] ?? 'list';
$type = $params['type'] ?? null;
$id = $params['id'] ?? null;
$contentType = $params['content-type'] ?? 'unknown';

// Handle content events (create, update, delete) by clearing cache
if (in_array($event, ['create', 'update', 'delete', 'clear'])) {
    if ($id) {
        cms()->clearCacheForAll($id, $type ?? 'content', $contentType);
        echo json_encode(['status' => 'ok', 'event' => $event, 'cleared' => ['type' => $type ?? 'content', 'id' => $id, 'content-type' => $contentType]]);
    } else {
        cms()->clearCache();
        echo json_encode(['status' => 'ok', 'event' => $event, 'cleared' => 'all']);
    }
    exit;
}

// List cache contents
$entries = [];
$prefix = 'velocity/cms/';

if (function_exists('apcu_cache_info') && class_exists('APCUIterator')) {
    foreach (new APCUIterator('/^' . preg_quote($prefix, '/') . '/') as $item) {
        $data = $item['value'];
        $entries[] = [
            'key' => $item['key'],
            'cached_at' => date('c', $data['cached_at'] ?? 0),
            'stale_at' => date('c', $data['stale_at'] ?? 0),
            'is_stale' => time() > ($data['stale_at'] ?? 0),
            'version' => $data['version'] ?? null,
            'content_length' => is_string($data['content'] ?? null) ? strlen($data['content']) : null,
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

/**
 * Helper function to get multiple content items in a single request
 *
 * Example usage:
 *   $content = contents([
 *       ['type' => 'pages', 'id' => 'header'],
 *       ['type' => 'pages', 'id' => 'footer'],
 *       ['type' => 'images', 'id' => 'hero', 'attributes' => ['url', 'metadata']],
 *   ]);
 *   echo $content['pages/header']['content'];
 *   $hero = $content['images/hero'];
 *   echo '<img src="' . $hero['url'] . '" alt="' . $hero['metadata']['alt'] . '">';
 */
function contents(array $items): array {
    return cms()->getAll($items);
}

/**
 * Helper function to get a URL for content
 * Useful for images and assets in src/href attributes
 *
 * Example: <img src="<?= contentUrl('logo', 'images') ?>">
 */
function contentUrl(string $id, string $type = 'content'): ?string {
    return cms()->getUrl($id, $type);
}

/**
 * Helper function to get metadata for content
 *
 * Example: $meta = contentMetadata('hero', 'images');
 *          echo $meta['alt'];
 */
function contentMetadata(string $id, string $type = 'content'): ?array {
    return cms()->getMetadata($id, $type);
}

/**
 * Helper function to get URL and metadata for an asset in one request
 *
 * Example: $hero = contentAsset('hero', 'images');
 *          echo '<img src="' . $hero['url'] . '" alt="' . $hero['metadata']['alt'] . '">';
 */
function contentAsset(string $id, string $type = 'content'): ?array {
    return cms()->getAsset($id, $type);
}
