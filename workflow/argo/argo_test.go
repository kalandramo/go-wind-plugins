package argo

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient(ClientOptions{})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
	if client.options.ServerURL != defaultServerURL {
		t.Errorf("expected default server URL %s, got %s", defaultServerURL, client.options.ServerURL)
	}
	if client.options.Namespace != defaultNamespace {
		t.Errorf("expected default namespace %s, got %s", defaultNamespace, client.options.Namespace)
	}
	_ = client.Close()
}

func TestNewClientWithCustomOptions(t *testing.T) {
	client, err := NewClient(ClientOptions{
		ServerURL:          "https://argo.example.com:2746",
		Namespace:          "production",
		Token:              "test-token",
		InsecureSkipVerify: true,
	})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
	if client.options.ServerURL != "https://argo.example.com:2746" {
		t.Errorf("expected custom server URL, got %s", client.options.ServerURL)
	}
	if client.options.Namespace != "production" {
		t.Errorf("expected production namespace, got %s", client.options.Namespace)
	}
	if client.options.Token != "test-token" {
		t.Error("expected token to be set")
	}
	_ = client.Close()
}

func TestPhaseIsTerminal(t *testing.T) {
	tests := []struct {
		phase    Phase
		terminal bool
	}{
		{PhasePending, false},
		{PhaseRunning, false},
		{PhaseSucceeded, true},
		{PhaseFailed, true},
		{PhaseError, true},
	}

	for _, tt := range tests {
		if got := tt.phase.IsTerminal(); got != tt.terminal {
			t.Errorf("Phase %s: IsTerminal() = %v, want %v", tt.phase, got, tt.terminal)
		}
	}
}

func TestNamespaceOverride(t *testing.T) {
	client, err := NewClient(ClientOptions{Namespace: "default"})
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	defer func() { _ = client.Close() }()

	// Default namespace
	if ns := client.namespace(""); ns != "default" {
		t.Errorf("expected default namespace, got %s", ns)
	}

	// Override
	if ns := client.namespace("custom"); ns != "custom" {
		t.Errorf("expected custom namespace, got %s", ns)
	}
}

func TestWorkflowTypes(t *testing.T) {
	wf := &Workflow{
		APIVersion: "argoproj.io/v1alpha1",
		Kind:       "Workflow",
		Metadata: ObjectMeta{
			GenerateName: "hello-world-",
			Namespace:    "default",
		},
		Spec: WorkflowSpec{
			Entrypoint: "whalesay",
			Templates: []Template{
				{
					Name: "whalesay",
					Container: &Container{
						Image:   "docker/whalesay:latest",
						Command: []string{"cowsay"},
						Args:    []string{"Hello World"},
					},
				},
			},
		},
	}

	if wf.APIVersion != "argoproj.io/v1alpha1" {
		t.Errorf("unexpected APIVersion: %s", wf.APIVersion)
	}
	if wf.Spec.Entrypoint != "whalesay" {
		t.Errorf("unexpected entrypoint: %s", wf.Spec.Entrypoint)
	}
	if len(wf.Spec.Templates) != 1 {
		t.Errorf("expected 1 template, got %d", len(wf.Spec.Templates))
	}
	if wf.Spec.Templates[0].Container.Image != "docker/whalesay:latest" {
		t.Errorf("unexpected container image: %s", wf.Spec.Templates[0].Container.Image)
	}
}
