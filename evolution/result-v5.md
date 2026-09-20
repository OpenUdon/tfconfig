# Result V5 - Bounded Static Instance Facts

P01 adds the minimum public fact delta demonstrated by Ramen's semantic-loss
gate. Collection values now preserve object/map/tuple/list/set identity, and a
top-level `toset` becomes a static set only when its argument is already wholly
known. Numeric count, collection keys/values, module inputs, provider mappings,
full module addresses, source references, and ranges remain deterministic.

Variable-backed `toset`, locals, conditionals, unknown functions, and other
runtime-dependent expressions remain symbolic. No general evaluator,
filesystem/environment/time/random function, module download, provider,
backend, state, network, plan, apply, or Terraform/OpenTofu execution is added.
