## **Refactor Verification & Polish Checklist**

### **Phase 1: Verify Foundation**

**Domain Layer (`internal/domain/`)**
- [+] All event types have `NewXEvent()` constructors returning pointers
- [+] All `EventName()` methods use pointer receivers: `func (e *XEvent) EventName()`
- [+] `BaseEvent.OccurredAt()` uses pointer receiver
- [+] Events have correct fields (no redundant data, all necessary context)
- [+] Domain entities (`User`, `Proposal`, `Course`, etc.) unchanged from original

**Errors Package (`internal/pkg/errors/`)**
- [+] All error types defined: `ErrNotFound`, `ErrUnauthorized`, `ErrForbidden`, etc.
- [+] `ValidationError` struct exists with `Field` and `Message`
- [+] `ValidationErrors` struct exists with `Errors []ValidationError`
- [+] Helper functions: `NewValidationError()`, `NewValidationErrors()`

**Events Package (`internal/pkg/events/`)**
- [+] `EventBus` interface defined
- [+] `InMemoryEventBus` implementation complete
- [+] `Subscribe()` and `Publish()` methods work correctly
- [+] Optional: `AsyncEventBus` if needed

**Validation Package (`internal/pkg/validation/`)**
- [+] `Validator` struct with `Validate()` method
- [+] Helper functions: `Required()`, `MinLength()`, `MaxLength()`, `Email()`, `EntityID()`
- [+] All helpers return `*ValidationError`

---

### **Phase 2: Verify Application Layer**

**For EACH entity (`proposal/`, `auth/`, `course/`, `module/`, `content/`):**

**commands.go:**
- [ ] All command structs defined
- [ ] Each command has `Validate(errs *ValidationErrors)` method
- [ ] Validation rules are complete and correct
- [ ] Field limits match database constraints

**queries.go:**
- [ ] All query structs defined
- [ ] Queries have necessary fields (IDs, UserID for auth)

**handlers.go:**
- [ ] Every command has a handler: `CreateHandler`, `UpdateHandler`, etc.
- [ ] Every query has a handler: `GetByIDHandler`, `ListAllHandler`, etc.
- [ ] Handler constructors: `NewCreateHandler()` accept correct dependencies
- [ ] Handler `Handle()` methods:
  - [ ] Accept commands/queries as **pointers**
  - [ ] Validate input first
  - [ ] Load entities from repository
  - [ ] Check authorization
  - [ ] Execute business logic
  - [ ] Publish events
  - [ ] Return appropriate types

---

### **Phase 3: Verify Infrastructure Layer**

**Persistence (`internal/infrastructure/persistence/`)**

**repository.go:**
- [ ] All repository interfaces defined
- [ ] Methods use short names: `Create()` not `CreateUser()`
- [ ] Return types: `(entity, bool)` or `([]entity, error)`
- [ ] All methods needed by handlers are present

**postgres/ implementation:**
- [ ] Moved from `internal/store/sqlstore/`
- [ ] All files renamed: `users.go` → `user.go` (singular)
- [ ] Package name is `postgres`
- [ ] Implements all repository interfaces
- [ ] No compilation errors

**memory/ implementation:**
- [ ] Moved from `internal/store/memstore/`
- [ ] All files renamed: `users.go` → `user.go` (singular)
- [ ] Package name is `memory`
- [ ] Implements all repository interfaces
- [ ] No compilation errors

**HTTP (`internal/infrastructure/http/`)**

**helpers.go:**
- [ ] `writeJSON()`, `decodeJSON()`, `handleServiceError()`
- [ ] `handleServiceError()` handles `ValidationErrors` specially
- [ ] `isHTTPS()`, `baseURL()` helpers present

**middleware.go:**
- [ ] Moved from `internal/http/middleware/`
- [ ] `RequireUser()`, `RequireLogin()`, `RequireProposal()`, `RequireCourse()`, `RequireModule()`
- [ ] Context helpers: `UserFromContext()`, `ProposalFromContext()`, etc.
- [ ] All middleware accept repository dependencies

**For EACH handler file (`auth.go`, `proposals.go`, `courses.go`, etc.):**
- [ ] Handler struct holds `*bootstrap.Container`
- [ ] Constructor: `NewAuthHandler(c *bootstrap.Container)`
- [ ] Each HTTP method delegates to command/query handler
- [ ] Commands/queries created from HTTP request data
- [ ] Errors handled via `handleServiceError()`
- [ ] No direct business logic in HTTP handlers

**router.go:**
- [ ] Moved from `internal/app/router.go`
- [ ] Package is `http` (not `app`)
- [ ] Function: `NewRouter(c *bootstrap.Container) http.Handler`
- [ ] All routes defined with correct middleware
- [ ] Handlers initialized with container

**server.go:**
- [ ] `Server` struct with HTTP server
- [ ] `NewServer()` constructor
- [ ] `Start()` and `Shutdown()` methods

**Auth (`internal/infrastructure/auth/`)**
- [ ] Moved from `internal/auth/`
- [ ] `session.go`: `SessionStore` interface
- [ ] `bcrypt.go`: `HashPassword()`, `VerifyPassword()`, `SetBcryptCost()`
- [ ] `token.go`: `GenerateToken()`
- [ ] `memory_session.go`: `MemorySessionStore` implementation

**Email (`internal/infrastructure/email/`)**
- [ ] Moved from `internal/notify/`
- [ ] `sender.go`: `Sender` interface
- [ ] `resend.go`: Resend implementation
- [ ] `null.go`: Null sender implementation
- [ ] All email methods defined

---

### **Phase 4: Verify Bootstrap**

**container.go:**
- [ ] All repositories as fields
- [ ] All command handlers as fields
- [ ] All query handlers as fields
- [ ] `EventBus`, `Validator`, `SessionStore`, `EmailSender` fields
- [ ] `NewContainer(ctx, cfg)` function
- [ ] `setupEmailSender()` method
- [ ] `setupPersistence()` method (switches between memory/postgres)
- [ ] `wireAuthHandlers()` - creates all auth handlers
- [ ] `wireProposalHandlers()` - creates all proposal handlers
- [ ] `wireCourseHandlers()` - creates all course handlers
- [ ] `wireModuleHandlers()` - creates all module handlers
- [ ] `wireContentHandlers()` - creates all content handlers
- [ ] `setupEventSubscribers()` - subscribes to events (e.g., send welcome email)
- [ ] `seedData()` method
- [ ] `Close()` method

**config.go:**
- [ ] `StorageType`: `StorageMemory`, `StoragePostgres`
- [ ] `EmailService`: `EmailServiceResend`, `EmailServiceNone`
- [ ] `Config` struct with all flags

---

### **Phase 5: Verify Main Entry Point**

**cmd/server/main.go:**
- [ ] Imports `internal/bootstrap`
- [ ] Imports `internal/infrastructure/http`
- [ ] Creates `bootstrap.Container`
- [ ] Creates `http.Server` with `NewRouter(container)`
- [ ] Graceful shutdown logic
- [ ] Flag parsing matches `Config` struct

---

### **Phase 6: Delete Old Code**

**Delete these directories/files:**
- [ ] `internal/app/` (replaced by `internal/bootstrap/`)
- [ ] `internal/services/` (replaced by `internal/application/`)
- [ ] `internal/store/` (replaced by `internal/infrastructure/persistence/`)
- [ ] `internal/http/handlers/` (replaced by `internal/infrastructure/http/`)
- [ ] `internal/http/middleware/` (merged into `internal/infrastructure/http/middleware.go`)
- [ ] `internal/notify/` (replaced by `internal/infrastructure/email/`)
- [ ] `internal/auth/` (replaced by `internal/infrastructure/auth/`)

---

### **Phase 7: Build & Test**

**Compilation:**
- [ ] `go build ./cmd/server` succeeds
- [ ] No import errors
- [ ] No undefined references

**Unit Tests:**
- [ ] `make go-test` passes
- [ ] Update test imports if needed

**E2E Tests:**
- [ ] `make py-test` passes
- [ ] Server starts successfully
- [ ] All endpoints work

**Integration:**
- [ ] Start server with memory storage: works
- [ ] Start server with postgres storage: works
- [ ] Seed users/proposals: works
- [ ] Events are published (check logs)

---

### **Phase 8: Polish**

**Code Quality:**
- [ ] Run `gofmt -w .`
- [ ] Run `go vet ./...`
- [ ] Fix any linter warnings

**Documentation:**
- [ ] Update `CLAUDE.md` with new structure
- [ ] Update README if needed
- [ ] Add comments to exported functions

**Final Verification:**
- [ ] Create a proposal via API - check logs for events
- [ ] Submit proposal - verify event subscribers work
- [ ] Test full workflow: register → create proposal → submit → approve → create course

---

**Estimated time:** 3-4 hours to verify everything thoroughly
