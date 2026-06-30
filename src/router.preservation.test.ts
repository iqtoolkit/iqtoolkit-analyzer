import { describe, it, expect } from 'vitest';
import * as fc from 'fast-check';
import { getTargetFromProxyTable, ProxyRequest, ProxyTable } from './router';

/**
 * Preservation Property Tests
 *
 * **Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5**
 *
 * These tests capture the EXISTING behavior of the unfixed code for all
 * non-bug-condition inputs. They establish a baseline that must be preserved
 * after the fix is applied.
 *
 * Property 2: FOR ALL X WHERE NOT isBugCondition(X):
 *   getTargetFromProxyTable(X) = getTargetFromProxyTable'(X)
 */

/**
 * Determines if the bug condition holds for a given input.
 * When this returns true, the input triggers the substring bypass bug.
 * Preservation tests only cover inputs where this returns FALSE.
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

describe('Observation: Verify unfixed code behavior for non-buggy inputs', () => {
  it('exact host+path match returns configured target', () => {
    const req: ProxyRequest = {
      headers: { host: 'localhost:3000' },
      url: '/api/users',
    };
    const table: ProxyTable = { 'localhost:3000/api': 'http://backend' };

    // Confirm not a bug condition
    expect(isBugCondition('localhost:3000', '/api/users', 'localhost:3000/api')).toBe(false);

    const result = getTargetFromProxyTable(req, table);
    expect(result).toBe('http://backend');
  });

  it('exact host-only match returns configured target', () => {
    const req: ProxyRequest = {
      headers: { host: 'localhost:3000' },
      url: '/',
    };
    const table: ProxyTable = { 'localhost:3000': 'http://backend' };

    // Confirm not a bug condition
    expect(isBugCondition('localhost:3000', '/', 'localhost:3000')).toBe(false);

    const result = getTargetFromProxyTable(req, table);
    expect(result).toBe('http://backend');
  });

  it('path-only key matches based on request path', () => {
    const req: ProxyRequest = {
      headers: { host: 'anyhost' },
      url: '/api/data',
    };
    const table: ProxyTable = { '/api': 'http://backend' };

    // Confirm not a bug condition (path-only keys never trigger it)
    expect(isBugCondition('anyhost', '/api/data', '/api')).toBe(false);

    const result = getTargetFromProxyTable(req, table);
    expect(result).toBe('http://backend');
  });

  it('no match returns undefined', () => {
    const req: ProxyRequest = {
      headers: { host: 'unknown' },
      url: '/other',
    };
    const table: ProxyTable = { 'localhost:3000': 'http://backend' };

    // Confirm not a bug condition
    expect(isBugCondition('unknown', '/other', 'localhost:3000')).toBe(false);

    const result = getTargetFromProxyTable(req, table);
    expect(result).toBeUndefined();
  });
});

describe('Preservation Property Tests: Legitimate Routing Behavior Unchanged', () => {
  /**
   * Property 2.1: Exact host matches route correctly
   *
   * FOR ALL host-only keys WHERE host === key:
   *   getTargetFromProxyTable returns the configured target
   *
   * **Validates: Requirements 3.1**
   */
  it('property: exact host-only match always returns configured target', () => {
    // Generator for valid hostnames (alphanumeric + dots + optional port)
    const hostGen = fc.tuple(
      fc.stringOf(
        fc.oneof(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz0123456789'.split(''))),
        { minLength: 1, maxLength: 10 }
      ),
      fc.constantFrom('.com', '.io', '.net', '.org', ':3000', ':8080', ':443', '')
    ).map(([name, suffix]) => name + suffix);

    // Generator for URL paths
    const pathGen = fc.oneof(
      fc.constant('/'),
      fc.tuple(
        fc.constant('/'),
        fc.stringOf(
          fc.oneof(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz0123456789/-_'.split(''))),
          { minLength: 1, maxLength: 15 }
        )
      ).map(([slash, rest]) => slash + rest)
    );

    const targetGen = fc.stringOf(
      fc.oneof(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz0123456789:/-_.'.split(''))),
      { minLength: 5, maxLength: 20 }
    ).map((s) => 'http://' + s);

    fc.assert(
      fc.property(hostGen, pathGen, targetGen, (host, path, target) => {
        // Key equals host exactly → not a bug condition
        const key = host;
        fc.pre(!isBugCondition(host, path, key));

        const req: ProxyRequest = { headers: { host }, url: path };
        const table: ProxyTable = { [key]: target };

        const result = getTargetFromProxyTable(req, table);
        // Exact host match: indexOf will find it since hostAndPath starts with key
        expect(result).toBe(target);
      }),
      { numRuns: 100 }
    );
  });

  /**
   * Property 2.2: Host+path matches with exact host and valid path prefix route correctly
   *
   * FOR ALL host+path keys WHERE host === keyHost AND path starts with keyPath:
   *   getTargetFromProxyTable returns the configured target
   *
   * **Validates: Requirements 3.2**
   */
  it('property: exact host + valid path prefix always returns configured target', () => {
    const hostGen = fc.tuple(
      fc.stringOf(
        fc.oneof(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz0123456789'.split(''))),
        { minLength: 1, maxLength: 10 }
      ),
      fc.constantFrom('.com', '.io', ':3000', ':8080', '')
    ).map(([name, suffix]) => name + suffix);

    const pathPrefixGen = fc.tuple(
      fc.constant('/'),
      fc.stringOf(
        fc.oneof(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz0123456789'.split(''))),
        { minLength: 1, maxLength: 8 }
      )
    ).map(([slash, seg]) => slash + seg);

    const pathSuffixGen = fc.oneof(
      fc.constant(''),
      fc.tuple(
        fc.constant('/'),
        fc.stringOf(
          fc.oneof(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz0123456789'.split(''))),
          { minLength: 1, maxLength: 8 }
        )
      ).map(([s, rest]) => s + rest)
    );

    const targetGen = fc.constant('http://backend');

    fc.assert(
      fc.property(hostGen, pathPrefixGen, pathSuffixGen, targetGen, (host, pathPrefix, pathSuffix, target) => {
        const key = host + pathPrefix; // e.g., "localhost:3000/api"
        const path = pathPrefix + pathSuffix; // e.g., "/api/users"

        // Ensure not a bug condition
        fc.pre(!isBugCondition(host, path, key));
        // Ensure key is a host+path key (contains / but doesn't start with /)
        fc.pre(key.includes('/') && !key.startsWith('/'));

        const req: ProxyRequest = { headers: { host }, url: path };
        const table: ProxyTable = { [key]: target };

        const result = getTargetFromProxyTable(req, table);
        expect(result).toBe(target);
      }),
      { numRuns: 100 }
    );
  });

  /**
   * Property 2.3: Path-only keys match using existing indexOf behavior
   *
   * FOR ALL path-only keys: the result matches the existing indexOf-based matching
   *
   * **Validates: Requirements 3.4**
   */
  it('property: path-only keys match when path contains the key substring', () => {
    const hostGen = fc.stringOf(
      fc.oneof(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz0123456789.:-'.split(''))),
      { minLength: 1, maxLength: 12 }
    );

    // Path-only key always starts with /
    const pathKeyGen = fc.tuple(
      fc.constant('/'),
      fc.stringOf(
        fc.oneof(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz0123456789'.split(''))),
        { minLength: 1, maxLength: 8 }
      )
    ).map(([slash, seg]) => slash + seg);

    const pathSuffixGen = fc.oneof(
      fc.constant(''),
      fc.tuple(
        fc.constant('/'),
        fc.stringOf(
          fc.oneof(fc.constantFrom(...'abcdefghijklmnopqrstuvwxyz0123456789'.split(''))),
          { minLength: 1, maxLength: 8 }
        )
      ).map(([s, rest]) => s + rest)
    );

    const targetGen = fc.constant('http://path-backend');

    fc.assert(
      fc.property(hostGen, pathKeyGen, pathSuffixGen, targetGen, (host, pathKey, pathSuffix, target) => {
        // Request path starts with the path key
        const path = pathKey + pathSuffix;

        // Path-only keys never trigger bug condition
        fc.pre(!isBugCondition(host, path, pathKey));
        fc.pre(pathKey.startsWith('/'));

        const req: ProxyRequest = { headers: { host }, url: path };
        const table: ProxyTable = { [pathKey]: target };

        // indexOf-based matching: (host + path).indexOf(pathKey) should be > -1
        // since path starts with pathKey
        const hostAndPath = host + path;
        const expectedMatch = hostAndPath.indexOf(pathKey) > -1;

        const result = getTargetFromProxyTable(req, table);
        if (expectedMatch) {
          expect(result).toBe(target);
        } else {
          expect(result).toBeUndefined();
        }
      }),
      { numRuns: 100 }
    );
  });

  /**
   * Property 2.4: No-match cases return undefined
   *
   * FOR ALL inputs where no key matches: result is undefined
   *
   * **Validates: Requirements 3.3**
   */
  it('property: when no key matches, result is undefined', () => {
    // Generate host and path that will NOT contain any of the table keys
    // Strategy: use a completely different domain than any key
    const hostGen = fc.tuple(
      fc.constant('nomatch'),
      fc.stringOf(
        fc.oneof(fc.constantFrom(...'xyz0123456789'.split(''))),
        { minLength: 1, maxLength: 5 }
      ),
      fc.constantFrom('.test', '.local')
    ).map(([prefix, mid, suffix]) => prefix + mid + suffix);

    const pathGen = fc.tuple(
      fc.constant('/unique'),
      fc.stringOf(
        fc.oneof(fc.constantFrom(...'xyz0123456789'.split(''))),
        { minLength: 1, maxLength: 5 }
      )
    ).map(([prefix, rest]) => prefix + rest);

    // Table keys that won't appear in the generated host+path
    const tableGen = fc.dictionary(
      fc.constantFrom(
        'localhost:3000',
        'api.example.com/v1',
        'backend.internal:9090',
        'production.server.com/admin'
      ),
      fc.constant('http://should-not-match'),
      { minKeys: 1, maxKeys: 3 }
    );

    fc.assert(
      fc.property(hostGen, pathGen, tableGen, (host, path, table) => {
        const hostAndPath = host + path;

        // Verify none of the keys appear as substrings
        const anyKeyMatches = Object.keys(table).some(
          (key) => hostAndPath.indexOf(key) > -1
        );
        fc.pre(!anyKeyMatches);

        // Also verify not a bug condition for any key
        fc.pre(
          Object.keys(table).every((key) => !isBugCondition(host, path, key))
        );

        const req: ProxyRequest = { headers: { host }, url: path };
        const result = getTargetFromProxyTable(req, table);
        expect(result).toBeUndefined();
      }),
      { numRuns: 100 }
    );
  });

  /**
   * Property 2.5: First-match-wins ordering is preserved in multi-key tables
   *
   * FOR ALL multi-key tables: the first matching key determines the result
   *
   * **Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5**
   */
  it('property: first-match-wins ordering is preserved for multi-key tables', () => {
    // Create tables with known ordering where multiple keys could match
    const scenarioGen = fc.oneof(
      // Scenario: host-only key first, then host+path key for same host
      fc.record({
        host: fc.constantFrom('localhost:3000', 'api.example.com', 'myapp.io:8080'),
        path: fc.constantFrom('/api/users', '/v1/data', '/admin/panel'),
      }).map(({ host, path }) => ({
        host,
        path,
        table: {
          [host]: 'http://first-target',
          [host + path.substring(0, path.indexOf('/', 1) > 0 ? path.indexOf('/', 1) : path.length)]: 'http://second-target',
        } as ProxyTable,
        expectedTarget: 'http://first-target', // host-only key listed first
      })),

      // Scenario: two path-only keys, first one matches
      fc.record({
        host: fc.constantFrom('anyhost.com', 'server.io'),
        pathPrefix: fc.constantFrom('/api', '/v1', '/admin'),
        pathSuffix: fc.constantFrom('/users', '/data', '/items'),
      }).map(({ host, pathPrefix, pathSuffix }) => ({
        host,
        path: pathPrefix + pathSuffix,
        table: {
          [pathPrefix]: 'http://first-target',
          [pathPrefix + pathSuffix]: 'http://second-target',
        } as ProxyTable,
        expectedTarget: 'http://first-target', // shorter prefix listed first
      }))
    );

    fc.assert(
      fc.property(scenarioGen, ({ host, path, table, expectedTarget }) => {
        // Ensure none of the keys trigger the bug condition
        fc.pre(
          Object.keys(table).every((key) => !isBugCondition(host, path, key))
        );

        const req: ProxyRequest = { headers: { host }, url: path };
        const result = getTargetFromProxyTable(req, table);
        expect(result).toBe(expectedTarget);
      }),
      { numRuns: 100 }
    );
  });
});
