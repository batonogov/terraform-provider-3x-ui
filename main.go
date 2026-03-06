package main

import (
	"context"
	"log"

	"github.com/batonogov/terraform-provider-threexui/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version string = "dev"

func main() {
	if err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/batonogov/threexui",
	}); err != nil {
		log.Fatal(err)
	}
}
