package modelregistry

import "testing"

func testRegistry(t *testing.T) Registry {
	t.Helper()
	registry, err := DefaultRegistry()
	if err != nil {
		t.Fatalf("DefaultRegistry() error = %v", err)
	}
	return registry
}

// Every id declared in the mapping must resolve in the default registry — a typo
// (or a model removed from the catalog) fails loudly here rather than silently
// making /fast a no-op.
func TestFastVariantPairsResolve(t *testing.T) {
	registry := testRegistry(t)
	for _, pair := range fastVariantPairs {
		for _, id := range []string{pair.base, pair.fast} {
			if _, ok := registry.Resolve(id); !ok {
				t.Errorf("fastVariantPairs references %q, which does not resolve in the default registry", id)
			}
		}
	}
}

// No model may be both a base and a fast, or the two directions become
// ambiguous (a model would be "currently fast" AND have a fast variant).
func TestFastVariantPairsUnambiguous(t *testing.T) {
	registry := testRegistry(t)
	role := map[string]string{} // canonical id -> "base" | "fast"
	for _, pair := range fastVariantPairs {
		for id, r := range map[string]string{pair.base: "base", pair.fast: "fast"} {
			entry, ok := registry.Resolve(id)
			if !ok {
				continue
			}
			if prev, seen := role[entry.ID]; seen && prev != r {
				t.Fatalf("model %q is declared as both a base and a fast variant", entry.ID)
			}
			role[entry.ID] = r
		}
	}
}

func TestFastVariantAndBaseVariant(t *testing.T) {
	registry := testRegistry(t)

	for _, pair := range fastVariantPairs {
		base, fast := pair.base, pair.fast

		// base -> fast (forward), and the fast id has NO fast variant of its own.
		if got, ok := registry.FastVariant(base); !ok {
			t.Errorf("FastVariant(%q) = not found, want the fast variant", base)
		} else if canon, _ := registry.Resolve(fast); got.ID != canon.ID {
			t.Errorf("FastVariant(%q).ID = %q, want %q", base, got.ID, canon.ID)
		}
		if _, ok := registry.FastVariant(fast); ok {
			t.Errorf("FastVariant(%q) resolved, but a fast model should have no fast variant", fast)
		}

		// fast -> base (reverse), and the base id is NOT itself a fast variant.
		if got, ok := registry.BaseVariant(fast); !ok {
			t.Errorf("BaseVariant(%q) = not found, want the base variant", fast)
		} else if canon, _ := registry.Resolve(base); got.ID != canon.ID {
			t.Errorf("BaseVariant(%q).ID = %q, want %q", fast, got.ID, canon.ID)
		}
		if _, ok := registry.BaseVariant(base); ok {
			t.Errorf("BaseVariant(%q) resolved, but a base model is not a fast variant", base)
		}
	}
}

func TestFastVariantAbsentAndUnknown(t *testing.T) {
	registry := testRegistry(t)

	// A real model with no configured fast/base variant reports none in both
	// directions (a valid, handled case — /fast will say "no fast mode available").
	const noVariant = "claude-opus-4.1"
	if _, ok := registry.Resolve(noVariant); !ok {
		t.Skipf("%q not in registry; skipping no-variant case", noVariant)
	}
	if _, ok := registry.FastVariant(noVariant); ok {
		t.Errorf("FastVariant(%q) resolved, want none", noVariant)
	}
	if _, ok := registry.BaseVariant(noVariant); ok {
		t.Errorf("BaseVariant(%q) resolved, want none", noVariant)
	}

	// An unknown id resolves to nothing in either direction.
	if _, ok := registry.FastVariant("definitely-not-a-real-model"); ok {
		t.Error("FastVariant(unknown) resolved, want none")
	}
	if _, ok := registry.BaseVariant("definitely-not-a-real-model"); ok {
		t.Error("BaseVariant(unknown) resolved, want none")
	}
}
