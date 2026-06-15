// Module github-cache is a CI-only tool (see issue #285). It is deliberately
// kept in a separate module so it is never part of the published provider and
// does not affect its dependency graph.
module github-cache

go 1.26
