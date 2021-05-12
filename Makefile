
OS ?= $(shell go env GOOS)
ARCH ?= $(shell go env GOARCH)
KUBEBUILDER_VERSION=2.3.2


test: _test/kubebuilder
	go test -v .

_test/kubebuilder:
	curl -fsSL https://github.com/kubernetes-sigs/kubebuilder/releases/download/v$(KUBEBUILDER_VERSION)/kubebuilder_$(KUBEBUILDER_VERSION)_$(OS)_$(ARCH).tar.gz -o kubebuilder-tools.tar.gz
	mkdir -p _test/kubebuilder
	tar -xvf kubebuilder-tools.tar.gz
	mv kubebuilder_$(KUBEBUILDER_VERSION)_$(OS)_$(ARCH)/bin _test/kubebuilder/
	rm kubebuilder-tools.tar.gz
	rm -R kubebuilder_$(KUBEBUILDER_VERSION)_$(OS)_$(ARCH)


clean: clean-kubebuilder

clean-kubebuilder:
	rm -Rf _test/kubebuilder


IMAGE_NAME = "cert-manager-webhook-hetzner"
IMAGE_TAG  = "latest"

build:
	DOCKER_BUILDKIT=1 docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .



