package provider

import (
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

// singletonListNestedBlock constrains a list-backed nested block that models
// one JSON object. Without the validator, additional blocks would be accepted
// even though the expand functions can only serialize the first element.
func singletonListNestedBlock(block schema.ListNestedBlock) schema.ListNestedBlock {
	block.Validators = append(block.Validators, listvalidator.SizeAtMost(1))
	return block
}
