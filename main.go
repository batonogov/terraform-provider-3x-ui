package main

import (
	"context"
	"log"

	"github.com/batonogov/terraform-provider-3x-ui/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	if err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/batonogov/threexui",
	}); err != nil {
		log.Fatal(err)
	}
}
