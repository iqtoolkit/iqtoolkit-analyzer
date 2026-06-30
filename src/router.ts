/**
 * Fixed implementation of getTargetFromProxyTable.
 *
 * Replaces the single unanchored indexOf check with a three-branch key
 * classifier that enforces exact host matching, preventing crafted Host
 * headers from bypassing routing boundaries via substring injection.
 *
 * Key classification:
 * - Path-only keys (start with "/"): preserve existing indexOf behavior
 * - Host+path keys (contain "/" but don't start with "/"): exact host + path prefix
 * - Host-only keys (no "/"): exact host equality
 */

export interface ProxyRequest {
  headers: { host?: string };
  url: string;
}

export type ProxyTable = Record<string, string>;

/**
 * Returns the target backend URL for the first matching proxy-table key,
 * or undefined if no key matches.
 *
 * Uses a three-branch classifier to match keys:
 * 1. Path-only keys (starts with "/"): indexOf on (host + path)
 * 2. Host+path keys (contains "/" but doesn't start with "/"): exact host === keyHost AND path.startsWith(keyPath)
 * 3. Host-only keys (no "/"): exact host === key
 *
 * Iteration order is preserved (first match wins).
 */
export function getTargetFromProxyTable(
  req: ProxyRequest,
  table: ProxyTable
): string | undefined {
  const host = req.headers.host || '';
  const path = req.url;

  for (const key of Object.keys(table)) {
    if (key.startsWith('/')) {
      // Path-only key: preserve existing indexOf behavior
      const hostAndPath = host + path;
      if (hostAndPath.indexOf(key) > -1) {
        return table[key];
      }
    } else if (key.indexOf('/') > -1) {
      // Host+path key: exact host match + path prefix check
      const slashIndex = key.indexOf('/');
      const keyHost = key.substring(0, slashIndex);
      const keyPath = key.substring(slashIndex); // includes leading "/"

      if (host === keyHost && path.startsWith(keyPath)) {
        return table[key];
      }
    } else {
      // Host-only key: exact host equality
      if (host === key) {
        return table[key];
      }
    }
  }

  return undefined;
}
