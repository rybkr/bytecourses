package app

import (
	"bytecourses/internal/auth"
	"bytecourses/internal/domain"
	"bytecourses/internal/store"
	"context"
)

const (
	adminEmail string = "admin@local.bytecourses.org"
	userEmail  string = "user@local.bytecourses.org"
)

func seedTestUsers(ctx context.Context, users store.UserStore) error {
	if _, ok := users.GetUserByEmail(ctx, adminEmail); !ok {
		hash, err := auth.HashPassword("admin")
		if err != nil {
			return err
		}

		if err := users.CreateUser(ctx, &domain.User{
			Email:        adminEmail,
			PasswordHash: hash,
			Role:         domain.UserRoleAdmin,
			Name:         "Admin User",
		}); err != nil {
			return err
		}
	}

	if _, ok := users.GetUserByEmail(ctx, userEmail); !ok {
		hash, err := auth.HashPassword("user")
		if err != nil {
			return err
		}

		if err := users.CreateUser(ctx, &domain.User{
			Email:        userEmail,
			PasswordHash: hash,
			Role:         domain.UserRoleStudent,
			Name:         "Guest User",
		}); err != nil {
			return err
		}
	}

	return nil
}

func seedTestProposals(ctx context.Context, users store.UserStore, proposals store.ProposalStore) error {
	if err := seedTestUsers(ctx, users); err != nil {
		return err
	}
	guestUser, _ := users.GetUserByEmail(ctx, userEmail)
	userID := guestUser.ID
	adminUser, _ := users.GetUserByEmail(ctx, adminEmail)
	adminID := adminUser.ID

	if err := proposals.CreateProposal(ctx, &domain.Proposal{
		Title:                "Practical Distributed Systems in Go",
		Summary:              "This course explores how to design and reason about distributed systems using Go, with an emphasis on tradeoffs, failure modes, and operational simplicity rather than academic formalisms.",
		Qualifications:       "I have designed and operated distributed Go services involving queues, background workers, retries, idempotency, and partial failure handling in production environments.",
		TargetAudience:       "Intermediate Go developers who want to understand how real distributed systems behave and how to build resilient services.",
		LearningObjectives:   "- Understand common distributed systems failure modes\n- Design idempotent APIs and background jobs\n- Apply retries, backoff, and timeouts correctly\n- Reason about consistency and tradeoffs",
		Outline:              "1. What makes systems distributed\n2. Failure modes and fallacies\n3. Timeouts, retries, and idempotency\n4. Background workers and queues\n5. Consistency models in practice\n6. Observability and debugging",
		AssumedPrerequisites: "- Solid Go fundamentals\n- Basic HTTP and concurrency knowledge",
		AuthorID:             userID,
		Status:               domain.ProposalStatusDraft,
	}); err != nil {
		return err
	}

	if err := proposals.CreateProposal(ctx, &domain.Proposal{
		Title:                "Building Secure APIs in Go",
		Summary:              "Students will learn how to design and implement secure HTTP APIs in Go, covering authentication, authorization, input validation, and common attack vectors.",
		Qualifications:       "I have implemented authentication and authorization systems for production Go APIs, including token-based auth, session security, and password handling.",
		TargetAudience:       "Backend developers who want to build APIs that are secure by default.",
		LearningObjectives:   "- Implement secure authentication flows\n- Apply authorization patterns correctly\n- Prevent common web vulnerabilities\n- Validate and sanitize input safely",
		Outline:              "1. Threat modeling basics\n2. Authentication strategies\n3. Authorization patterns\n4. Input validation and encoding\n5. Common attacks and defenses\n6. Security testing and reviews",
		AssumedPrerequisites: "- Go web development experience\n- Basic understanding of HTTP",
		AuthorID:             userID,
		Status:               domain.ProposalStatusSubmitted,
	}); err != nil {
		return err
	}

	if err := proposals.CreateProposal(ctx, &domain.Proposal{
		Title:                "Designing and Shipping a Go Web App",
		Summary:              "In this course, students will build a real-world web application in Go from scratch, focusing on clean architecture, persistence boundaries, authentication, and deployment. The course emphasizes pragmatic decision-making and incremental design rather than frameworks or tutorials.",
		Qualifications:       "I built and deployed ByteCourses from scratch, including authentication, persistence, user workflows, and deployment to production infrastructure. This work involved designing clean domain boundaries, implementing both in-memory and SQL-backed storage, writing automated tests, and operating the system in a live environment.",
		TargetAudience:       "Intermediate developers who already know basic Go and want to learn how to design, structure, and ship a maintainable backend service with real users and real constraints.",
		LearningObjectives:   "- Implement authentication, authorization, and session management\n- Design clear domain, handler, service, and persistence boundaries in Go\n- Write effective tests at multiple layers (unit, integration, e2e)\n- Deploy a production Go service with a database and migration system",
		Outline:              "1. Project goals and architectural boundaries\n2. Domain modeling and invariants\n3. HTTP handlers and middleware design\n4. Persistence interfaces and store implementations\n5. Authentication, sessions, and password security\n6. Testing strategies (memstore vs SQL, API tests)\n7. Migrations and schema evolution\n8. Deployment and operational concerns",
		AssumedPrerequisites: "- Comfortable with Go syntax and tooling\n- Basic understanding of HTTP and REST APIs\n- Familiarity with SQL fundamentals is helpful but not required",
		AuthorID:             userID,
		Status:               domain.ProposalStatusApproved,
	}); err != nil {
		return err
	}

	if err := proposals.CreateProposal(ctx, &domain.Proposal{
		Title:                "Testing Go Applications End to End",
		Summary:              "This course focuses on building confidence in Go systems through effective testing strategies across unit, integration, and end-to-end layers.",
		Qualifications:       "I have written and maintained extensive automated test suites for Go services, including database-backed integration tests and full API-level tests.",
		TargetAudience:       "Go developers who want to improve test quality and reduce production regressions.",
		LearningObjectives:   "- Structure code for testability\n- Write meaningful unit and integration tests\n- Manage test data and environments\n- Balance test speed and coverage",
		Outline:              "1. Testing philosophy and tradeoffs\n2. Unit testing domains and services\n3. Integration testing with databases\n4. End-to-end API tests\n5. Test data and fixtures\n6. CI considerations",
		AssumedPrerequisites: "- Comfortable writing Go code\n- Familiarity with Go’s testing tools",
		AuthorID:             userID,
		Status:               domain.ProposalStatusChangesRequested,
		ReviewNotes:          "It seems that there should be more prerequisites than you have listed here.",
		ReviewerID:           &adminID,
	}); err != nil {
		return err
	}

	return nil
}

func seedTestState(ctx context.Context, users store.UserStore, proposals store.ProposalStore) error {
	if err := seedTestUsers(ctx, users); err != nil {
		return err
	}
	if err := seedTestProposals(ctx, users, proposals); err != nil {
		return err
	}
	return nil
}
