
default: install

.PHONY: install testacc
# Install to use dev_overrides in provider_installation of Terraform
install:
	go install
# Run acceptance tests
testacc:
	cd bastion ; TF_ACC=1 go test -v --timeout 0 -coverprofile=../coverage.out $(TESTARGS)
	go tool cover -html=coverage.out
