<?php
/**
 * Lightspeed version utilities
 */

function lightspeed_version() {
    $path = __DIR__ . '/version.properties';
    if (!file_exists($path)) {
        return '0.1.0';
    }
    $props = parse_ini_file($path);
    return $props['version'] ?? '0.1.0';
}
