package rustfs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/opencenter-cloud/opencenter-cli/internal/localdev"
)

type fakeExecutor struct {
	calls   []localdev.RunOptions
	handler func(localdev.RunOptions) ([]byte, error)
}

func (f *fakeExecutor) Run(ctx context.Context, opts localdev.RunOptions) ([]byte, error) {
	f.calls = append(f.calls, opts)
	if f.handler != nil {
		return f.handler(opts)
	}
	return nil, nil
}

func (f *fakeExecutor) RunStreaming(ctx context.Context, opts localdev.RunOptions) error {
	_, err := f.Run(ctx, opts)
	return err
}

func TestDefaultSettingsUsesVerifiedImmutableImageAndOfficialPorts(t *testing.T) {
	settings := DefaultSettings("podman")
	if settings.Image != DefaultImage || !pinnedImage(settings.Image) {
		t.Fatalf("unexpected default image: %q", settings.Image)
	}
	if settings.APIPort != 9000 || settings.ConsolePort != 9001 {
		t.Fatalf("unexpected ports: api=%d console=%d", settings.APIPort, settings.ConsolePort)
	}
	if settings.AutoAttachKind {
		t.Fatal("expected Kind auto-attachment to be disabled by default")
	}
}

func TestEnsureCredentialsGeneratesProtectedFiles(t *testing.T) {
	service, err := NewService(&fakeExecutor{}, t.TempDir(), DefaultSettings("docker"))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.layout.ensure(); err != nil {
		t.Fatal(err)
	}
	creds, err := service.ensureCredentials(false)
	if err != nil {
		t.Fatal(err)
	}
	if creds.AccessKey == "" || creds.SecretKey == "" {
		t.Fatal("expected generated credentials")
	}
	for _, path := range []string{service.layout.CredentialsPath, service.layout.EnvPath} {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Fatalf("%s mode = %o, want 600", path, info.Mode().Perm())
		}
	}
}

func TestRunContainerUsesOwnershipUIDHostBindAndPodmanRelabel(t *testing.T) {
	settings := DefaultSettings("podman")
	settings.APIPort = 19000
	settings.ConsolePort = 19001
	executor := &fakeExecutor{}
	service, err := NewService(executor, t.TempDir(), settings)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.layout.ensure(); err != nil {
		t.Fatal(err)
	}
	creds := Credentials{AccessKey: "access", SecretKey: "secret"}
	if err := service.writeCredentialsEnv(creds); err != nil {
		t.Fatal(err)
	}
	if err := service.runContainer(context.Background(), creds); err != nil {
		t.Fatal(err)
	}
	args := strings.Join(executor.calls[0].Args, " ")
	for _, want := range []string{
		"--user 10001:10001",
		"-v " + service.layout.DataDir + ":/data:Z",
		"-p 127.0.0.1:19000:9000",
		"-p 127.0.0.1:19001:9001",
		managedByLabel + "=" + managedByValue,
		serviceLabel + "=" + serviceValue,
	} {
		if !strings.Contains(args, want) {
			t.Fatalf("container args %q do not contain %q", args, want)
		}
	}
	if strings.Contains(args, "access") || strings.Contains(args, "secret") {
		t.Fatal("credential value leaked into container command arguments")
	}
}

func TestExistingContainerRequiresOwnershipAndRejectsDrift(t *testing.T) {
	settings := DefaultSettings("docker")
	service, err := NewService(&fakeExecutor{}, t.TempDir(), settings)
	if err != nil {
		t.Fatal(err)
	}
	container := serviceContainer(service, Credentials{AccessKey: "access", SecretKey: "secret"})
	container.Config.Labels[managedByLabel] = "another-tool"
	if err := service.validateManagedContainer(container, nil); err == nil || !strings.Contains(err.Error(), "ownership") {
		t.Fatalf("collision error = %v", err)
	}
	container = serviceContainer(service, Credentials{AccessKey: "access", SecretKey: "secret"})
	container.Config.Labels[specHashLabel] = "drifted"
	if err := service.validateManagedContainer(container, nil); err == nil || !strings.Contains(err.Error(), "drift") {
		t.Fatalf("drift error = %v", err)
	}
}

func TestContainerValidationFailsClosedOnMissingInspectFields(t *testing.T) {
	service, err := NewService(&fakeExecutor{}, t.TempDir(), DefaultSettings("docker"))
	if err != nil {
		t.Fatal(err)
	}
	container := &containerInspect{Config: struct {
		Image  string            `json:"Image"`
		User   string            `json:"User"`
		Cmd    []string          `json:"Cmd"`
		Labels map[string]string `json:"Labels"`
	}{Labels: service.containerLabels(Credentials{AccessKey: "a", SecretKey: "b"})}}
	if err := service.validateManagedContainer(container, nil); err == nil {
		t.Fatal("expected absent inspect fields to fail closed")
	}
	complete := serviceContainer(service, Credentials{AccessKey: "a", SecretKey: "b"})
	complete.HostConfig.PortBindings["9000/tcp"][0].HostIP = "0.0.0.0"
	if err := service.validateManagedContainer(complete, nil); err == nil || !strings.Contains(err.Error(), "loopback") {
		t.Fatalf("loopback validation error = %v", err)
	}
}

func TestDestroyRefusesUnownedContainerThroughExecutor(t *testing.T) {
	service, err := NewService(&fakeExecutor{}, t.TempDir(), DefaultSettings("docker"))
	if err != nil {
		t.Fatal(err)
	}
	container := serviceContainer(service, Credentials{AccessKey: "access", SecretKey: "secret"})
	container.Config.Labels[managedByLabel] = "unrelated-tool"
	payload, err := json.Marshal(container)
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeExecutor{handler: func(opts localdev.RunOptions) ([]byte, error) {
		if len(opts.Args) >= 1 && opts.Args[0] == "inspect" {
			return payload, nil
		}
		t.Fatalf("unexpected mutating command: %s", strings.Join(opts.Args, " "))
		return nil, errors.New("unreachable")
	}}
	service.executor = executor
	if err := service.Destroy(context.Background()); err == nil || !strings.Contains(err.Error(), "ownership") {
		t.Fatalf("destroy error = %v", err)
	}
	if len(executor.calls) != 1 {
		t.Fatalf("unowned container caused mutation calls: %#v", executor.calls)
	}
}

func TestUpRejectsDriftThroughExecutorBeforeStarting(t *testing.T) {
	service, err := NewService(&fakeExecutor{}, t.TempDir(), DefaultSettings("docker"))
	if err != nil {
		t.Fatal(err)
	}
	container := serviceContainer(service, Credentials{AccessKey: "access", SecretKey: "secret"})
	container.Config.Labels[specHashLabel] = "drifted"
	payload, err := json.Marshal(container)
	if err != nil {
		t.Fatal(err)
	}
	executor := &fakeExecutor{handler: func(opts localdev.RunOptions) ([]byte, error) {
		return payload, nil
	}}
	service.executor = executor
	if _, err := service.Up(context.Background()); err == nil || !strings.Contains(err.Error(), "drift") {
		t.Fatalf("up error = %v", err)
	}
	if len(executor.calls) != 1 {
		t.Fatalf("drifted container caused additional calls: %#v", executor.calls)
	}
}

func TestExistingContainerDoesNotRegenerateMissingOrCorruptCredentials(t *testing.T) {
	service, err := NewService(&fakeExecutor{}, t.TempDir(), DefaultSettings("docker"))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.layout.ensure(); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ensureCredentials(true); err == nil {
		t.Fatal("expected missing credentials error")
	}
	if err := os.WriteFile(service.layout.CredentialsPath, []byte("not-json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ensureCredentials(true); err == nil {
		t.Fatal("expected corrupt credentials error")
	}
}

func TestPrivateWritesRejectSymlinks(t *testing.T) {
	service, err := NewService(&fakeExecutor{}, t.TempDir(), DefaultSettings("docker"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(service.layout.Root, 0o700); err != nil {
		t.Fatal(err)
	}
	target := t.TempDir()
	if err := os.Symlink(target, service.layout.CredentialsPath); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ensureCredentials(false); err == nil || !strings.Contains(err.Error(), "symlink") {
		t.Fatalf("symlink error = %v", err)
	}
}

func TestReadinessRequiresValidS3Credentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" || r.Header.Get("x-amz-date") == "" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		if !strings.Contains(r.Header.Get("Authorization"), "Credential=access/") {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	client := server.Client()
	if !probeS3API(client, server.URL+"/", Credentials{AccessKey: "access", SecretKey: "secret"}) {
		t.Fatal("expected signed S3 request to be ready")
	}
}

func TestSigV4VerificationVector(t *testing.T) {
	request, err := signedS3Request(
		"http://examplebucket.s3.amazonaws.com/",
		Credentials{AccessKey: "AKIDEXAMPLE", SecretKey: "wJalrXUtnFEMI/K7MDENG+bPxRfiCYEXAMPLEKEY"},
		time.Date(2013, 5, 24, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	const expected = "AWS4-HMAC-SHA256 Credential=AKIDEXAMPLE/20130524/us-east-1/s3/aws4_request, SignedHeaders=host;x-amz-content-sha256;x-amz-date, Signature=59d704f75576980a08333a71d37c8c7a666713ef305e7b9a6ac767f3f05dcd6d"
	if got := request.Header.Get("Authorization"); got != expected {
		t.Fatalf("authorization = %q, want %q", got, expected)
	}
}

func TestMetadataIsEffectiveAndCorruptionPropagates(t *testing.T) {
	root := t.TempDir()
	settings := DefaultSettings("docker")
	service, err := NewService(&fakeExecutor{}, root, settings)
	if err != nil {
		t.Fatal(err)
	}
	if err := service.layout.ensure(); err != nil {
		t.Fatal(err)
	}
	if err := service.saveMetadata(Credentials{AccessKey: "a", SecretKey: "b"}); err != nil {
		t.Fatal(err)
	}
	effective, err := NewService(&fakeExecutor{}, root, DefaultSettings("podman"))
	if err != nil {
		t.Fatal(err)
	}
	if effective.settings.Runtime != "docker" || effective.settings.Image != settings.Image || effective.settings.APIPort != settings.APIPort {
		t.Fatalf("metadata was not effective: %#v", effective.settings)
	}
	explicit := DefaultSettings("podman")
	explicit.APIPort = 19000
	explicit.Explicit.APIPort = true
	withOverride, err := NewService(&fakeExecutor{}, root, explicit)
	if err != nil {
		t.Fatal(err)
	}
	if withOverride.settings.APIPort != 19000 || withOverride.settings.Runtime != "docker" {
		t.Fatalf("explicit merge incorrect: %#v", withOverride.settings)
	}
	if err := os.WriteFile(service.layout.MetadataPath, []byte("{"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewService(&fakeExecutor{}, root, settings); err == nil || !strings.Contains(err.Error(), "metadata") {
		t.Fatalf("corrupt metadata error = %v", err)
	}
}

func TestExplicitCredentialsMustMatchPersistedCredentials(t *testing.T) {
	root := t.TempDir()
	service, err := NewService(&fakeExecutor{}, root, DefaultSettings("docker"))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.layout.ensure(); err != nil {
		t.Fatal(err)
	}
	persisted := Credentials{AccessKey: "persisted-access", SecretKey: "persisted-secret"}
	if err := service.saveMetadata(persisted); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(persisted)
	if err := writePrivateFile(service.layout.CredentialsPath, data); err != nil {
		t.Fatal(err)
	}
	settings := DefaultSettings("docker")
	settings.Credentials = &Credentials{AccessKey: "other-access", SecretKey: "other-secret"}
	other, err := NewService(&fakeExecutor{}, root, settings)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := other.ensureCredentials(false); err == nil || !strings.Contains(err.Error(), "destroying and recreating") {
		t.Fatalf("credential mismatch error = %v", err)
	}
}

func TestKindWorkloadProbeUsesDisposablePodAndFailsThroughExecutor(t *testing.T) {
	executor := &fakeExecutor{handler: func(opts localdev.RunOptions) ([]byte, error) {
		if len(opts.Args) >= 4 && opts.Args[2] == "wait" {
			return nil, errors.New("pod failed")
		}
		return nil, nil
	}}
	service, err := NewService(executor, t.TempDir(), DefaultSettings("docker"))
	if err != nil {
		t.Fatal(err)
	}
	credentials := Credentials{AccessKey: "access-value", SecretKey: "secret-value"}
	if err := service.probeKindWorkload(context.Background(), "/tmp/kubeconfig", "10.89.0.2", credentials); err == nil || !strings.Contains(err.Error(), "authenticated RustFS S3 probe") {
		t.Fatalf("probe error = %v", err)
	}
	if len(executor.calls) != 4 {
		t.Fatalf("calls = %#v, want apply, wait, pod cleanup, secret cleanup", executor.calls)
	}
	if executor.calls[0].Stdin == nil {
		t.Fatal("expected pod manifest on stdin")
	}
	manifest, err := io.ReadAll(executor.calls[0].Stdin)
	if err != nil {
		t.Fatal(err)
	}
	manifestText := string(manifest)
	for _, required := range []string{
		"image: " + kindProbeImage,
		"automountServiceAccountToken: false",
		"runAsNonRoot: true",
		"allowPrivilegeEscalation: false",
		"readOnlyRootFilesystem: true",
		"- ALL",
		"name: HOME\n      value: /dev",
		"name: AWS_CONFIG_FILE\n      value: /dev/null",
		"name: AWS_SHARED_CREDENTIALS_FILE\n      value: /dev/null",
		"name: AWS_CLI_HISTORY_FILE\n      value: /dev/null",
		"name: AWS_PAGER\n      value: \"\"",
		"name: AWS_EC2_METADATA_DISABLED\n      value: \"true\"",
		"secretKeyRef:",
	} {
		if !strings.Contains(manifestText, required) {
			t.Fatalf("probe manifest missing %q: %s", required, manifestText)
		}
	}
	for _, call := range executor.calls {
		joined := strings.Join(call.Args, " ")
		if strings.Contains(joined, credentials.AccessKey) || strings.Contains(joined, credentials.SecretKey) {
			t.Fatalf("probe leaked credentials in argv: %q", joined)
		}
	}
}

func TestUpNeverAutoAttachesKind(t *testing.T) {
	settings := DefaultSettings("docker")
	executor := &fakeExecutor{handler: func(opts localdev.RunOptions) ([]byte, error) {
		if len(opts.Args) >= 1 && opts.Args[0] == "inspect" {
			return nil, errors.New("no such container")
		}
		return nil, nil
	}}
	service, err := NewService(executor, t.TempDir(), settings)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := service.Up(ctx); err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("up error = %v", err)
	}
	for _, call := range executor.calls {
		if len(call.Args) >= 2 && call.Args[0] == "network" && call.Args[1] == "inspect" {
			t.Fatalf("up attempted automatic Kind attachment: %#v", executor.calls)
		}
	}
}

func TestAttachKindCannotClaimReachabilityWithoutKubeconfig(t *testing.T) {
	service, err := NewService(&fakeExecutor{}, t.TempDir(), DefaultSettings("docker"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.AttachKind(context.Background()); err == nil || !strings.Contains(err.Error(), "kubeconfig") {
		t.Fatalf("attach error = %v", err)
	}
}

func TestInspectNonNotFoundErrorPropagates(t *testing.T) {
	executor := &fakeExecutor{handler: func(opts localdev.RunOptions) ([]byte, error) {
		return nil, errors.New("permission denied by runtime")
	}}
	service, err := NewService(executor, t.TempDir(), DefaultSettings("docker"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Status(context.Background()); err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("status error = %v", err)
	}
}

func serviceContainer(service *Service, creds Credentials) *containerInspect {
	var container containerInspect
	container.Config.Image = service.settings.Image
	container.Config.User = "10001:10001"
	container.Config.Cmd = expectedCommand()
	container.Config.Labels = service.containerLabels(creds)
	container.Mounts = append(container.Mounts, struct {
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
	}{Source: service.layout.DataDir, Destination: "/data"})
	container.HostConfig.PortBindings = map[string][]struct {
		HostIP   string `json:"HostIp"`
		HostPort string `json:"HostPort"`
	}{
		"9000/tcp": {{HostIP: "127.0.0.1", HostPort: fmt.Sprint(service.settings.APIPort)}},
		"9001/tcp": {{HostIP: "127.0.0.1", HostPort: fmt.Sprint(service.settings.ConsolePort)}},
	}
	container.State.Running = false
	return &container
}

func TestStatusDoesNotSerializeSecrets(t *testing.T) {
	status := Status{CredentialsPath: "/private/credentials.json"}
	data, err := json.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "secret") || strings.Contains(string(data), "access") {
		t.Fatalf("status contains credential material: %s", data)
	}
}

func TestMetadataContainsHashesButNotCredentialValues(t *testing.T) {
	service, err := NewService(&fakeExecutor{}, t.TempDir(), DefaultSettings("docker"))
	if err != nil {
		t.Fatal(err)
	}
	if err := service.layout.ensure(); err != nil {
		t.Fatal(err)
	}
	credentials := Credentials{AccessKey: "access-value", SecretKey: "secret-value"}
	if err := service.saveMetadata(credentials); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(service.layout.MetadataPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), credentials.AccessKey) || strings.Contains(string(data), credentials.SecretKey) {
		t.Fatalf("metadata leaked credentials: %s", data)
	}
	if !strings.Contains(string(data), "credential_hash") || !strings.Contains(string(data), "spec_hash") {
		t.Fatalf("metadata missing non-secret hashes: %s", data)
	}
}

func TestParseCredentialsRejectsUnknownFields(t *testing.T) {
	if _, err := ParseCredentials(strings.NewReader(`{"access_key":"a","secret_key":"b","secret":"leak"}`)); err == nil {
		t.Fatal("expected unknown credential field rejection")
	}
}

func TestPinnedImageRejectsMutableReferences(t *testing.T) {
	for _, image := range []string{"docker.io/rustfs/rustfs:latest", "docker.io/rustfs/rustfs:1.0.0", "docker.io/rustfs/rustfs@sha256:short"} {
		if _, err := NewService(nil, t.TempDir(), Settings{Image: image}); err == nil {
			t.Fatalf("expected image %q to be rejected", image)
		}
	}
	if !pinnedImage(fmt.Sprintf("docker.io/rustfs/rustfs@sha256:%064x", 1)) {
		t.Fatal("expected valid digest reference")
	}
}
