# Implementation Plan

## Overview

Fix the substring bypass vulnerability in `getTargetFromProxyTable()` in `src/router.ts`. Replace the unanchored `indexOf` check with a three-branch key classifier (path-only, host+path, host-only) that enforces exact host matching, preventing crafted Host headers from bypassing routing boundaries.

## Tasks

- [x] 1. Write bug condition exploration test
  - **Property 1: Bug Condition** - Substring Bypass Allows Routing to Unintended Backend
  - **CRITICAL**: This test MUST FAIL on unfixed code - failure confirms the bug exists
  - **DO NOT attempt to fix the test or the code when it fails**
  - **NOTE**: This test encodes the expected behavior - it will validate the fix when it passes after implementation
  - **GOAL**: Surface counterexamples that demonstrate the substring bypass vulnerability exists
  - **Scoped PBT Approach**: Scope the property to concrete failing cases — crafted Host headers that are superstrings of configured router keys
  - Bug Condition from design: `isBugCondition(input)` returns true when `hostAndPath.indexOf(key) > -1` AND `hostHeader ≠ keyHost`
  - Test that `getTargetFromProxyTable()` returns `undefined` (no match) for all inputs satisfying the bug condition:
    - Host+path bypass: Key `localhost:3000/api`, Host `evillocalhost:3000`, path `/api/data` → assert no match
    - Host-only bypass: Key `localhost:3000`, Host `maliciouslocalhost:3000`, path `/` → assert no match
    - Prefix injection: Key `api.example.com/v1`, Host `fakeapi.example.com`, path `/v1/resource` → assert no match
    - Suffix injection: Key `example.com`, Host `example.com.evil.net`, path `/` → assert no match
  - Property: FOR ALL X WHERE isBugCondition(X): getTargetFromProxyTable(X) = undefined
  - Run test on UNFIXED code
  - **EXPECTED OUTCOME**: Test FAILS (this is correct - it proves the substring bypass bug exists)
  - Document counterexamples found (e.g., `getTargetFromProxyTable({host: "evillocalhost:3000", path: "/api/data"}, {"localhost:3000/api": "http://backend"})` incorrectly returns `"http://backend"` instead of `undefined`)
  - Mark task complete when test is written, run, and failure is documented
  - _Requirements: 1.1, 1.2, 1.3, 2.1, 2.2, 2.3_

- [x] 2. Write preservation property tests (BEFORE implementing fix)
  - **Property 2: Preservation** - Legitimate Routing Behavior Unchanged
  - **IMPORTANT**: Follow observation-first methodology
  - Observe behavior on UNFIXED code for non-buggy inputs (cases where isBugCondition returns false):
    - Observe: `getTargetFromProxyTable({host: "localhost:3000", path: "/api/users"}, {"localhost:3000/api": "http://backend"})` returns `"http://backend"` on unfixed code
    - Observe: `getTargetFromProxyTable({host: "localhost:3000", path: "/"}, {"localhost:3000": "http://backend"})` returns `"http://backend"` on unfixed code
    - Observe: `getTargetFromProxyTable({host: "anyhost", path: "/api/data"}, {"/api": "http://backend"})` returns `"http://backend"` on unfixed code (path-only key)
    - Observe: `getTargetFromProxyTable({host: "unknown", path: "/other"}, {"localhost:3000": "http://backend"})` returns `undefined` on unfixed code (no match)
  - Write property-based tests capturing observed behavior patterns from Preservation Requirements in design:
    - For all exact host matches: result equals configured target
    - For all host+path matches with exact host and valid path prefix: result equals configured target
    - For all path-only keys: result matches existing indexOf behavior
    - For all no-match cases: result is undefined
    - For multi-key tables: first-match-wins ordering is preserved
  - Property: FOR ALL X WHERE NOT isBugCondition(X): getTargetFromProxyTable(X) = getTargetFromProxyTable'(X)
  - Verify tests pass on UNFIXED code
  - **EXPECTED OUTCOME**: Tests PASS (this confirms baseline behavior to preserve)
  - Mark task complete when tests are written, run, and passing on unfixed code
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 3. Fix for router substring bypass vulnerability in getTargetFromProxyTable()

  - [x] 3.1 Implement the three-branch key classifier fix
    - Replace the single `hostAndPath.indexOf(key) > -1` check with a three-branch classifier in `src/router.ts` `getTargetFromProxyTable()`:
    - **Path-only keys** (key starts with `/`): preserve existing `indexOf` behavior — `(host + path).indexOf(key) > -1`
    - **Host+path keys** (key contains `/` but does not start with `/`): split key at first `/` into `keyHost` and `keyPath`, then check `host === keyHost AND path.startsWith(keyPath)`
    - **Host-only keys** (key contains no `/`): check `host === key` (exact equality)
    - Preserve iteration order over proxy-table keys (first match wins)
    - Handle edge cases: keys with ports (`localhost:3000`), multiple path segments (`api.example.com/v1/users`), trailing slashes (`localhost:3000/`), empty/undefined Host header
    - _Bug_Condition: isBugCondition(input) where hostAndPath.indexOf(key) > -1 AND hostHeader ≠ keyHost_
    - _Expected_Behavior: FOR ALL X WHERE isBugCondition(X): getTargetFromProxyTable'(X) = undefined_
    - _Preservation: FOR ALL X WHERE NOT isBugCondition(X): getTargetFromProxyTable(X) = getTargetFromProxyTable'(X)_
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 3.1, 3.2, 3.3, 3.4, 3.5_

  - [x] 3.2 Verify bug condition exploration test now passes
    - **Property 1: Expected Behavior** - Substring Bypass Prevention Confirmed
    - **IMPORTANT**: Re-run the SAME test from task 1 - do NOT write a new test
    - The test from task 1 encodes the expected behavior (bypass attempts return undefined)
    - When this test passes, it confirms the fix correctly rejects all substring-based bypass attempts
    - Run bug condition exploration test from step 1
    - **EXPECTED OUTCOME**: Test PASSES (confirms bug is fixed — crafted Host headers no longer bypass routing)
    - _Requirements: 2.1, 2.2, 2.3_

  - [x] 3.3 Verify preservation tests still pass
    - **Property 2: Preservation** - Legitimate Routing Behavior Unchanged
    - **IMPORTANT**: Re-run the SAME tests from task 2 - do NOT write new tests
    - Run preservation property tests from step 2
    - **EXPECTED OUTCOME**: Tests PASS (confirms no regressions — exact host matches, host+path matches, path-only keys, and no-match fall-through all behave identically)
    - Confirm all tests still pass after fix (no regressions)
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [-] 4. Checkpoint - Ensure all tests pass
  - Run full test suite to confirm no regressions beyond the scope of this fix
  - Verify exploration test (Property 1) passes — bypass is prevented
  - Verify preservation tests (Property 2) pass — legitimate routing unchanged
  - Confirm edge cases are handled: empty Host header, keys with ports, keys with multiple slashes, trailing slashes
  - Ensure function-based routers (non-proxy-table) are unaffected
  - Ask the user if questions arise

## Task Dependency Graph

```json
{
  "waves": [
    ["1", "2"],
    ["3.1"],
    ["3.2", "3.3"],
    ["4"]
  ]
}
```

## Notes

- Tasks 1 and 2 are independent and can be written in parallel, but both must run BEFORE the fix is implemented
- The exploration test (task 1) is expected to FAIL on unfixed code — this confirms the bug exists
- The preservation tests (task 2) are expected to PASS on unfixed code — this establishes the behavioral baseline
- After implementing the fix (task 3.1), re-running the same tests validates correctness without writing new tests
- The fix in `src/router.ts` only modifies the `getTargetFromProxyTable()` function — no other files should require changes
- Path-only keys retain `indexOf` behavior for backward compatibility
