# Course Proposal System - Pre-Ship Security Review

## Critical Issues 🔴

## High Priority Issues 🟠
---

### 6. Silent Error Ignoring
**Location:** Multiple handlers

**In `ListMine` (line 171):**
```go
response, _ := h.proposals.ListProposalsByAuthorID(r.Context(), u.ID)
```
Errors are silently ignored.

**In `List` (line 154):**
```go
response, _ := h.proposals.ListAllSubmittedProposals(r.Context())
```
Errors are silently ignored.

**Risk:** Users see empty lists when there's actually a database error.

**Fix:** Handle errors properly and return appropriate HTTP status codes.

---

### 7. Missing Status Transition Validation in Update
**Location:** `internal/http/handlers/proposal.go:Update`

The `Update` handler modifies proposal fields but doesn't validate that status transitions are allowed. While it checks `IsAmendable()`, a malicious client could send a status field in the JSON that gets ignored, or the status could be changed indirectly.

**Note:** The handler doesn't update status directly, but there's no explicit prevention.

**Fix:** Ensure status cannot be changed via Update endpoint (should only change via Action endpoint).

---

### 8. Missing CSRF Protection
**Location:** All POST/PATCH/DELETE endpoints

No CSRF token validation on state-changing operations. While cookie-based auth with SameSite helps, explicit CSRF protection is recommended for sensitive operations.

**Fix:** Add CSRF middleware for all state-changing endpoints.

---

### 9. Missing Error Response Body on Delete
**Location:** `internal/http/handlers/proposal.go:Delete`

When `DeleteProposalByID` returns an error, the error message is sent but the status code is `500`. For `ErrNotFound`, should return `404` with appropriate message.

**Fix:** Check error type and return appropriate status codes.

---

## Medium Priority Issues 🟡

### 10. Template Bug - Spaced Status String
**Location:** `web/templates/pages/proposal_view.html:67`

```html
{{if eq .Proposal.Status " changes_requested"}}
```

Has leading space in status string - will never match, CSS class won't apply.

**Fix:** Remove space: `"changes_requested"`

---

### 11. Unused/Unsafe `safeJS` Template Function
**Location:** `internal/http/handlers/render.go:21-23`

The `safeJS` function exists but appears unused. If used with user input, it could lead to XSS.

**Current status:** Not found in templates (good), but should be removed or documented as unsafe.

**Fix:** Remove if unused, or add clear warnings about XSS risk.

---

### 12. Missing Input Sanitization for Markdown
**Location:** `internal/http/handlers/render.go:24-30`

Markdown is rendered as `template.HTML` (trusted HTML). While goldmark should sanitize, review notes and proposal content come from users and could potentially contain unsafe HTML if markdown parsing is misconfigured.

**Fix:** Ensure goldmark is configured with HTML sanitization enabled, or add explicit sanitization step.

---

### 13. No Rate Limiting
**Location:** All endpoints

No rate limiting on authentication or proposal creation endpoints.

**Risk:** Brute force attacks on login, DoS via rapid proposal creation.

**Fix:** Add rate limiting middleware, especially for:
- `/api/login`
- `/api/register`
- `/api/proposals` (POST)

---

### 14. Missing Content-Type Validation
**Location:** JSON decoding in handlers

Handlers decode JSON without verifying `Content-Type` header. Could accept non-JSON input.

**Fix:** Add middleware to verify `Content-Type: application/json` for JSON endpoints.

---

### 15. Database Transaction Safety
**Location:** `internal/store/sqlstore/proposals.go`

`UpdateProposal` performs multiple field updates but doesn't use transactions. If update partially fails, data could be inconsistent.

**Note:** PostgreSQL UPDATE is atomic, so this is lower risk, but worth noting for future complex operations.

---

### 16. Missing Validation on Review Notes Length
**Location:** `internal/http/handlers/proposal.go:Action`

Review notes can be arbitrarily long. Should have reasonable limit (e.g., 10,000 characters).

---

### 17. Proposal ID in Template Could Cause XSS
**Location:** `web/templates/pages/proposal_view.html:110`

```javascript
const proposalId = {{.Proposal.ID}};
```

While `ID` is an integer (safe), this pattern should use proper JSON encoding for safety.

**Fix:** Use `json.Marshal` or template's built-in JSON encoding.

---

## Low Priority / Code Quality Issues 🟢

### 18. Inconsistent Error Handling Patterns
Some handlers use `http.Error`, others use helper functions. Standardize error handling.

---

### 19. Missing Context Timeout Enforcement
Database operations don't have explicit timeouts (beyond context). Should add timeouts for long-running queries.

---

### 20. No Request Size Limits
Large JSON payloads could cause memory issues. Should add request size limits.

---

### 21. Missing Logging
No structured logging for:
- Failed authentication attempts
- Failed authorization checks
- Database errors
- Proposal status changes

**Fix:** Add structured logging for security events and errors.

---

### 22. Delete Operation Doesn't Check Status
**Location:** `internal/http/handlers/proposal.go:Delete`

Should probably prevent deletion of submitted/approved proposals, or require special handling.

**Note:** Domain logic allows deletion of drafts/withdrawn/rejected, which seems reasonable.

---

### 23. Missing Index on `updated_at` for List Queries
**Location:** `migrations/002_create_proposals_table.sql`

Queries order by `updated_at DESC` but there's no index. Could impact performance with many proposals.

**Fix:** Add index: `CREATE INDEX proposals_updated_at_idx ON proposals(updated_at DESC);`

---

### 24. HTTP Redirect Status Code
**Location:** `internal/http/handlers/proposal.go:160`

Uses `http.StatusSeeOther` (303) for API redirect. Should use `307` or `308` if redirect is needed, but better to avoid redirects in API.

---

## Summary

**Critical:** 4 issues that must be fixed before ship
**High:** 5 issues that should be fixed
**Medium:** 7 issues worth addressing
**Low:** 8 code quality improvements

### Recommended Pre-Ship Fixes (Critical + High Priority)
1. Remove or fix incomplete `Approve` handler
2. Add input validation for all proposal fields
3. Fix path parsing in `ProposalView`
4. Fix authorization checks to use domain logic
5. Fix API redirect in `List` handler
6. Handle errors properly in list handlers
7. Add CSRF protection
8. Add rate limiting

### Security Hardening Recommendations
- Add request size limits
- Add structured logging for security events
- Add content-type validation
- Ensure markdown sanitization
- Add database query timeouts
