<?php
/**
 * Lightspeed version utilities
 */

function lightspeed_version() {
    $props = parse_ini_file('/opt/lightspeed/version.properties');
    return $props['version'] ?? 'unknown';
}
