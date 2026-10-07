package compatibilityconsumer_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	authentication "github.com/faustbrian/go-authentication/v2"
	"github.com/faustbrian/go-authentication/v2/apikey"
	authorization "github.com/faustbrian/go-authorization"
	"github.com/faustbrian/go-authorization/acl"
	controlplane "github.com/faustbrian/go-queue-control-plane/v3"
	"github.com/faustbrian/go-queue-control-plane/v3/apihttp"
	"github.com/faustbrian/go-queue-control-plane/v3/authz"
	"github.com/faustbrian/go-queue-control-plane/v3/server"
)

func TestPublicStaticAccessAndAdministrativeComposition(t *testing.T) {
	access, err := server.NewStaticAccess([]apikey.Entry{{
		ID: "fixture-key", Key: "fixture-only-not-a-real-secret",
		Principal: authentication.PrincipalSpec{Subject: "fixture-operator"},
	}}, []acl.Entry{{
		ID: "fixture-view", Subject: authorization.Subject{
			Kind: authorization.SubjectAPIKey, ID: "fixture-operator",
		},
		Action:       authorization.Action(controlplane.PermissionView),
		ResourceType: authorization.ResourceType(controlplane.TargetWorkload),
		ResourceID:   "fixture-fleet", Tenant: "fixture-tenant", Effect: authorization.Allow,
	}})
	if err != nil {
		t.Fatalf("static access: %v", err)
	}
	apiCalls := 0
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalls++
		if err := access.Authorizer.Authorize(r.Context(), "fixture-tenant", "fixture-operator",
			controlplane.PermissionView, controlplane.Target{Kind: controlplane.TargetWorkload, Name: "fixture-fleet"}); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler, err := server.NewAdministrativeHandler(api, access.Extractor, access.Authenticator,
		access.Challenge, apihttp.SecurityConfig{})
	if err != nil {
		t.Fatalf("administrative handler: %v", err)
	}
	for _, tc := range []struct {
		name, secret          string
		wantStatus, wantCalls int
	}{
		{"accepted", "fixture-only-not-a-real-secret", http.StatusNoContent, 1},
		{"rejected", "fixture-wrong-secret", http.StatusUnauthorized, 0},
		{"missing", "", http.StatusUnauthorized, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			apiCalls = 0
			request := httptest.NewRequest(http.MethodGet, "/v1/tenants/fixture-tenant/workers", nil)
			if tc.secret != "" {
				request.Header.Set(server.APIKeyIDHeader, "fixture-key")
				request.Header.Set(server.APIKeySecretHeader, tc.secret)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tc.wantStatus || apiCalls != tc.wantCalls {
				t.Fatalf("status/calls = %d/%d, want %d/%d", response.Code, apiCalls, tc.wantStatus, tc.wantCalls)
			}
			if tc.name == "rejected" && response.Header().Get("WWW-Authenticate") != "QueueControlKey" {
				t.Fatal("rejected credential omitted the configured challenge")
			}
		})
	}
	credentialRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	credentialRequest.Header.Set(server.APIKeyIDHeader, "fixture-key")
	credentialRequest.Header.Set(server.APIKeySecretHeader, "fixture-only-not-a-real-secret")
	credential, err := access.Extractor.Extract(credentialRequest)
	if err != nil {
		t.Fatal(err)
	}
	result, err := access.Authenticator.Authenticate(context.Background(), credential)
	if err != nil {
		t.Fatal(err)
	}
	principal, ok := result.Principal()
	if !ok || principal.Subject() != "fixture-operator" {
		t.Fatal("accepted static key lost its principal identity")
	}
	ctx := authentication.ContextWithPrincipal(context.Background(), principal)
	target := controlplane.Target{Kind: controlplane.TargetWorkload, Name: "fixture-fleet"}
	if err := access.Authorizer.Authorize(ctx, "other-tenant", "fixture-operator", controlplane.PermissionView, target); !errors.Is(err, authz.ErrDenied) {
		t.Fatalf("cross-tenant decision: %v", err)
	}
	if err := access.Authorizer.Authorize(ctx, "fixture-tenant", "other-actor", controlplane.PermissionView, target); !errors.Is(err, authz.ErrActorMismatch) {
		t.Fatalf("actor mismatch decision: %v", err)
	}
}
