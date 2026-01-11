package integration

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"ncoe/internal/domain"
	"ncoe/internal/testutil"
)

// TestPublicSubmissionFlows tests the complete lifecycle of public form submissions.
func TestPublicSubmissionFlows(t *testing.T) {
	t.Run("AdvisoryOpinion", func(t *testing.T) {
		ts := testutil.NewTestServer(t)
		defer ts.Close()

		// Count cases before
		beforeCount := len(ts.Repos.Case.List("", "", ""))

		// Submit form
		form := testutil.AdvisoryOpinionForm()
		resp := ts.POST("/submit/advisory-opinion", form)

		// Should redirect to confirmation
		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("expected 303 redirect, got %d", resp.StatusCode)
		}

		location := resp.Header.Get("Location")
		if !strings.Contains(location, "/submit/confirmation") {
			t.Errorf("expected redirect to confirmation, got %s", location)
		}
		if !strings.Contains(location, "case=AO-") {
			t.Errorf("expected AO case number in redirect, got %s", location)
		}

		// Verify case was created
		afterCount := len(ts.Repos.Case.List("", "", ""))
		if afterCount != beforeCount+1 {
			t.Errorf("case count: expected %d, got %d", beforeCount+1, afterCount)
		}

		// Verify case data
		newCase := findLatestCase(t, ts, "AO")
		assertCase(t, newCase, domain.CaseTypeAdvisoryOpinion, "John Test", "john@test.gov")

		// Verify deadline calculated (45 business days)
		assertDeadlineInRange(t, newCase.DueDate, 55, 70)

		// Follow redirect to confirmation
		confResp := ts.GET(location)
		dom := testutil.ParseDOM(t, confResp.Body)
		dom.AssertFullPage()
		dom.AssertContainsText(newCase.CaseNumber)
	})

	t.Run("EthicsComplaint", func(t *testing.T) {
		ts := testutil.NewTestServer(t)
		defer ts.Close()

		form := testutil.EthicsComplaintForm()
		resp := ts.POST("/submit/ethics-complaint", form)

		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("expected 303 redirect, got %d", resp.StatusCode)
		}

		location := resp.Header.Get("Location")
		if !strings.Contains(location, "case=EC-") {
			t.Errorf("expected EC case number, got %s", location)
		}

		// Verify subject info stored
		newCase := findLatestCase(t, ts, "EC")
		if newCase.SubjectName != "Bob Official" {
			t.Errorf("subject name: expected Bob Official, got %s", newCase.SubjectName)
		}
		if newCase.SubjectAgency != "Test County" {
			t.Errorf("subject agency: expected Test County, got %s", newCase.SubjectAgency)
		}
	})

	t.Run("Acknowledgment", func(t *testing.T) {
		ts := testutil.NewTestServer(t)
		defer ts.Close()

		form := testutil.AcknowledgmentForm()
		resp := ts.POST("/submit/acknowledgment", form)

		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("expected 303 redirect, got %d", resp.StatusCode)
		}

		location := resp.Header.Get("Location")
		if !strings.Contains(location, "case=EA-") {
			t.Errorf("expected EA case number, got %s", location)
		}

		newCase := findLatestCase(t, ts, "EA")
		assertCase(t, newCase, domain.CaseTypeEthicsAcknowledgment, "Alice Board", "alice@state.gov")
	})

	t.Run("RecordsRequest", func(t *testing.T) {
		ts := testutil.NewTestServer(t)
		defer ts.Close()

		form := testutil.RecordsRequestForm()
		resp := ts.POST("/submit/records-request", form)

		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("expected 303 redirect, got %d", resp.StatusCode)
		}

		location := resp.Header.Get("Location")
		if !strings.Contains(location, "case=PRR-") {
			t.Errorf("expected PRR case number, got %s", location)
		}

		newCase := findLatestCase(t, ts, "PRR")

		// Verify 5 business day deadline
		assertDeadlineInRange(t, newCase.DueDate, 4, 10)
	})
}

// TestAuthenticationFlow tests the complete auth lifecycle.
func TestAuthenticationFlow(t *testing.T) {
	ts := testutil.NewTestServer(t)
	defer ts.Close()

	t.Run("LoginCreatesSession", func(t *testing.T) {
		// Login
		resp := ts.POST("/staff/login", url.Values{
			"email":    {"test@test.gov"},
			"password": {"password"},
		})

		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("expected 303, got %d", resp.StatusCode)
		}
		if loc := resp.Header.Get("Location"); loc != "/staff/dashboard" {
			t.Errorf("expected redirect to dashboard, got %s", loc)
		}

		// Verify session created
		token := ts.SessionToken()
		if token == "" {
			t.Error("session token not set")
		}

		session := ts.Repos.Session.GetByToken(token)
		if session == nil {
			t.Error("session not found in repository")
		}
	})

	t.Run("SessionAllowsDashboardAccess", func(t *testing.T) {
		ts.Login("test@test.gov", "password")

		resp := ts.GET("/staff/dashboard")
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		dom := testutil.ParseDOM(t, resp.Body)
		dom.AssertFullPage()
		dom.AssertContainsText("Dashboard")
	})

	t.Run("LogoutRedirectsToLogin", func(t *testing.T) {
		ts.Login("test@test.gov", "password")

		resp := ts.GET("/staff/logout")
		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("expected 303, got %d", resp.StatusCode)
		}
		if loc := resp.Header.Get("Location"); loc != "/staff/login" {
			t.Errorf("expected redirect to login, got %s", loc)
		}
	})

	t.Run("AfterLogoutProtectedRoutesRedirect", func(t *testing.T) {
		ts.Login("test@test.gov", "password")
		ts.GET("/staff/logout")
		ts.ClearCookies()

		resp := ts.GET("/staff/dashboard")
		if resp.StatusCode != http.StatusSeeOther {
			t.Errorf("expected 303 redirect, got %d", resp.StatusCode)
		}
	})
}

// TestStaffCaseWorkflow tests staff case management operations.
func TestStaffCaseWorkflow(t *testing.T) {
	ts := testutil.NewTestServer(t)
	defer ts.Close()
	ts.Login("test@test.gov", "password")

	t.Run("DashboardShowsStats", func(t *testing.T) {
		resp := ts.GET("/staff/dashboard")
		dom := testutil.ParseDOM(t, resp.Body)
		dom.AssertFullPage()
		dom.AssertContainsText("Dashboard")
	})

	t.Run("CaseListShowsSeededCases", func(t *testing.T) {
		resp := ts.GET("/staff/cases")
		dom := testutil.ParseDOM(t, resp.Body)
		dom.AssertContainsText("AO-2024-042")
		dom.AssertContainsText("EC-2024-018")
	})

	t.Run("CaseListFiltersWork", func(t *testing.T) {
		resp := ts.GET("/staff/cases?type=AO")
		dom := testutil.ParseDOM(t, resp.Body)
		dom.AssertContainsText("AO-2024")
		dom.AssertNotContainsText("EC-2024-018")
	})

	t.Run("CaseListSearchByName", func(t *testing.T) {
		// Search for "John Smith" - should find AO-2024-042
		resp := ts.GET("/staff/cases?q=John+Smith")
		dom := testutil.ParseDOM(t, resp.Body)
		dom.AssertContainsText("John Smith")
		dom.AssertContainsText("AO-2024-042")
	})

	t.Run("CaseListSearchByPartialName", func(t *testing.T) {
		// Search for "Garcia" - should find Maria Garcia's case
		resp := ts.GET("/staff/cases?q=Garcia")
		dom := testutil.ParseDOM(t, resp.Body)
		dom.AssertContainsText("Maria Garcia")
	})

	t.Run("CaseDetailShowsFullInfo", func(t *testing.T) {
		resp := ts.GET("/staff/cases/1")
		dom := testutil.ParseDOM(t, resp.Body)
		dom.AssertContainsText("AO-2024-042")
		dom.AssertContainsText("John Smith")
		dom.AssertContainsText("City of Henderson")
	})

	t.Run("CaseDetailMatchesRepository", func(t *testing.T) {
		c := ts.Repos.Case.GetByID("1")
		if c == nil {
			t.Fatal("1 not in repository")
		}

		resp := ts.GET("/staff/cases/1")
		dom := testutil.ParseDOM(t, resp.Body)
		dom.AssertContainsText(c.CaseNumber)
		dom.AssertContainsText(c.SubmitterName)
	})

	t.Run("StatusUpdateReturnsHTMXTrigger", func(t *testing.T) {
		resp := ts.POST("/staff/cases/1/_status", url.Values{
			"status": {"under_review"},
		})

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
		if trigger := resp.Header.Get("HX-Trigger"); trigger != "caseUpdated" {
			t.Errorf("expected HX-Trigger=caseUpdated, got %s", trigger)
		}
	})
}

// TestCaseNumberFormat verifies case numbers follow the correct format.
func TestCaseNumberFormat(t *testing.T) {
	ts := testutil.NewTestServer(t)
	defer ts.Close()

	cases := []struct {
		name     string
		form     url.Values
		path     string
		prefix   string
		pattern  string
	}{
		{"AO", testutil.AdvisoryOpinionForm(), "/submit/advisory-opinion", "AO-", `^AO-\d{4}-\d{3}$`},
		{"EC", testutil.EthicsComplaintForm(), "/submit/ethics-complaint", "EC-", `^EC-\d{4}-\d{3}$`},
		{"EA", testutil.AcknowledgmentForm(), "/submit/acknowledgment", "EA-", `^EA-\d{4}-\d{3}$`},
		{"PRR", testutil.RecordsRequestForm(), "/submit/records-request", "PRR-", `^PRR-\d{4}-\d{3}$`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := ts.POST(tc.path, tc.form)
			location := resp.Header.Get("Location")

			// Extract case number from redirect URL
			if !strings.Contains(location, "case="+tc.prefix) {
				t.Fatalf("case number not in redirect: %s", location)
			}

			newCase := findLatestCase(t, ts, tc.prefix[:len(tc.prefix)-1])
			if !regexp.MustCompile(tc.pattern).MatchString(newCase.CaseNumber) {
				t.Errorf("case number %q doesn't match pattern %s", newCase.CaseNumber, tc.pattern)
			}
		})
	}
}

// TestPublicSearchFlow tests the public opinion search.
func TestPublicSearchFlow(t *testing.T) {
	ts := testutil.NewTestServer(t)
	defer ts.Close()

	t.Run("EmptySearchShowsPrompt", func(t *testing.T) {
		resp := ts.GET("/search")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		dom := testutil.ParseDOM(t, resp.Body)
		dom.AssertFullPage()
		dom.AssertContainsText("Search")
	})

	t.Run("SearchWithQueryReturnsResults", func(t *testing.T) {
		resp := ts.GET("/search?q=contractor")
		dom := testutil.ParseDOM(t, resp.Body)
		dom.AssertContainsText("Contractor")
	})

	t.Run("ViewOpinionWorks", func(t *testing.T) {
		resp := ts.GET("/opinions/AO-2024-010")
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
		dom := testutil.ParseDOM(t, resp.Body)
		dom.AssertContainsText("AO-2024-010")
	})
}

// TestEndToEndCaseWorkflow verifies the complete case lifecycle:
// Submit → Dashboard shows it → View details → Update status → Status persists
func TestEndToEndCaseWorkflow(t *testing.T) {
	ts := testutil.NewTestServer(t)
	defer ts.Close()

	t.Run("SubmitToDashboardToStatusUpdate", func(t *testing.T) {
		// Step 1: Count cases before submission
		beforeCount := len(ts.Repos.Case.List("", "", ""))

		// Step 2: Submit a new Advisory Opinion
		form := testutil.AdvisoryOpinionForm()
		resp := ts.POST("/submit/advisory-opinion", form)
		if resp.StatusCode != http.StatusSeeOther {
			t.Fatalf("submission failed: expected 303, got %d", resp.StatusCode)
		}

		// Extract case number from redirect
		location := resp.Header.Get("Location")
		if !strings.Contains(location, "case=AO-") {
			t.Fatalf("expected AO case number in redirect, got %s", location)
		}

		// Step 3: Verify case was created
		afterCount := len(ts.Repos.Case.List("", "", ""))
		if afterCount != beforeCount+1 {
			t.Fatalf("case not created: expected %d cases, got %d", beforeCount+1, afterCount)
		}

		// Find the new case
		newCase := findLatestCase(t, ts, "AO")
		if newCase.Status != domain.StatusSubmitted {
			t.Errorf("new case should have status 'submitted', got %s", newCase.Status)
		}

		// Step 4: Login as staff
		ts.Login("test@test.gov", "password")

		// Step 5: Verify new case appears on dashboard (recent submissions)
		dashResp := ts.GET("/staff/dashboard")
		if dashResp.StatusCode != http.StatusOK {
			t.Fatalf("dashboard failed: %d", dashResp.StatusCode)
		}
		dom := testutil.ParseDOM(t, dashResp.Body)
		if !dom.ContainsText(newCase.CaseNumber) {
			t.Errorf("dashboard should show new case %s", newCase.CaseNumber)
		}

		// Step 6: View case detail
		detailResp := ts.GET("/staff/cases/" + newCase.ID)
		if detailResp.StatusCode != http.StatusOK {
			t.Fatalf("case detail failed: %d", detailResp.StatusCode)
		}
		detailDom := testutil.ParseDOM(t, detailResp.Body)
		detailDom.AssertContainsText(newCase.CaseNumber)
		detailDom.AssertContainsText("John Test") // From fixture

		// Step 7: Update status to "under_review"
		statusResp := ts.POST("/staff/cases/"+newCase.ID+"/_status", url.Values{
			"status": {"under_review"},
		})
		if statusResp.StatusCode != http.StatusOK {
			t.Fatalf("status update failed: %d", statusResp.StatusCode)
		}

		// Step 8: Verify status was persisted
		updatedCase := ts.Repos.Case.GetByID(newCase.ID)
		if updatedCase.Status != domain.StatusUnderReview {
			t.Errorf("status not persisted: expected under_review, got %s", updatedCase.Status)
		}
	})
}

// TestNavigationIntegrity verifies that links shown in the UI point to valid resources.
// This catches mismatches between hardcoded display data and actual repository data.
func TestNavigationIntegrity(t *testing.T) {
	ts := testutil.NewTestServer(t)
	defer ts.Close()
	ts.Login("test@test.gov", "password")

	t.Run("DashboardRecentCasesExistInRepository", func(t *testing.T) {
		// The dashboard shows recent cases - verify each one exists and is navigable
		recentCaseIDs := []string{"1", "2", "3", "4", "5"}
		for _, id := range recentCaseIDs {
			c := ts.Repos.Case.GetByID(id)
			if c == nil {
				t.Errorf("dashboard references case ID %q but it doesn't exist in repository", id)
				continue
			}

			// Verify the case detail page loads
			resp := ts.GET("/staff/cases/" + id)
			if resp.StatusCode != http.StatusOK {
				t.Errorf("case %s (%s): detail page returned %d", id, c.CaseNumber, resp.StatusCode)
			}
		}
	})

	t.Run("DashboardDeadlinesReferenceValidCases", func(t *testing.T) {
		// Deadline case IDs that appear on dashboard
		deadlineCaseIDs := []string{"1", "2", "6", "7"}
		for _, id := range deadlineCaseIDs {
			c := ts.Repos.Case.GetByID(id)
			if c == nil {
				t.Errorf("deadline references case ID %q but it doesn't exist in repository", id)
				continue
			}
			if c.DueDate.IsZero() {
				t.Errorf("case %s (%s) shown in deadlines but has no due date", id, c.CaseNumber)
			}
		}
	})

	t.Run("CaseListLinksAreNavigable", func(t *testing.T) {
		// Get all cases from repository
		cases := ts.Repos.Case.List("", "", "")
		for _, c := range cases {
			resp := ts.GET("/staff/cases/" + c.ID)
			if resp.StatusCode != http.StatusOK {
				t.Errorf("case %s: detail page returned %d", c.CaseNumber, resp.StatusCode)
			}
		}
	})

	t.Run("CaseDetailShowsCorrectData", func(t *testing.T) {
		// Verify case detail page shows data matching repository
		cases := ts.Repos.Case.List("", "", "")
		for _, c := range cases {
			resp := ts.GET("/staff/cases/" + c.ID)
			dom := testutil.ParseDOM(t, resp.Body)

			// Case number should appear on detail page
			if !dom.ContainsText(c.CaseNumber) {
				t.Errorf("case %s: detail page missing case number", c.CaseNumber)
			}
			// Submitter name should appear
			if !dom.ContainsText(c.SubmitterName) {
				t.Errorf("case %s: detail page missing submitter name %q", c.CaseNumber, c.SubmitterName)
			}
		}
	})
}

// --- Helper Functions ---

func findLatestCase(t *testing.T, ts *testutil.TestServer, typePrefix string) *domain.Case {
	t.Helper()
	cases := ts.Repos.Case.List(typePrefix, "", "")
	if len(cases) == 0 {
		t.Fatalf("no %s cases found", typePrefix)
	}

	var latest *domain.Case
	for _, c := range cases {
		if latest == nil || c.CaseNumber > latest.CaseNumber {
			latest = c
		}
	}
	return latest
}

func assertCase(t *testing.T, c *domain.Case, expectedType domain.CaseType, expectedName, expectedEmail string) {
	t.Helper()
	if c.Type != expectedType {
		t.Errorf("type: expected %s, got %s", expectedType, c.Type)
	}
	if c.SubmitterName != expectedName {
		t.Errorf("submitter name: expected %s, got %s", expectedName, c.SubmitterName)
	}
	if c.SubmitterEmail != expectedEmail {
		t.Errorf("submitter email: expected %s, got %s", expectedEmail, c.SubmitterEmail)
	}
	if c.Status != domain.StatusSubmitted {
		t.Errorf("status: expected submitted, got %s", c.Status)
	}
}

func assertDeadlineInRange(t *testing.T, deadline time.Time, minDays, maxDays int) {
	t.Helper()
	if deadline.IsZero() {
		t.Error("deadline not calculated")
		return
	}

	minExpected := time.Now().AddDate(0, 0, minDays)
	maxExpected := time.Now().AddDate(0, 0, maxDays)

	if deadline.Before(minExpected) || deadline.After(maxExpected) {
		t.Errorf("deadline %v outside expected range [%d, %d] days", deadline, minDays, maxDays)
	}
}
