package memory

import "testing"

func TestValidateSearchFilter(t *testing.T) {
	if err := ValidateSearchFilter(SearchFilter{
		Project: "mimisbrunnr",
		Query:   "jwt",
	}); err != nil {
		t.Fatalf("valid filter: %v", err)
	}

	if err := ValidateSearchFilter(SearchFilter{Query: "jwt"}); err == nil {
		t.Fatal("expected project error")
	}
	if err := ValidateSearchFilter(SearchFilter{Project: "p"}); err == nil {
		t.Fatal("expected query error")
	}
	if err := ValidateSearchFilter(SearchFilter{
		Project: "p",
		Query:   "   ",
	}); err == nil {
		t.Fatal("expected empty query error")
	}
	if err := ValidateSearchFilter(SearchFilter{
		Project: "p",
		Query:   "jwt",
		Scope:   "invalid",
	}); err == nil {
		t.Fatal("expected scope error")
	}
}
