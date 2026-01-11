package testutil

import "net/url"

// AdvisoryOpinionForm returns valid form data for advisory opinion submission.
func AdvisoryOpinionForm() url.Values {
	return url.Values{
		"name":             {"John Test"},
		"title":            {"City Manager"},
		"agency":           {"Test City"},
		"email":            {"john@test.gov"},
		"phone":            {"555-1234"},
		"question_summary": {"Test ethics question about contractor relationships"},
		"question_detail":  {"May I participate in contract discussions with a vendor who employs my relative?"},
	}
}

// EthicsComplaintForm returns valid form data for ethics complaint submission.
func EthicsComplaintForm() url.Values {
	return url.Values{
		"complainant_name":   {"Jane Complainant"},
		"complainant_email":  {"jane@example.com"},
		"complainant_phone":  {"555-5678"},
		"subject_name":       {"Bob Official"},
		"subject_title":      {"County Commissioner"},
		"subject_agency":     {"Test County"},
		"allegation_summary": {"Gift violation allegation"},
		"allegation_detail":  {"The subject accepted tickets to a show from a vendor seeking county contracts."},
		"statutes":           {"NRS 281A.400"},
	}
}

// AcknowledgmentForm returns valid form data for ethics acknowledgment submission.
func AcknowledgmentForm() url.Values {
	return url.Values{
		"official_name":  {"Alice Board"},
		"official_title": {"Board Member"},
		"agency":         {"State Board of Education"},
		"email":          {"alice@state.gov"},
		"phone":          {"555-9012"},
	}
}

// RecordsRequestForm returns valid form data for public records request submission.
func RecordsRequestForm() url.Values {
	return url.Values{
		"requester_name":  {"Press Association"},
		"requester_email": {"records@press.org"},
		"requester_phone": {"555-3456"},
		"request_summary": {"Request for complaint statistics"},
		"request_detail":  {"Requesting all complaint data from 2020-2024."},
	}
}

// LoginForm returns login form data.
func LoginForm(email, password string) url.Values {
	return url.Values{
		"email":    {email},
		"password": {password},
	}
}
