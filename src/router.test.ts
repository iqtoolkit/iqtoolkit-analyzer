import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';
import { getTargetFromProxyTable, ProxyRequest, ProxyTable } from './router';

/**
 * Bug Condition Exploration Test
 *
 * **Validates: Requirements 1.1, 1.2, 1.3, 2.1, 2.2, 2.3**
 *
 * This test encodes the EXPECTED (correct) behavior: when a Host header is a
 * superstring of a router key's host portion but does NOT match it exactly,
 * getTargetFromProxyTable() should return undefined (no match).
 *
 * On UNFIXED code, this test will FAIL — proving the substring bypass bug exists.
 * The failure serves as a counterexample documenting the vulnerability.
 */

/**
 * Determines if the bug condition holds for a given input.
 * The bug triggers when hostAndPath.indexOf(key) > -1 AND host !== keyHost.
 */
function isBugCondition(host: string, path: string, key: string): boolean {
  const hostAndPath = host + path;

  if (key.startsWith('/')) {
    // Path-only keys are not affected by this bug
    return false;
  }

  if (key.includes('/')) {
    const keyHost = key.substring(0, key.indexOf('/'));
    return hostAndPath.indexOf(key) > -1 && host !== keyHost;
  } else {
    return hostAndPath.indexOf(key) > -1 && host !== key;
  }
}

describe('Bug Condition Exploration: Substring Bypass Allows Routing to Unintended Backend', () => {
  /**
   * Property 1: Bug Condition — Substring Bypass Prevention
   *
   * FOR ALL X WHERE isBugCondition(X): getTargetFromProxyTable(X) = undefined
   *
   * This property asserts that crafted Host headers which are superstrings of
   * configured router keys should NOT match. On unfixed code, this will FAIL
   * because indexOf performs unanchored substring matching.
   */
  it('should NOT match when Host header is a superstring of a host+path key (bypass attempt)', () => {
    // Key: localhost:3000/api, Host: evillocalhost:3000, path: /api/data
    const req: ProxyRequest = {
      headers: { host: 'evillocalhost:3000' },
      url: '/api/data',
    };
    const table: ProxyTable = { 'localhost:3000/api': 'http://backend' };

    // Assert the bug condition holds
    expect(isBugCondition('evillocalhost:3000', '/api/data', 'localhost:3000/api')).toBe(true);

    // Expected behavior: no match (bypass should be rejected)
    const result = getTargetFromProxyTable(req, table);
    expect(result).toBeUndefined();
  });

  it('should NOT match when Host header is a superstring of a host-only key (bypass attempt)', () => {
    // Key: localhost:3000, Host: maliciouslocalhost:3000, path: /
    const req: ProxyRequest = {
      headers: { host: 'maliciouslocalhost:3000' },
      url: '/',
    };
    const table: ProxyTable = { 'localhost:3000': 'http://backend' };

    // Assert the bug condition holds
    expect(isBugCondition('maliciouslocalhost:3000', '/', 'localhost:3000')).toBe(true);

    // Expected behavior: no match (bypass should be rejected)
    const result = getTargetFromProxyTable(req, table);
    expect(result).toBeUndefined();
  });

  it('should NOT match when Host header has a prefix injection of a host+path key', () => {
    // Key: api.example.com/v1, Host: fakeapi.example.com, path: /v1/resource
    const req: ProxyRequest = {
      headers: { host: 'fakeapi.example.com' },
      url: '/v1/resource',
    };
    const table: ProxyTable = { 'api.example.com/v1': 'http://backend' };

    // Assert the bug condition holds
    expect(isBugCondition('fakeapi.example.com', '/v1/resource', 'api.example.com/v1')).toBe(true);

    // Expected behavior: no match (bypass should be rejected)
    const result = getTargetFromProxyTable(req, table);
    expect(result).toBeUndefined();
  });

  it('should NOT match when Host header has a suffix injection of a host-only key', () => {
    // Key: example.com, Host: example.com.evil.net, path: /
    const req: ProxyRequest = {
      headers: { host: 'example.com.evil.net' },
      url: '/',
    };
    const table: ProxyTable = { 'example.com': 'http://backend' };

    // Assert the bug condition holds
    expect(isBugCondition('example.com.evil.net', '/', 'example.com')).toBe(true);

    // Expected behavior: no match (bypass should be rejected)
    const result = getTargetFromProxyTable(req, table);
    expect(result).toBeUndefined();
  });

  /**
   * Property-Based Test: Scoped to concrete bypass patterns
   *
   * Generates crafted Host headers that are superstrings of configured keys
   * and asserts getTargetFromProxyTable returns undefined for all of them.
   *
   * **Validates: Requirements 1.1, 1.2, 1.3, 2.1, 2.2, 2.3**
   */
  it('property: FOR ALL X WHERE isBugCondition(X): getTargetFromProxyTable(X) = undefined', () => {
    // Generator: produces (host, path, key) triples where the bug condition holds
    const bugConditionInputs = fc.oneof(
      // Host+path key with prefix injection on host
      fc.record({
        prefix: fc.stringOf(fc.char().filter((c) => c !== '/' && c !== ':' && c.trim() !== ''), { minLength: 1, maxLength: 10 }),
        keyHost: fc.constantFrom('localhost:3000', 'api.example.com', 'myapp.io:8080'),
        keyPath: fc.constantFrom('/api', '/v1', '/v1/resource', '/data'),
        requestPath: fc.constantFrom('/api/data', '/v1/resource', '/v1/users', '/data/items'),
      }).filter(({ prefix, keyHost, keyPath, requestPath }) => {
        const host = prefix + keyHost;
        const key = keyHost + keyPath;
        return isBugCondition(host, requestPath, key);
      }).map(({ prefix, keyHost, keyPath, requestPath }) => ({
        host: prefix + keyHost,
        path: requestPath,
        key: keyHost + keyPath,
        target: 'http://backend-' + keyHost,
      })),

      // Host-only key with prefix injection
      fc.record({
        prefix: fc.stringOf(fc.char().filter((c) => c !== '/' && c !== ':' && c.trim() !== ''), { minLength: 1, maxLength: 10 }),
        key: fc.constantFrom('localhost:3000', 'example.com', 'api.internal:9090'),
        path: fc.constantFrom('/', '/anything', '/some/path'),
      }).filter(({ prefix, key, path }) => {
        const host = prefix + key;
        return isBugCondition(host, path, key);
      }).map(({ prefix, key, path }) => ({
        host: prefix + key,
        path,
        key,
        target: 'http://backend-' + key,
      })),

      // Host-only key with suffix injection
      fc.record({
        suffix: fc.constantFrom('.evil.net', '.attacker.com', '.pwned.io'),
        key: fc.constantFrom('example.com', 'api.internal', 'myapp.io'),
        path: fc.constantFrom('/', '/anything'),
      }).filter(({ suffix, key, path }) => {
        const host = key + suffix;
        return isBugCondition(host, path, key);
      }).map(({ suffix, key, path }) => ({
        host: key + suffix,
        path,
        key,
        target: 'http://backend-' + key,
      }))
    );

    fc.assert(
      fc.property(bugConditionInputs, ({ host, path, key, target }) => {
        const req: ProxyRequest = {
          headers: { host },
          url: path,
        };
        const table: ProxyTable = { [key]: target };

        // Under the bug condition, the function should return undefined
        // (i.e., the crafted Host should NOT match the key)
        const result = getTargetFromProxyTable(req, table);
        expect(result).toBeUndefined();
      }),
      { numRuns: 100 }
    );
  });
});
