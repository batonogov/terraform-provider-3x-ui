package main

import (
	"context"
	"log"

	"github.com/batonogov/terraform-provider-3x-ui/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version string = "dev"

func main() {
	if err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.opentofu.org/batonogov/threexui",
	}); err != nil {
		log.Fatal(err)
	}
}
