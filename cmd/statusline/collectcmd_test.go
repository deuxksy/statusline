package main

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseProviderArg(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		want    string
		wantErr error
	}{
		{"없음", nil, "", nil},
		{"등호 형식", []string{"--provider=zai"}, "zai", nil},
		{"빈 값", []string{"--provider="}, "", nil},
		{"중복 지정 마지막 우선", []string{"--provider=zai", "--provider=chatgpt"}, "chatgpt", nil},
		{"단독 토큰", []string{"--provider", "zai"}, "", errProviderUsage},
		{"다른 인자 무시", []string{"extra", "--provider=zai"}, "zai", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseProviderArg(tc.args)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestSelectProviders(t *testing.T) {
	all := []string{"chatgpt", "zai"}
	cases := []struct {
		name           string
		provider       string
		want           []string
		wantErr        error
		wantUnknownErr bool
	}{
		{"all", "all", all, nil, false},
		{"미지정 = all", "", all, nil, false},
		{"chatgpt", "chatgpt", []string{"chatgpt"}, nil, false},
		{"zai", "zai", []string{"zai"}, nil, false},
		{"zhipu 별칭", "zhipu", []string{"zai"}, nil, false},
		{"anthropic 미지원", "anthropic", nil, errNotYetSupported, false},
		{"gemini 미지원", "gemini", nil, errNotYetSupported, false},
		{"알 수 없는 값", "foo", nil, nil, true},
		{"openai 미허용", "openai", nil, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := selectProviders(tc.provider)
			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected %v, got %v", tc.wantErr, err)
				}
				return
			}
			if tc.wantUnknownErr {
				if err == nil {
					t.Fatal("expected unknown-provider error, got nil")
				}
				if errors.Is(err, errNotYetSupported) {
					t.Fatalf("expected unknown-provider error, got not-yet-supported: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}
