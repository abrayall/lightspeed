<?php

require_once __DIR__ . '/../test.php';
require_once __DIR__ . '/../crypto.php';

test('encrypt and decrypt roundtrip', function() {
    $crypto = new Crypto('my-test-salt');
    $plaintext = 'hello world';
    $encrypted = $crypto->encrypt($plaintext);
    $decrypted = $crypto->decrypt($encrypted);
    assert_equals($plaintext, $decrypted);
});

test('encrypted value is different from plaintext', function() {
    $crypto = new Crypto('my-test-salt');
    $plaintext = 'secret password';
    $encrypted = $crypto->encrypt($plaintext);
    assert_true($encrypted !== $plaintext);
});

test('same plaintext produces different ciphertext each time', function() {
    $crypto = new Crypto('my-test-salt');
    $plaintext = 'same input';
    $encrypted1 = $crypto->encrypt($plaintext);
    $encrypted2 = $crypto->encrypt($plaintext);
    assert_true($encrypted1 !== $encrypted2, 'Encrypted values should differ due to random IV');
});

test('different salts produce different ciphertext', function() {
    $crypto1 = new Crypto('salt-one');
    $crypto2 = new Crypto('salt-two');
    $plaintext = 'test message';
    $encrypted1 = $crypto1->encrypt($plaintext);
    $encrypted2 = $crypto2->encrypt($plaintext);
    assert_true($encrypted1 !== $encrypted2);
});

test('wrong salt fails to decrypt', function() {
    $crypto1 = new Crypto('correct-salt');
    $crypto2 = new Crypto('wrong-salt');
    $encrypted = $crypto1->encrypt('secret');
    assert_throws(function() use ($crypto2, $encrypted) {
        $crypto2->decrypt($encrypted);
    }, 'RuntimeException');
});

test('decrypt handles special characters', function() {
    $crypto = new Crypto('test-salt');
    $plaintext = "Line1\nLine2\tTabbed \"quoted\" 'single' <html> & more!";
    $encrypted = $crypto->encrypt($plaintext);
    $decrypted = $crypto->decrypt($encrypted);
    assert_equals($plaintext, $decrypted);
});

test('decrypt handles unicode', function() {
    $crypto = new Crypto('test-salt');
    $plaintext = '日本語 中文 émojis 🚀🔥';
    $encrypted = $crypto->encrypt($plaintext);
    $decrypted = $crypto->decrypt($encrypted);
    assert_equals($plaintext, $decrypted);
});

test('decrypt handles empty string', function() {
    $crypto = new Crypto('test-salt');
    $plaintext = '';
    $encrypted = $crypto->encrypt($plaintext);
    $decrypted = $crypto->decrypt($encrypted);
    assert_equals($plaintext, $decrypted);
});

test('decrypt handles long strings', function() {
    $crypto = new Crypto('test-salt');
    $plaintext = str_repeat('a', 10000);
    $encrypted = $crypto->encrypt($plaintext);
    $decrypted = $crypto->decrypt($encrypted);
    assert_equals($plaintext, $decrypted);
});

run_tests();
