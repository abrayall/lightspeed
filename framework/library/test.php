<?php

$tests = [];
$results = ['passed' => 0, 'failed' => 0, 'errors' => []];

function test(string $name, callable $fn): void {
    global $tests;
    $tests[$name] = $fn;
}

function assert_equals($expected, $actual, string $message = ''): void {
    if ($expected !== $actual) {
        $msg = $message ?: "Expected " . var_export($expected, true) . " but got " . var_export($actual, true);
        throw new AssertionError($msg);
    }
}

function assert_true($value, string $message = ''): void {
    if ($value !== true) {
        throw new AssertionError($message ?: "Expected true but got " . var_export($value, true));
    }
}

function assert_false($value, string $message = ''): void {
    if ($value !== false) {
        throw new AssertionError($message ?: "Expected false but got " . var_export($value, true));
    }
}

function assert_contains(string $haystack, string $needle, string $message = ''): void {
    if (strpos($haystack, $needle) === false) {
        throw new AssertionError($message ?: "Expected '$haystack' to contain '$needle'");
    }
}

function assert_throws(callable $fn, string $exceptionClass = 'Exception', string $message = ''): void {
    try {
        $fn();
        throw new AssertionError($message ?: "Expected $exceptionClass to be thrown");
    } catch (Throwable $e) {
        if (!($e instanceof $exceptionClass)) {
            throw new AssertionError($message ?: "Expected $exceptionClass but got " . get_class($e));
        }
    }
}

function run_tests(): void {
    global $tests, $results;

    echo "\n";
    echo "Running " . count($tests) . " tests...\n";
    echo "\n";

    foreach ($tests as $name => $fn) {
        try {
            $fn();
            $results['passed']++;
            echo "  ✓ $name\n";
        } catch (Throwable $e) {
            $results['failed']++;
            $results['errors'][] = ['name' => $name, 'error' => $e->getMessage()];
            echo "  ✗ $name\n";
            echo "    → " . $e->getMessage() . "\n";
        }
    }

    echo "\n";
    echo "Results: {$results['passed']} passed, {$results['failed']} failed\n";
    echo "\n";

    if ($results['failed'] > 0) {
        exit(1);
    }
}
