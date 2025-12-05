<?php
/**
 * Lightspeed CMS (Velocity)
 *
 * Read-only client for the Velocity headless CMS
 */

require_once __DIR__ . '/site.php';

class CMS {
    private string $endpoint;
    private string $tenant;

    /**
     * Create a new CMS client
     *
     * @param string|null $endpoint The Velocity API endpoint (defaults to velocity.ee)
     * @param string|null $tenant The tenant identifier (defaults to site name)
     */
    public function __construct(?string $endpoint = null, ?string $tenant = null) {
        $site = site();

        $this->endpoint = $endpoint
            ?? $site->get('velocity.endpoint')
            ?: 'https://velocity.ee';

        $this->tenant = $tenant
            ?? $site->get('velocity.tenant')
            ?: $site->name();

        $this->endpoint = rtrim($this->endpoint, '/');
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
        $url = sprintf('%s/api/content/%s/%s', $this->endpoint, urlencode($type), urlencode($id));

        $headers = [
            'X-Tenant: ' . $this->tenant,
            'Accept: ' . $this->mimeType($contentType),
        ];

        $content = $this->fetch($url, $headers);
        return $content !== false ? $content : null;
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
