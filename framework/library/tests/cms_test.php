<?php

require_once __DIR__ . '/../test.php';
require_once __DIR__ . '/../cms.php';

test('constructor uses velocity.ee as default endpoint', function() {
    $cms = new CMS(null, 'test-tenant');
    assert_equals('https://velocity.ee', $cms->endpoint());
});

test('constructor uses provided endpoint', function() {
    $cms = new CMS('https://custom.cms.com', 'test-tenant');
    assert_equals('https://custom.cms.com', $cms->endpoint());
});

test('constructor strips trailing slash from endpoint', function() {
    $cms = new CMS('https://custom.cms.com/', 'test-tenant');
    assert_equals('https://custom.cms.com', $cms->endpoint());
});

test('constructor uses provided tenant', function() {
    $cms = new CMS('https://velocity.ee', 'my-tenant');
    assert_equals('my-tenant', $cms->tenant());
});

test('constructor uses velocity.tenant from site.properties', function() {
    $site = new Site(__DIR__ . '/fixtures/cms.properties');

    // Create CMS manually using site properties
    $endpoint = $site->get('velocity.endpoint') ?: 'https://velocity.ee';
    $tenant = $site->get('velocity.tenant') ?: $site->name();

    $cms = new CMS($endpoint, $tenant);
    assert_equals('https://cms.example.com', $cms->endpoint());
    assert_equals('custom-tenant', $cms->tenant());
});

test('constructor falls back to site name for tenant', function() {
    $site = new Site(__DIR__ . '/fixtures/simple.properties');

    $endpoint = $site->get('velocity.endpoint') ?: 'https://velocity.ee';
    $tenant = $site->get('velocity.tenant') ?: $site->name();

    $cms = new CMS($endpoint, $tenant);
    assert_equals('https://velocity.ee', $cms->endpoint());
    assert_equals('mysite', $cms->tenant());
});

test('cms helper returns CMS instance', function() {
    $cms = cms('https://test.com', 'tenant');
    assert_true($cms instanceof CMS);
});

test('cms helper caches instance', function() {
    $cms1 = cms('https://test1.com', 'tenant1');
    $cms2 = cms(); // Should return cached instance
    assert_equals($cms1->endpoint(), $cms2->endpoint());
});

test('cms helper creates new instance when args provided', function() {
    cms('https://first.com', 'first');
    $cms = cms('https://second.com', 'second');
    assert_equals('https://second.com', $cms->endpoint());
    assert_equals('second', $cms->tenant());
});

run_tests();
