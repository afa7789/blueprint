# Possible Improvements (deferred)

Items surfaced by the dead-code sweep that were resolved pragmatically and may
be worth revisiting. Each entry: what was chosen, what would be ideal, rough
effort to migrate.

## Email verification flow — half-built, removed

**Removed:** `AuthHandler.VerifyEmail` + `POST /api/v1/auth/verify-email` route.

**State at time of removal:** the handler read a Redis key `verify:<token>` and
called `users.VerifyEmail(userID)` to flip `email_verified=true`. **No code
anywhere in the project ever wrote the `verify:<token>` Redis key**, because
there was no email-sending integration. Registration auto-verifies via
`h.users.VerifyEmail(c.Context(), u.ID)` (auth.go after register), making the
HTTP endpoint unreachable in practice.

**To rebuild later:**
1. Add an SMTP/transactional-email integration (config already has SMTP env
   vars per `cmd/health/main.go`).
2. On registration, generate a token, write it to Redis with TTL, and email a
   `?token=…` link.
3. Restore the verify-email handler (the deleted version is recoverable from
   git history before this cleanup).
4. Stop auto-verifying in the registration handler.
5. Wire a frontend page that POSTs the token. Add a `email_verified` gate to
   the relevant routes/middleware.

**DB:** `users.email_verified` and `users.email_verified_at` columns remain
(migration 015) — no schema change needed to restore.

**Effort:** ~half a day, mostly on the email-sender side.

---

## Password reset — placeholder, kept

`AuthHandler.ResetPassword` (auth.go) returns success **without actually
resetting any password**. It's a security smell: the UI thinks the reset
worked, but the password is unchanged. `ForgotPassword` is also a no-op (good
fail-silent UX, but no email is sent).

Not removed because the routes/UI exist and removing them would visibly break
the "forgot password" link. Keep until a real reset flow is built.

**To fix:** same email-sender prerequisite as email verification, plus token
storage and password update in the handler.

---

## Generic admin upload — replaced cover-specific

**Was:** `POST /api/v1/admin/blog/:id/cover` (`BlogHandler.AdminUploadCover`),
which required the post to already exist.

**Is:** `POST /api/v1/admin/upload` (`AdminHandler.Upload`), which accepts a
`prefix` form field (allowlist: `uploads`, `covers`, `banners`, `linktree`,
`brand-kit`, `products`) and returns `{url}`. The frontend
(`AdminBlog.vue:uploadCover`) was already calling this URL — the cover-
specific route was dead in practice and the frontend was broken (404). Fixed
by aligning backend with the upload-before-create pattern the frontend uses.

**Trade-off:** prefix is now client-controlled (with allowlist). For S3 / CDN
migration this surface is the right place to add presigned URLs.

---

## `health_checks` table — orphaned

Migration `008_create_health.up.sql` creates a `health_checks` table for
storing service health snapshots. **No Go code reads or writes it.** The
`cmd/health` binary is a synchronous probe that returns JSON to a caller —
it doesn't persist history.

Not removed because dropping a migration in-place breaks production. To clean
up properly, write a new migration `0XX_drop_health_checks.up.sql` that
`DROP TABLE IF EXISTS health_checks` and the matching `.down.sql`.

**Alternative:** wire `cmd/health` to insert into the table on each run, which
gives free historical health data. Probably more useful than dropping it.

---

## Test-coverage gaps (informational)

Many handler methods in `internal/handlers/admin.go`, `blog.go`, `auth.go`
have 0% coverage from the unit-test suite. They are alive (registered routes
with frontend callers) but lack regression coverage. Not dead code; flagged
for the test-debt backlog.

