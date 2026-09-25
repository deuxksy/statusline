package model_test

import (
	"testing"

	"statusline/internal/model"
)

func TestQuotaCategoryMonthlyFields(t *testing.T) {
	cat := model.QuotaCategory{
		Name:            "zai",
		FiveH:           0.97,
		Monthly:         0.66,
		FiveHResetsAt:   1761368400000,
		MonthlyResetsAt: 1762923600000,
	}
	if cat.Monthly != 0.66 {
		t.Errorf("Monthly = %v, want 0.66", cat.Monthly)
	}
	if cat.MonthlyResetsAt != 1762923600000 {
		t.Errorf("MonthlyResetsAt = %v, want 1762923600000", cat.MonthlyResetsAt)
	}
}

func TestUnifiedStatusProviderField(t *testing.T) {
	st := model.NewUnifiedStatus("claude")
	if st.Provider != "" {
		t.Errorf("default Provider = %q, want empty", st.Provider)
	}
	st.Provider = "zai"
	if st.Provider != "zai" {
		t.Errorf("Provider = %q, want zai", st.Provider)
	}
}
