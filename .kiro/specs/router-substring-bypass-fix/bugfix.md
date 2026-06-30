# Bugfix Requirements Document

## Introduction

The `getTargetFromProxyTable()` function in `src/router.ts` of the http-proxy-middleware library (v4.0.0-beta.5) contains a security vulnerability. It uses substring matching (`hostAndPath.indexOf(key) > -1`) to determine if a router proxy-table entry matches an incoming request. This allows an attacker to craft a `Host` header that is a superstring of the configured host, causing the request to be routed to an unintended backend. Applications relying on host+path router-table rules for backend segmentation, tenant routing, or separation of public/sensitive upstreams can have routing boundaries bypassed by an unauthenticated external client.

## Bug Analysis

### Current Behavior (Defect)

1.1 WHEN a router proxy-table key contains a slash (e.g., `localhost:3000/api`) AND the request's `Host` header is a superstring of the configured host (e.g., `evillocalhost:3000`) with a path matching the configured path suffix THEN the system incorrectly matches the route and proxies the request to the configured target backend

1.2 WHEN the concatenated `host + url` string contains the router key as a non-anchored substring (e.g., `"evillocalhost:3000/api".indexOf("localhost:3000/api") > -1`) THEN the system treats this as a valid match and routes the request

1.3 WHEN a router proxy-table key is a host-only key (e.g., `localhost:3000`) AND the request's `Host` header contains the key as a substring (e.g., `maliciouslocalhost:3000`) THEN the system incorrectly matches the route

### Expected Behavior (Correct)

2.1 WHEN a router proxy-table key contains a slash (e.g., `localhost:3000/api`) AND the request's `Host` header is a superstring of the configured host (e.g., `evillocalhost:3000`) THEN the system SHALL NOT match the route and SHALL NOT proxy the request to that target backend

2.2 WHEN matching a router proxy-table key against the request THEN the system SHALL require the host portion of the key to match the request's `Host` header exactly (anchored at the start of the string) before checking the path portion

2.3 WHEN a router proxy-table key is a host-only key (e.g., `localhost:3000`) AND the request's `Host` header contains the key as a substring but does not match it exactly THEN the system SHALL NOT match the route

2.4 WHEN a router proxy-table key contains a path (e.g., `localhost:3000/api`) AND the request's host matches exactly AND the request's path starts with the configured path THEN the system SHALL match the route and proxy the request to the configured target

### Unchanged Behavior (Regression Prevention)

3.1 WHEN the request's `Host` header matches a router proxy-table key exactly (e.g., Host is `localhost:3000` and key is `localhost:3000`) THEN the system SHALL CONTINUE TO route the request to the configured target

3.2 WHEN the request's `Host` header matches the host portion of a host+path key exactly AND the request path starts with the configured path (e.g., Host is `localhost:3000`, path is `/api/users`, key is `localhost:3000/api`) THEN the system SHALL CONTINUE TO route the request to the configured target

3.3 WHEN no router proxy-table key matches the request THEN the system SHALL CONTINUE TO fall through without matching (returning undefined/null)

3.4 WHEN a router proxy-table key is a path-only key (e.g., `/api`) THEN the system SHALL CONTINUE TO match based on the request path as before

3.5 WHEN the `router` option is a function instead of a proxy-table object THEN the system SHALL CONTINUE TO invoke the function and use its return value as the target

---

## Bug Condition (Formal Specification)

### Bug Condition Function

```pascal
FUNCTION isBugCondition(X)
  INPUT: X of type ProxyRequest (containing host header, url path, and router table key)
  OUTPUT: boolean

  // The bug triggers when the router key appears as a substring in the
  // concatenated host+path, but the host portion does NOT match exactly.
  LET key = X.routerKey
  LET hostAndPath = X.hostHeader + X.urlPath

  IF key contains "/" THEN
    LET keyHost = key substring before first "/"
    RETURN hostAndPath.indexOf(key) > -1 AND X.hostHeader ≠ keyHost
  ELSE
    RETURN hostAndPath.indexOf(key) > -1 AND X.hostHeader ≠ key
  END IF
END FUNCTION
```

### Property Specification — Fix Checking

```pascal
// Property: Fix Checking - Substring bypass is prevented
FOR ALL X WHERE isBugCondition(X) DO
  result ← getTargetFromProxyTable'(X)
  ASSERT result = undefined  // No route should match
END FOR
```

### Property Specification — Preservation Checking

```pascal
// Property: Preservation Checking - Legitimate matches still work
FOR ALL X WHERE NOT isBugCondition(X) DO
  ASSERT getTargetFromProxyTable(X) = getTargetFromProxyTable'(X)
END FOR
```
