# Prompt V4 - Parallel-Lane Memory-Bank Harness

Migrate the private `tfconfig` harness to the lane-aware memory-bank contract
used by the current unattended runner. Preserve completed parser, mirror, and
historical downstream-coordination records while giving future static-parser
and upstream-mirror work separate ownership lanes.

Use permanent zero-padded IDs, parser-compatible state tables, an explicit
lane map, cross-lane dependency rules, and unnumbered candidate directions.
Do not alter the public parser boundary or backfill the unused M01 ID.
