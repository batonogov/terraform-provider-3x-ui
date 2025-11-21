package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/batonogov/terraform-provider-3x-ui/internal/provider"
)

var (
	// version is replaced by goreleaser during the release process.
	version = "dev"
)

func main() {
	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/batonogov/3x-ui",
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
