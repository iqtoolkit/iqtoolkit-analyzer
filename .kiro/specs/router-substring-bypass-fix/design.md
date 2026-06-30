# Router Substring Bypass Fix — Bugfix Design

## Overview

The `getTargetFromProxyTable()` function in `src/router.ts` uses unanchored substring matching (`hostAndPath.indexOf(key) > -1`) to decide whether a router proxy-table entry matches an incoming request. This allows attackers to craft `Host` headers that contain a configured key as a substring (e.g., `evillocalhost:3000` contains `localhost:3000`), bypassing routing boundaries and reaching unintended backends.

The fix replaces the single `indexOf` check with a two-part algorithm: exact host matching followed by path-prefix matching. This eliminates substring-based bypass while preserving all legitimate routing behavior.

## Glossary

- **Bug_Condition (C)**: The condition that triggers the vulnerability — when a crafted Host header containing a router key as a non-anchored substring causes an illegitimate route match
- **Property (P)**: The desired behavior when the bug condition holds — no route should match (function returns `undefined`)
- **Preservation**: Existing routing behavior that must remain unchanged — exact host matches, host+path matches, path-only keys, function routers, and fall-through on no match
- **`getTargetFromProxyTable()`**: The function in `src/router.ts` that iterates over proxy-table keys and returns the target URL for the first matching key
- **`hostAndPath`**: The concatenation of the request's `Host` header and URL path (e.g., `"localhost:3000" + "/api/users"`)
- **Router proxy-table**: An object mapping host/path patterns (keys) to backend target URLs (values)
- **Host+path key**: A router key containing a `/` separator (e.g., `localhost:3000/api`) where the portion before the first `/` is the host and the remainder is the path prefix
- **Host-only key**: A router key with no `/` (e.g., `localhost:3000`) that should match only by exact host comparison
- **Path-only key**: A router key starting with `/` (e.g., `/api`) that matches based on request path only

## Bug Details

### Bug Condition

The bug manifests when an attacker sends a request with a `Host` header that is a superstring of a configured router key's host portion. The `getTargetFromProxyTable()` function uses `hostAndPath.indexOf(key) > -1` which performs unanchored substring search, causing false-positive matches when the key appears anywhere within the concatenated host+path string.

**Formal Specification:**
```
FUNCTION isBugCondition(input)
  INPUT: input of type { hostHeader: string, urlPath: string, routerKey: string }
  OUTPUT: boolean

  LET key = input.routerKey
  LET hostAndPath = input.hostHeader + input.urlPath

  IF key starts with "/" THEN
    // Path-only keys are not affected by this bug
    RETURN false
  END IF

  IF key contains "/" THEN
    LET keyHost = key.substring(0, key.indexOf("/"))
    RETURN hostAndPath.indexOf(key) > -1
           AND input.hostHeader ≠ keyHost
  ELSE
    RETURN hostAndPath.indexOf(key) > -1
           AND input.hostHeader ≠ key
  END IF
END FUNCTION
```

### Examples

- **Host+path bypass**: Key is `localhost:3000/api`, request Host is `evillocalhost:3000`, path is `/api/data`. The concatenated `"evillocalhost:3000/api/data".indexOf("localhost:3000/api")` returns a positive index, so the route incorrectly matches. Expected: no match.
- **Host-only bypass**: Key is `localhost:3000`, request Host is `maliciouslocalhost:3000`, path is `/anything`. `"maliciouslocalhost:3000/anything".indexOf("localhost:3000")` matches. Expected: no match.
- **Prefix bypass**: Key is `api.example.com`, request Host is `fakeapi.example.com`. Substring match succeeds. Expected: no match.
- **Legitimate exact match (not a bug)**: Key is `localhost:3000/api`, Host is `localhost:3000`, path is `/api/users`. This should match and continue to match after the fix.

## Expected Behavior

### Preservation Requirements

**Unchanged Behaviors:**
- Exact host matches (Host is `localhost:3000`, key is `localhost:3000`) must continue to route correctly
- Host+path matches where the host matches exactly and path starts with the key's path (Host `localhost:3000`, path `/api/users`, key `localhost:3000/api`) must continue to route
- Path-only keys (e.g., `/api`) must continue to match based on the request path using the existing `indexOf` logic
- Function routers (`router` option as a function) must continue to be invoked and their return value used
- When no key matches, `getTargetFromProxyTable()` must continue to return `undefined`
- Key iteration order must remain unchanged (first match wins)

**Scope:**
All inputs that do NOT involve substring-based false positives on host matching should be completely unaffected by this fix. This includes:
- Requests where the Host header matches a key's host portion exactly
- Requests using path-only router keys
- Requests using function-based routers
- Requests where no router key matches at all

## Hypothesized Root Cause

Based on the bug description, the root cause is:

1. **Unanchored `indexOf` on concatenated string**: The current code concatenates `host + url` into `hostAndPath` and checks `hostAndPath.indexOf(key) > -1`. This was likely designed for convenience but fails to enforce boundary constraints between the host and path portions.

2. **No semantic distinction between key types**: The code treats all keys identically — it does not distinguish between host-only keys, host+path keys, and path-only keys. Each type requires different matching semantics:
   - Host-only: exact string equality with the Host header
   - Host+path: exact host equality + path prefix check
   - Path-only: path prefix/substring check (existing behavior is acceptable)

3. **Missing host anchoring**: The `indexOf` check allows the key to match at any position within the concatenated string. A proper fix must anchor the host comparison to the full Host header value, preventing substring injection via crafted headers.

## Correctness Properties

Property 1: Bug Condition — Substring Bypass Prevention

_For any_ request where the bug condition holds (isBugCondition returns true — the Host header is a superstring of the router key's host portion but does not match it exactly), the fixed `getTargetFromProxyTable()` function SHALL NOT match that key, returning `undefined` for that entry.

**Validates: Requirements 2.1, 2.2, 2.3**

Property 2: Preservation — Legitimate Routing Unchanged

_For any_ request where the bug condition does NOT hold (isBugCondition returns false — the Host header matches the key's host portion exactly, or the key is path-only, or no key matches), the fixed `getTargetFromProxyTable()` function SHALL produce the same result as the original function, preserving all existing legitimate routing behavior.

**Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5**

## Fix Implementation

### Changes Required

Assuming our root cause analysis is correct:

**File**: `src/router.ts`

**Function**: `getTargetFromProxyTable()`

**Specific Changes**:

1. **Classify each router key by type**: Before matching, determine if the key is:
   - A path-only key (starts with `/`)
   - A host+path key (contains `/` but does not start with `/`)
   - A host-only key (contains no `/`)

2. **Path-only keys — preserve existing behavior**: If the key starts with `/`, continue using the existing `indexOf`-based matching against the request path (or `hostAndPath`). This preserves backward compatibility for path-only routing.

3. **Host+path keys — split and match exactly**:
   - Split the key at the first `/` to extract `keyHost` and `keyPath`
   - Compare `keyHost` against the request's `Host` header using strict equality (`===`)
   - If hosts match, check that the request's URL path starts with `keyPath` (i.e., `urlPath.startsWith("/" + keyPath)` or equivalently `urlPath.startsWith(key.substring(key.indexOf("/")))`)
   - Only match if both conditions are true

4. **Host-only keys — exact host match**:
   - Compare the key directly against the request's `Host` header using strict equality (`===`)
   - Only match if the comparison is true

5. **Preserve iteration and first-match semantics**: The fix must not change the order of iteration over proxy-table keys. The first key that matches still wins.

### Algorithm Pseudocode

```
FUNCTION getTargetFromProxyTable_fixed(req, table)
  LET host = req.headers.host
  LET path = req.url

  FOR EACH (key, target) IN table DO
    IF key starts with "/" THEN
      // Path-only key: preserve existing indexOf behavior
      IF (host + path).indexOf(key) > -1 THEN
        RETURN target
      END IF
    ELSE IF key contains "/" THEN
      // Host+path key: exact host + path prefix
      LET slashIndex = key.indexOf("/")
      LET keyHost = key.substring(0, slashIndex)
      LET keyPath = key.substring(slashIndex)  // includes leading "/"

      IF host === keyHost AND path.startsWith(keyPath) THEN
        RETURN target
      END IF
    ELSE
      // Host-only key: exact host match
      IF host === key THEN
        RETURN target
      END IF
    END IF
  END FOR

  RETURN undefined
END FUNCTION
```

### Edge Cases

- **Keys with port numbers** (e.g., `localhost:3000`): Ports are part of the Host header value. Exact equality handles this correctly since `"localhost:3000" === "localhost:3000"` and `"evillocalhost:3000" !== "localhost:3000"`.
- **Keys with multiple slashes** (e.g., `api.example.com/v1/users`): The split is at the first `/` only. `keyHost = "api.example.com"`, `keyPath = "/v1/users"`. The path prefix check `path.startsWith("/v1/users")` handles nested paths correctly.
- **Trailing slashes in keys** (e.g., `localhost:3000/`): `keyHost = "localhost:3000"`, `keyPath = "/"`. Every request path starts with `/`, so this effectively matches any path on that host — same semantic as a host-only key.
- **Case sensitivity**: Host headers are case-insensitive per HTTP spec, but the original code uses case-sensitive `indexOf`. The fix preserves case-sensitive comparison to maintain backward compatibility (lowercase normalization could be a separate enhancement).
- **Empty Host header**: If `req.headers.host` is `undefined` or empty, no host-based key can match via exact equality, which is safe behavior.

## Testing Strategy

### Validation Approach

The testing strategy follows a two-phase approach: first, surface counterexamples that demonstrate the bug on unfixed code, then verify the fix works correctly and preserves existing behavior.

### Exploratory Bug Condition Checking

**Goal**: Surface counterexamples that demonstrate the bug BEFORE implementing the fix. Confirm or refute the root cause analysis. If we refute, we will need to re-hypothesize.

**Test Plan**: Write tests that construct proxy-table configurations with known keys, then send requests with crafted Host headers containing those keys as substrings. Run these tests on the UNFIXED code to observe incorrect matches.

**Test Cases**:
1. **Host+path substring bypass**: Key `localhost:3000/api`, Host `evillocalhost:3000`, path `/api/data` — expect match on unfixed code (will demonstrate bug)
2. **Host-only substring bypass**: Key `localhost:3000`, Host `maliciouslocalhost:3000`, path `/` — expect match on unfixed code (will demonstrate bug)
3. **Prefix injection**: Key `api.example.com/v1`, Host `fakeapi.example.com`, path `/v1/resource` — expect match on unfixed code (will demonstrate bug)
4. **Suffix injection**: Key `example.com`, Host `example.com.evil.net`, path `/` — expect match on unfixed code (will demonstrate bug)

**Expected Counterexamples**:
- All crafted Host headers that contain the router key as a non-anchored substring will incorrectly match
- Root cause confirmed: `indexOf` performs unanchored search allowing any position match

### Fix Checking

**Goal**: Verify that for all inputs where the bug condition holds, the fixed function produces `undefined` (no match for the offending key).

**Pseudocode:**
```
FOR ALL input WHERE isBugCondition(input) DO
  result := getTargetFromProxyTable_fixed(input.req, input.table)
  ASSERT result = undefined OR result matches a DIFFERENT key (not the bypassed one)
END FOR
```

### Preservation Checking

**Goal**: Verify that for all inputs where the bug condition does NOT hold, the fixed function produces the same result as the original function.

**Pseudocode:**
```
FOR ALL input WHERE NOT isBugCondition(input) DO
  ASSERT getTargetFromProxyTable_original(input.req, input.table)
       = getTargetFromProxyTable_fixed(input.req, input.table)
END FOR
```

**Testing Approach**: Property-based testing is recommended for preservation checking because:
- It generates many combinations of Host headers, paths, and router keys automatically
- It catches edge cases like empty strings, special characters, and unusual port numbers
- It provides strong guarantees that legitimate routing behavior is unchanged across the entire input domain

**Test Plan**: Observe behavior on UNFIXED code first for legitimate (non-bypass) inputs, then write property-based tests capturing that behavior and asserting the fixed function matches.

**Test Cases**:
1. **Exact host match preservation**: Generate random host-only keys and requests with matching Host headers; verify both original and fixed produce the same target
2. **Host+path match preservation**: Generate random host+path keys and requests with exact host + valid path prefix; verify both versions route identically
3. **Path-only key preservation**: Generate random path-only keys and requests; verify the fixed function uses the same `indexOf` logic and produces identical results
4. **No-match fall-through preservation**: Generate requests where no key matches; verify both return `undefined`
5. **Multi-key table ordering preservation**: Generate tables with multiple keys where only one matches; verify first-match-wins semantics are unchanged

### Unit Tests

- Test each key type (host-only, host+path, path-only) with exact matches → expect correct target
- Test each key type with substring bypass attempts → expect no match
- Test edge cases: empty host, keys with ports, keys with multiple path segments
- Test first-match-wins with multiple matching keys in a table
- Test that `undefined`/empty host header does not cause crashes

### Property-Based Tests

- Generate random `(host, path, key)` triples satisfying the bug condition and assert fixed function rejects them
- Generate random `(host, path, key)` triples NOT satisfying the bug condition and assert fixed function matches original
- Generate random proxy-table configurations and verify ordering/first-match semantics are preserved
- Generate adversarial host headers (prefixed, suffixed, containing the key at various positions) and verify rejection

### Integration Tests

- Test full proxy middleware flow with crafted Host headers to verify requests are NOT forwarded to the wrong backend
- Test full proxy middleware flow with legitimate Host headers to verify requests ARE forwarded correctly
- Test that function-based routers are unaffected by the change
- Test configuration with mixed key types (host-only, host+path, path-only) in the same table
