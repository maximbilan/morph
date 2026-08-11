package category

import (
	"strings"
	"testing"
)

func TestGetCodeAsString(t *testing.T) {
	code := int32(123456)
	result := getCodeAsString(code)
	if result != "123456" {
		t.Errorf("Expected 123456, got %s", result)
	}
}

func TestPathsCoversEveryLeafInStableOrder(t *testing.T) {
	want := 0
	for _, subs := range categories {
		if len(subs) == 0 {
			want++
			continue
		}
		want += len(subs)
	}

	paths := Paths()
	if len(paths) != want {
		t.Fatalf("Paths() returned %d entries, want %d", len(paths), want)
	}
	if len(order) != len(categories) {
		t.Fatalf("order lists %d categories, taxonomy has %d", len(order), len(categories))
	}

	seen := map[string]bool{}
	for _, path := range paths {
		if seen[path] {
			t.Errorf("duplicate path %q", path)
		}
		seen[path] = true

		name, sub := SplitPath(path)
		if name+"/"+sub != path && name != path {
			t.Errorf("SplitPath(%q) = (%q, %q), which does not round-trip", path, name, sub)
		}
	}

	// A reordered enum silently breaks prompt caching.
	for i := range paths {
		if paths[i] != Paths()[i] {
			t.Fatalf("Paths() is not stable at index %d", i)
		}
	}
}

func TestSplitPath(t *testing.T) {
	tests := []struct {
		path    string
		wantCat string
		wantSub string
	}{
		{"Food/Shop", "Food", "Shop"},
		{"  Food/Shop  ", "Food", "Shop"},
		{"Waste", "Waste", ""},
		{"Other", "Other", ""},
		// A category with no subcategories never gets one.
		{"Waste/Anything", "Waste", ""},
		{"Food/Groceries", "Food", "Other"},
		{"Groceries/Shop", "Other", ""},
		{"", "Other", ""},
		{"food/shop", "Other", ""},
	}

	for _, tt := range tests {
		gotCat, gotSub := SplitPath(tt.path)
		if gotCat != tt.wantCat || gotSub != tt.wantSub {
			t.Errorf("SplitPath(%q) = (%q, %q), want (%q, %q)", tt.path, gotCat, gotSub, tt.wantCat, tt.wantSub)
		}
	}
}

func TestNormalizeDropsSubcategoryWhenCategoryHasNoOtherLeaf(t *testing.T) {
	// Every current category with subcategories has an "Other" leaf; cover the
	// branch for one that does not.
	categories["Temp"] = []string{"OnlyLeaf"}
	defer delete(categories, "Temp")

	if name, sub := Normalize("Temp", "Nonsense"); name != "Temp" || sub != "" {
		t.Errorf("Normalize(Temp, Nonsense) = (%q, %q), want (Temp, \"\")", name, sub)
	}
}

func TestClassificationPromptDescribesEveryPath(t *testing.T) {
	prompt := ClassificationPrompt()
	for _, path := range Paths() {
		if !strings.Contains(prompt, path) {
			t.Errorf("prompt is missing taxonomy path %q", path)
		}
	}
	// The old prompt's unconditional "use Other" escape hatch is what pushed small
	// models onto the Other leaves.
	if !strings.Contains(prompt, "LAST RESORT") {
		t.Error("prompt no longer tells the model that Other is a last resort")
	}
}

func TestLeafHintsReferenceRealPaths(t *testing.T) {
	valid := map[string]bool{}
	for _, path := range Paths() {
		valid[path] = true
	}
	for path := range leafHints {
		if !valid[path] {
			t.Errorf("leafHints describes %q, which is not in the taxonomy", path)
		}
	}
}

func TestHintsCoverEveryCategory(t *testing.T) {
	for name := range categories {
		if hints[name] == "" {
			t.Errorf("category %q has no hint", name)
		}
	}
}
