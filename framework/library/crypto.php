<?php
/**
 * Lightspeed Crypto
 *
 * Encryption utilities for securing sensitive values in site.properties
 */

class Crypto {
    private const CIPHER = 'aes-256-gcm';
    private const IV_LENGTH = 12;
    private const TAG_LENGTH = 16;

    // Named keys - each will be mixed with runtime salt
    // Keys are stored as arrays of segments for obfuscation
    private const KEYS = [
        'alpha'   => ['f7a2c9e1', 'd4b8063f', '5e2a1c9d', '7b4e8f0a', '3c6d9e2f', '5a8b1c4d', '7e0f3a6b', '9c2d5e8f'],
        'beta'    => ['3e9f2a6c', '8d1b4e7f', '0a3c6d9e', '2f5a8b1c', '4d7e0f3a', '6b9c2d5e', '8f1a4b7c', '0d3e6f9a'],
        'gamma'   => ['b4e7f0a3', 'c6d9e2f5', 'a8b1c4d7', 'e0f3a6b9', 'c2d5e8f1', 'a4b7c0d3', 'e6f93e9f', '2a6c8d1b'],
        'delta'   => ['9c2d5e8f', '1a4b7c0d', '3e6f93e9', 'f2a6c8d1', 'b4e7f0a3', 'c6d9e2f5', 'a8b1c4d7', 'e0f3a6bc'],
        'epsilon' => ['1c4d7e0f', '3a6b9c2d', '5e8f1a4b', '7c0d3e6f', '9b4e7f0a', '3c6d9e2f', '5a8b3e9f', '2a6c8d1e'],
        'zeta'    => ['d7e0f3a6', 'b9c2d5e8', 'f1a4b7c0', 'd3e6f93e', '9f2a6c8d', '1b4e7f0a', '3c6d9e2f', '5a8b1c4d'],
    ];

    private string $salt;

    /**
     * Create a new Crypto instance with the given salt
     */
    public function __construct(?string $salt = null) {
        $this->salt = $salt ?? $this->loadSalt();
    }

    private function loadSalt(): string {
        $salt = getenv('LIGHTSPEED_KEY');
        if ($salt !== false && $salt !== '') {
            return $salt;
        }

        $home = getenv('HOME') ?: (getenv('USERPROFILE') ?: '');
        $keyFile = $home . '/.lightspeed/key';
        if (file_exists($keyFile)) {
            $content = trim(file_get_contents($keyFile));
            if ($content !== '') {
                return $content;
            }
        }

        $props = $this->loadSiteProperties();
        if (isset($props['key']) && $props['key'] !== '') {
            return $this->deriveKeyFromIdentifier($props['key']);
        }

        $plaintext = null;
        if (isset($props['domain']) && $props['domain'] !== '') {
            $plaintext = $props['domain'];
        } elseif (isset($props['domains']) && is_array($props['domains']) && count($props['domains']) > 0) {
            $plaintext = $props['domains'][0];
        } elseif (isset($props['name']) && $props['name'] !== '') {
            $plaintext = $props['name'];
        } else {
            $plaintext = basename(getcwd());
        }

        return $this->deriveKeyFromIdentifier($plaintext);
    }

    private function loadSiteProperties(): array {
        $path = $this->findSiteProperties();
        if ($path === null) {
            return [];
        }

        $result = [];
        $lines = file($path, FILE_IGNORE_NEW_LINES | FILE_SKIP_EMPTY_LINES);

        foreach ($lines as $line) {
            $line = trim($line);
            if ($line === '' || str_starts_with($line, '#')) {
                continue;
            }

            $eqPos = strpos($line, '=');
            $colonPos = strpos($line, ':');

            $delimPos = false;
            if ($eqPos !== false && ($colonPos === false || $eqPos < $colonPos)) {
                $delimPos = $eqPos;
            } elseif ($colonPos !== false) {
                $delimPos = $colonPos;
            }

            if ($delimPos !== false) {
                $key = trim(substr($line, 0, $delimPos));
                $value = trim(substr($line, $delimPos + 1));

                if (str_contains($value, ',')) {
                    $result[$key] = array_map('trim', explode(',', $value));
                } else {
                    $result[$key] = $value;
                }
            }
        }

        return $result;
    }

    private function findSiteProperties(): ?string {
        $candidates = [
            'site.properties',
            $_SERVER['DOCUMENT_ROOT'] . '/site.properties',
            dirname($_SERVER['SCRIPT_FILENAME'] ?? '') . '/site.properties',
        ];

        foreach ($candidates as $path) {
            if ($path && file_exists($path)) {
                return $path;
            }
        }

        return null;
    }

    private function deriveKeyFromIdentifier(string $identifier): string {
        $reversed = strrev($identifier);
        $constant = 'lightspeed';
        $interleaved = '';
        $maxLen = max(strlen($reversed), strlen($constant));
        for ($i = 0; $i < $maxLen; $i++) {
            if ($i < strlen($reversed)) $interleaved .= $reversed[$i];
            if ($i < strlen($constant)) $interleaved .= $constant[$i];
        }

        $transformed = '';
        for ($i = 0; $i < strlen($interleaved); $i++) {
            $char = $interleaved[$i];
            $transformed .= chr((ord($char) + $i * 7 + 13) % 256);
        }

        $hash1 = hash('sha256', 'v1:' . $transformed . ':' . strlen($identifier));

        $half1 = substr($hash1, 0, 32);
        $half2 = substr($hash1, 32);
        $folded = '';
        for ($i = 0; $i < 32; $i++) {
            $folded .= $half1[$i] . $half2[31 - $i];
        }

        return hash('sha256', $folded . ':' . $identifier);
    }

    /**
     * Mix salt into a key
     * Splits salt into 3 parts: prepend, insert middle, append
     */
    private function mixKey(string $key): string {
        if ($this->salt === '') {
            return $key;
        }

        $saltLen = strlen($this->salt);
        $partLen = (int) ceil($saltLen / 3);

        $part1 = substr($this->salt, 0, $partLen);
        $part2 = substr($this->salt, $partLen, $partLen);
        $part3 = substr($this->salt, $partLen * 2);

        $midpoint = (int) (strlen($key) / 2);

        // Prepend part1, insert part2 in middle, append part3
        $mixed = $part1 . substr($key, 0, $midpoint) . $part2 . substr($key, $midpoint) . $part3;

        return $mixed;
    }

    /**
     * Derive a 32-byte encryption key from the mixed key
     */
    private function deriveKey(string $mixedKey): string {
        return hash('sha256', $mixedKey, true);
    }

    /**
     * ROT13 encode/decode
     */
    private function rot13(string $text): string {
        return str_rot13($text);
    }

    /**
     * Split string in half and swap the parts
     */
    private function swapHalves(string $text): string {
        $midpoint = (int) ceil(strlen($text) / 2);
        $first = substr($text, 0, $midpoint);
        $second = substr($text, $midpoint);
        return $second . $first;
    }

    /**
     * Assemble a key from its segments
     */
    private function assembleKey(array $segments): string {
        return implode('', $segments);
    }

    /**
     * Encrypt a value
     */
    public function encrypt(string $plaintext): string {
        if ($this->salt === '') {
            throw new RuntimeException('No encryption key configured. Set LIGHTSPEED_KEY environment variable.');
        }

        // Pick a random key
        $keyNames = array_keys(self::KEYS);
        $keyName = $keyNames[random_int(0, count($keyNames) - 1)];
        $baseKey = $this->assembleKey(self::KEYS[$keyName]);

        // Mix salt into key and derive encryption key
        $mixedKey = $this->mixKey($baseKey);
        $encryptionKey = $this->deriveKey($mixedKey);

        // Prepend key name to plaintext
        $data = $keyName . ':' . $plaintext;

        // Generate IV
        $iv = random_bytes(self::IV_LENGTH);

        // Encrypt
        $tag = '';
        $ciphertext = openssl_encrypt($data, self::CIPHER, $encryptionKey, OPENSSL_RAW_DATA, $iv, $tag, '', self::TAG_LENGTH);
        if ($ciphertext === false) {
            throw new RuntimeException('Encryption failed');
        }

        // Combine IV + tag + ciphertext
        $combined = $iv . $tag . $ciphertext;

        // Base64 encode
        $encoded = base64_encode($combined);

        // ROT13
        $rotated = $this->rot13($encoded);

        // Swap halves
        $swapped = $this->swapHalves($rotated);

        return $swapped;
    }

    /**
     * Decrypt a value
     */
    public function decrypt(string $encrypted): string {
        if ($this->salt === '') {
            throw new RuntimeException('No encryption key configured. Set LIGHTSPEED_KEY environment variable.');
        }

        // Reverse swap halves
        $unswapped = $this->swapHalves($encrypted);

        // Reverse ROT13
        $unrotated = $this->rot13($unswapped);

        // Base64 decode
        $combined = base64_decode($unrotated);
        if ($combined === false) {
            throw new RuntimeException('Invalid encrypted value: base64 decode failed');
        }

        // Extract IV, tag, ciphertext
        if (strlen($combined) < self::IV_LENGTH + self::TAG_LENGTH) {
            throw new RuntimeException('Invalid encrypted value: too short');
        }

        $iv = substr($combined, 0, self::IV_LENGTH);
        $tag = substr($combined, self::IV_LENGTH, self::TAG_LENGTH);
        $ciphertext = substr($combined, self::IV_LENGTH + self::TAG_LENGTH);

        // Try each key until one works
        foreach (self::KEYS as $keyName => $segments) {
            $baseKey = $this->assembleKey($segments);
            $mixedKey = $this->mixKey($baseKey);
            $encryptionKey = $this->deriveKey($mixedKey);

            $decrypted = openssl_decrypt($ciphertext, self::CIPHER, $encryptionKey, OPENSSL_RAW_DATA, $iv, $tag);
            if ($decrypted === false) {
                continue;
            }

            // Check if decrypted text starts with key name
            $prefix = $keyName . ':';
            if (str_starts_with($decrypted, $prefix)) {
                return substr($decrypted, strlen($prefix));
            }
        }

        throw new RuntimeException('Decryption failed: no matching key found');
    }

}

/**
 * Helper function to get a Crypto instance
 */
function crypto(?string $salt = null): Crypto {
    static $instance = null;
    if ($instance === null || $salt !== null) {
        $instance = new Crypto($salt);
    }
    return $instance;
}
