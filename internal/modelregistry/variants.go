package modelregistry

// fastVariantPairs is the single source of truth for the base<->fast model
// mapping. Each entry links a base model to a genuinely faster same-family
// variant (usually lower-latency, sometimes lower-quality). This is data, not
// logic: adding a pair is a one-line change here — never a code change in the
// /fast command. Both directions are derived from this one table (FastVariant
// walks base->fast, BaseVariant walks fast->base), which is why a model's
// "fast mode" state is encoded by its identity and needs no separate flag.
//
// Only pairs with an unambiguous 1:1 base<->fast relationship are listed. A model
// with no clean counterpart (e.g. a top-tier Opus, where both Sonnet and Haiku
// would be candidates) is intentionally omitted, so /fast reports "no fast mode
// available" for it rather than guessing.
//
// Invariants (guarded by tests): every id resolves in the default registry, and
// no id appears as both a base and a fast, so the two directions are unambiguous.
var fastVariantPairs = []struct {
	base string
	fast string
}{
	{base: "claude-sonnet-4.5", fast: "claude-haiku-4.5"},
	{base: "gpt-4.1", fast: "gpt-4.1-mini"},
	{base: "gpt-4o", fast: "gpt-4o-mini"},
	{base: "gemini-2.5-pro", fast: "gemini-2.5-flash"},
}

// FastVariant resolves id to its faster same-family counterpart. It resolves id
// to a concrete entry, finds the pair whose base matches, resolves the fast id,
// and returns it only when the target is available and not deprecated. Returns
// (_, false) when id is unknown, has no fast variant, or the variant resolves to
// an unavailable/deprecated model. Mirrors Registry.UpgradeTarget.
func (registry Registry) FastVariant(id string) (ModelEntry, bool) {
	return registry.resolveVariant(id, true)
}

// BaseVariant resolves id to the base model that id is the fast variant of. Same
// resolution and availability rules as FastVariant, in the reverse direction. A
// non-empty result therefore means id IS a fast model — this is exactly how the
// /fast command derives "currently in fast mode" without storing any flag.
func (registry Registry) BaseVariant(id string) (ModelEntry, bool) {
	return registry.resolveVariant(id, false)
}

// resolveVariant walks fastVariantPairs in the requested direction. It resolves
// both the source and each candidate to canonical entries before comparing, so
// aliases/patterns on either side (e.g. "claude-sonnet-4.5" vs a dated id) match
// correctly.
func (registry Registry) resolveVariant(id string, wantFast bool) (ModelEntry, bool) {
	source, ok := registry.Resolve(id)
	if !ok {
		return ModelEntry{}, false
	}
	for _, pair := range fastVariantPairs {
		fromID, toID := pair.base, pair.fast
		if !wantFast {
			fromID, toID = pair.fast, pair.base
		}
		from, ok := registry.Resolve(fromID)
		if !ok || from.ID != source.ID {
			continue
		}
		target, ok := registry.Resolve(toID)
		if !ok || target.Status == ModelStatusDeprecated {
			return ModelEntry{}, false
		}
		return target, true
	}
	return ModelEntry{}, false
}
