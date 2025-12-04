<?php
/**
 * Lightspeed Site Configuration
 *
 * Parses site.properties files (properties/YAML hybrid format)
 */

require_once __DIR__ . '/crypto.php';

class Site {
    private static ?Site $instance = null;
    private array $properties = [];
    private string $path;

    /**
     * Get the singleton Site instance
     */
    public static function instance(): Site {
        if (self::$instance === null) {
            self::$instance = new Site();
        }
        return self::$instance;
    }

    /**
     * Create a new Site instance, loading from site.properties
     */
    public function __construct(string $path = 'site.properties') {
        $this->path = $path;
        $this->load();
    }

    /**
     * Load and parse the site.properties file
     */
    private function load(): void {
        if (!file_exists($this->path)) {
            return;
        }

        $content = file_get_contents($this->path);
        $this->properties = $this->parse($content);
    }

    /**
     * Parse properties/YAML hybrid content
     */
    private function parse(string $content): array {
        $result = [];
        $lines = explode("\n", $content);
        $currentKey = null;
        $listItems = [];
        $inList = false;

        foreach ($lines as $line) {
            $trimmed = trim($line);

            // Skip empty lines and comments
            if ($trimmed === '' || str_starts_with($trimmed, '#')) {
                continue;
            }

            // Check for YAML list item
            if (str_starts_with($trimmed, '- ')) {
                if ($currentKey !== null) {
                    $listItems[] = trim(substr($trimmed, 2));
                    $inList = true;
                }
                continue;
            }

            // If we were building a list, save it
            if ($inList && $currentKey !== null) {
                $result[$currentKey] = $listItems;
                $listItems = [];
                $inList = false;
                $currentKey = null;
            }

            // Parse key=value or key: value
            $eqPos = strpos($trimmed, '=');
            $colonPos = strpos($trimmed, ':');

            // Determine which delimiter to use
            $delimiter = null;
            $delimiterPos = false;

            if ($eqPos !== false && ($colonPos === false || $eqPos < $colonPos)) {
                $delimiter = '=';
                $delimiterPos = $eqPos;
            } elseif ($colonPos !== false) {
                $delimiter = ':';
                $delimiterPos = $colonPos;
            }

            if ($delimiterPos !== false) {
                $key = trim(substr($trimmed, 0, $delimiterPos));
                $value = trim(substr($trimmed, $delimiterPos + 1));

                // Remove quotes if present
                if ((str_starts_with($value, '"') && str_ends_with($value, '"')) ||
                    (str_starts_with($value, "'") && str_ends_with($value, "'"))) {
                    $value = substr($value, 1, -1);
                }

                // Empty value might indicate a list follows
                if ($value === '') {
                    $currentKey = $key;
                    continue;
                }

                // Check for comma-separated list
                if (str_contains($value, ',')) {
                    $items = array_map('trim', explode(',', $value));
                    $result[$key] = array_filter($items, fn($item) => $item !== '');
                } else {
                    // Convert boolean strings
                    if ($value === 'true') {
                        $result[$key] = true;
                    } elseif ($value === 'false') {
                        $result[$key] = false;
                    } elseif (is_numeric($value)) {
                        // Keep as string for now - numeric values often need to stay strings
                        $result[$key] = $value;
                    } else {
                        $result[$key] = $value;
                    }
                }

                $currentKey = $key;
            }
        }

        // Don't forget trailing list
        if ($inList && $currentKey !== null) {
            $result[$currentKey] = $listItems;
        }

        return $result;
    }

    /**
     * Get a string value
     */
    public function get(string $key, string $default = ''): string {
        $value = $this->properties[$key] ?? null;

        if ($value === null) {
            return $default;
        }

        if (is_array($value)) {
            return implode(',', $value);
        }

        return (string) $value;
    }

    /**
     * Get a boolean value
     */
    public function getBoolean(string $key, bool $default = false): bool {
        $value = $this->properties[$key] ?? null;

        if ($value === null) {
            return $default;
        }

        if (is_bool($value)) {
            return $value;
        }

        if (is_string($value)) {
            return !in_array(strtolower($value), ['', 'false', 'no', '0']);
        }

        return (bool) $value;
    }

    /**
     * Get a list value (supports comma-separated strings and YAML lists)
     */
    public function getList(string $key, array $default = []): array {
        $value = $this->properties[$key] ?? null;

        if ($value === null) {
            return $default;
        }

        if (is_array($value)) {
            return $value;
        }

        if (is_string($value) && $value !== '') {
            return array_map('trim', explode(',', $value));
        }

        return $default;
    }

    /**
     * Get an integer value
     */
    public function getInt(string $key, int $default = 0): int {
        $value = $this->properties[$key] ?? null;

        if ($value === null) {
            return $default;
        }

        return (int) $value;
    }

    public function getEncrypted(string $key, string $default = ''): string {
        $value = $this->get($key, $default);

        if ($value === '' || $value === $default) {
            return $default;
        }

        $cacheKey = 'lightspeed_dec_' . md5($this->path . ':' . $key);

        if (function_exists('apcu_fetch')) {
            $this->checkCacheValidity();

            $cached = apcu_fetch($cacheKey, $success);
            if ($success) {
                return $cached;
            }
        }

        try {
            $decrypted = crypto()->decrypt($value);

            if (function_exists('apcu_store')) {
                apcu_store($cacheKey, $decrypted, 0);
            }

            return $decrypted;
        } catch (RuntimeException $e) {
            return $value;
        }
    }

    private function checkCacheValidity(): void {
        if (!file_exists($this->path)) {
            return;
        }

        $mtimeKey = 'lightspeed_mtime_' . md5($this->path);
        $currentMtime = filemtime($this->path);
        $cachedMtime = apcu_fetch($mtimeKey, $success);

        if (!$success || $cachedMtime < $currentMtime) {
            $this->clearCache();
            apcu_store($mtimeKey, $currentMtime, 0);
        }
    }

    private function clearCache(): void {
        if (function_exists('apcu_delete') && class_exists('APCUIterator')) {
            $prefix = 'lightspeed_dec_' . md5($this->path . ':');
            foreach (new APCUIterator('/^' . preg_quote($prefix) . '/') as $item) {
                apcu_delete($item['key']);
            }
        }
    }

    /**
     * Check if a key exists
     */
    public function has(string $key): bool {
        return array_key_exists($key, $this->properties);
    }

    /**
     * Get all properties
     */
    public function all(): array {
        return $this->properties;
    }

    /**
     * Get the site name
     */
    public function name(): string {
        return $this->get('name', basename(getcwd()));
    }

    /**
     * Get the primary domain
     */
    public function domain(): string {
        return $this->get('domain', '');
    }

    /**
     * Get all domains
     */
    public function domains(): array {
        $domains = [];

        $domain = $this->get('domain');
        if ($domain !== '') {
            $domains[] = $domain;
        }

        $domainList = $this->getList('domains');
        $domains = array_merge($domains, $domainList);

        return array_unique($domains);
    }
}

/**
 * Helper function to get the site instance
 */
function site(): Site {
    return Site::instance();
}
