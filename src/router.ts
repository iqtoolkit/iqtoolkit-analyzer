/**
 * UNFIXED (buggy) implementation of getTargetFromProxyTable.
 *
 * This function uses unanchored substring matching (indexOf) to determine
 * route matches from a proxy table. This allows crafted Host headers that
 * contain a router key as a substring to bypass routing boundaries.
 *
 * Bug: hostAndPath.indexOf(key) > -1 performs unanchored search,
 * so "evillocalhost:3000/api".indexOf("localhost:3000/api") matches.
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
 * KNOWN BUG: Uses indexOf (unanchored substring match) which allows
 * Host header spoofing via superstring injection.
 */
export function getTargetFromProxyTable(
  req: ProxyRequest,
  table: ProxyTable
): string | undefined {
  const hostAndPath = (req.headers.host || '') + req.url;

  for (const key of Object.keys(table)) {
    if (hostAndPath.indexOf(key) > -1) {
      return table[key];
    }
  }

  return undefined;
}
