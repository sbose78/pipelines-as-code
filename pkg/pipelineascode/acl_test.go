package pipelineascode

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"

	"github.com/openshift-pipelines/pipelines-as-code/pkg/cli"
	ghtesthelper "github.com/openshift-pipelines/pipelines-as-code/pkg/test/github"
	"github.com/openshift-pipelines/pipelines-as-code/pkg/webvcs"
	rtesting "knative.dev/pkg/reconciler/testing"
)

func TestAclCheck(t *testing.T) {
	fakeclient, mux, _, teardown := ghtesthelper.SetupGH()
	defer teardown()

	orgallowed := "allowed"
	orgdenied := "denied"
	repoOwnerFileAllowed := "repoOwnerAllowed"

	errit := "err"

	mux.HandleFunc("/orgs/"+orgallowed+"/public_members", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, `[{"login": "login_%s"}]`, orgallowed)
	})

	mux.HandleFunc("/orgs/"+orgdenied+"/public_members", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, `[]`)
	})

	mux.HandleFunc("/orgs/"+errit+"/public_members", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, `x1x`)
	})

	mux.HandleFunc("/orgs/"+repoOwnerFileAllowed+"/public_members", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, `[]`)
	})

	mux.HandleFunc("/repos/"+repoOwnerFileAllowed+"/contents/OWNERS", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprint(rw, `{"name": "OWNERS", "path": "OWNERS", "sha": "ownerssha"}`)
	})

	mux.HandleFunc("/repos/"+repoOwnerFileAllowed+"/git/blobs/ownerssha", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, `{"content": "%s"}`, base64.RawStdEncoding.EncodeToString([]byte("approvers:\n  - approved\n")))
	})

	ctx, _ := rtesting.SetupFakeContext(t)
	gvcs := webvcs.GithubVCS{
		Client: fakeclient,
	}

	tests := []struct {
		name    string
		runinfo *webvcs.RunInfo
		allowed bool
		wantErr bool
	}{
		{
			name: "sender allowed in org",
			runinfo: &webvcs.RunInfo{
				Owner:  orgallowed,
				Sender: "login_allowed",
			},
			allowed: true,
			wantErr: false,
		},
		{
			name: "sender allowed from owner file",
			runinfo: &webvcs.RunInfo{
				Owner:  repoOwnerFileAllowed,
				Sender: "approved",
			},
			allowed: true,
			wantErr: false,
		},
		{
			name: "owner is sender is allowed",
			runinfo: &webvcs.RunInfo{
				Owner:  orgallowed,
				Sender: "allowed",
			},
			allowed: true,
			wantErr: false,
		},
		{
			name: "sender not allowed in org",
			runinfo: &webvcs.RunInfo{
				Owner:  orgdenied,
				Sender: "notallowed",
			},
			allowed: false,
			wantErr: false,
		},
		{
			name: "err it",
			runinfo: &webvcs.RunInfo{
				Owner:  errit,
				Sender: "error",
			},
			allowed: false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cs := cli.Clients{
				GithubClient: gvcs,
			}

			got, err := aclCheck(ctx, &cs, tt.runinfo)
			if (err != nil) != tt.wantErr {
				t.Errorf("aclCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.allowed {
				t.Errorf("aclCheck() = %v, want %v", got, tt.allowed)
			}
		})
	}
}
