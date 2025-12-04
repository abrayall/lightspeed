<?php

require_once __DIR__ . '/../test.php';
require_once __DIR__ . '/../site.php';

test('parse simple key=value', function() {
    $site = new Site(__DIR__ . '/fixtures/simple.properties');
    assert_equals('mysite', $site->get('name'));
    assert_equals('example.com', $site->get('domain'));
});

test('parse key: value yaml style', function() {
    $site = new Site(__DIR__ . '/fixtures/yaml.properties');
    assert_equals('yamlsite', $site->get('name'));
});

test('get with default value', function() {
    $site = new Site(__DIR__ . '/fixtures/simple.properties');
    assert_equals('default', $site->get('nonexistent', 'default'));
});

test('getBoolean returns true', function() {
    $site = new Site(__DIR__ . '/fixtures/booleans.properties');
    assert_true($site->getBoolean('enabled'));
    assert_true($site->getBoolean('yes_value'));
    assert_true($site->getBoolean('one_value'));
});

test('getBoolean returns false', function() {
    $site = new Site(__DIR__ . '/fixtures/booleans.properties');
    assert_false($site->getBoolean('disabled'));
    assert_false($site->getBoolean('no_value'));
    assert_false($site->getBoolean('zero_value'));
});

test('getBoolean default value', function() {
    $site = new Site(__DIR__ . '/fixtures/simple.properties');
    assert_true($site->getBoolean('nonexistent', true));
    assert_false($site->getBoolean('nonexistent', false));
});

test('getList comma separated', function() {
    $site = new Site(__DIR__ . '/fixtures/lists.properties');
    $domains = $site->getList('domains');
    assert_equals(['a.com', 'b.com', 'c.com'], $domains);
});

test('getList default value', function() {
    $site = new Site(__DIR__ . '/fixtures/simple.properties');
    $list = $site->getList('nonexistent', ['default']);
    assert_equals(['default'], $list);
});

test('getInt', function() {
    $site = new Site(__DIR__ . '/fixtures/numbers.properties');
    assert_equals(8080, $site->getInt('port'));
    assert_equals(0, $site->getInt('nonexistent'));
    assert_equals(3000, $site->getInt('nonexistent', 3000));
});

test('has key', function() {
    $site = new Site(__DIR__ . '/fixtures/simple.properties');
    assert_true($site->has('name'));
    assert_false($site->has('nonexistent'));
});

test('name helper', function() {
    $site = new Site(__DIR__ . '/fixtures/simple.properties');
    assert_equals('mysite', $site->name());
});

test('domain helper', function() {
    $site = new Site(__DIR__ . '/fixtures/simple.properties');
    assert_equals('example.com', $site->domain());
});

test('domains helper combines domain and domains', function() {
    $site = new Site(__DIR__ . '/fixtures/multidomains.properties');
    $domains = $site->domains();
    assert_true(in_array('primary.com', $domains));
    assert_true(in_array('www.primary.com', $domains));
    assert_true(in_array('alt.com', $domains));
});

test('comments are ignored', function() {
    $site = new Site(__DIR__ . '/fixtures/comments.properties');
    assert_equals('value', $site->get('key'));
    assert_equals('', $site->get('# commented'));
});

run_tests();
