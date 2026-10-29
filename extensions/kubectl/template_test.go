package kubectl

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rancher/shepherd/clients/rancher"
	"github.com/rancher/shepherd/pkg/session"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	fakeClusterID = "c-m-fakecluster"
	fakeTokenID   = "token-abc"
)

type recordedRequest struct {
	method        string
	path          string
	authorization string
}

// TestCreateJobAndRunKubectlCommandsProvisionsViaProxy asserts that SA/CRB/Job creation
// and the job watch go through the Steve proxy path on the client's configured host.
func TestCreateJobAndRunKubectlCommandsProvisionsViaProxy(t *testing.T) {
	proxyPrefix := "/k8s/clusters/" + fakeClusterID + "/"
	wantCreatePaths := []string{
		proxyPrefix + "api/v1/namespaces/" + Namespace + "/serviceaccounts",
		proxyPrefix + "apis/rbac.authorization.k8s.io/v1/clusterrolebindings",
		proxyPrefix + "apis/batch/v1/namespaces/" + Namespace + "/jobs",
	}
	wantWatchPath := proxyPrefix + "apis/batch/v1/namespaces/" + Namespace + "/jobs"

	var mu sync.Mutex
	var requests []recordedRequest

	record := func(r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		requests = append(requests, recordedRequest{method: r.Method, path: r.URL.Path, authorization: r.Header.Get("Authorization")})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record(r)

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v3":
			// Management API schemas; only "token" is needed (client looks up its own token).
			w.Header().Set("X-API-Schemas", "https://"+r.Host+"/v3")
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"data":[{"id":"token","type":"schema","pluralName":"tokens",`+
				`"links":{"collection":"https://%s/v3/tokens"},`+
				`"collectionMethods":["GET"],"resourceMethods":["GET"]}]}`, r.Host)
		case r.Method == http.MethodGet && r.URL.Path == "/v3/tokens/"+fakeTokenID:
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id":"token-abc","type":"token","userID":"u-test"}`)
		case r.Method == http.MethodGet && r.URL.Path == "/v1":
			// Steve API schemas; nothing in this test reads steve types.
			w.Header().Set("X-API-Schemas", "https://"+r.Host+"/v1")
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"data":[]}`)
		case r.Method == http.MethodPost && isOneOf(r.URL.Path, wantCreatePaths):
			// Echo the created object back, as the API server would.
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("reading create body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write(body)
		case r.Method == http.MethodGet && r.URL.Path == wantWatchPath:
			// Watch stream: immediately report the job as succeeded.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, `{"type":"ADDED","object":{"apiVersion":"batch/v1","kind":"Job",`+
				`"metadata":{"name":"%s","namespace":"%s"},"status":{"succeeded":1}}}`+"\n", JobName, Namespace)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			// Hold the stream open until the client stops watching.
			select {
			case <-r.Context().Done():
			case <-time.After(30 * time.Second):
			}
		default:
			http.Error(w, fmt.Sprintf("unexpected request %s %s", r.Method, r.URL.Path), http.StatusNotFound)
		}
	})

	ts := httptest.NewTLSServer(handler)
	defer ts.Close()

	insecure, cleanup := true, true
	token := fakeTokenID + ":secret"
	rancherConfig := &rancher.Config{
		Host:     strings.TrimPrefix(ts.URL, "https://"),
		Insecure: &insecure,
		Cleanup:  &cleanup,
	}

	client, err := rancher.NewClientForConfig(token, rancherConfig, session.NewSession())
	if err != nil {
		t.Fatalf("constructing rancher client: %v", err)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Namespace: Namespace},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers:    []corev1.Container{{Name: "kubectl", Image: "rancher/shell:v0.3.0"}},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	if err := CreateJobAndRunKubectlCommands(fakeClusterID, JobName, job, client); err != nil {
		t.Fatalf("CreateJobAndRunKubectlCommands: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()

	var createPaths []string
	sawProxyWatch, sawGenerateKubeconfig := false, false
	wantAuthorization := "Bearer " + token
	for _, req := range requests {
		if strings.Contains(req.path, "generateKubeconfig") {
			sawGenerateKubeconfig = true
		}

		isProxyCreate := req.method == http.MethodPost && strings.HasPrefix(req.path, proxyPrefix)
		isProxyWatch := req.method == http.MethodGet && req.path == wantWatchPath
		if isProxyCreate {
			createPaths = append(createPaths, req.path)
		}
		if isProxyWatch {
			sawProxyWatch = true
		}
		if (isProxyCreate || isProxyWatch) && req.authorization != wantAuthorization {
			t.Errorf("request %s %s has authorization %q, want %q", req.method, req.path, req.authorization, wantAuthorization)
		}
	}

	if sawGenerateKubeconfig {
		t.Error("creation hit generateKubeconfig; provisioning must not depend on the generated kubeconfig")
	}
	if !reflect.DeepEqual(createPaths, wantCreatePaths) {
		t.Errorf("unexpected create paths\n got: %v\nwant: %v", createPaths, wantCreatePaths)
	}
	if !sawProxyWatch {
		t.Errorf("job watch did not hit proxy path %s", wantWatchPath)
	}
}

func isOneOf(path string, paths []string) bool {
	for _, p := range paths {
		if path == p {
			return true
		}
	}
	return false
}
