package module

import (
	"strings"
	"testing"
)

var legalTypes = []int8{
	TYPE_DOWNLOADER,
	TYPE_ANALYZER,
	TYPE_PIPELINE,
}

var illegalTypes = []int8{
	10,
}

func TestTypeCheck(t *testing.T) {
	if IsMatch(TYPE_DOWNLOADER, nil) {
		t.Fatal("The module is nil, but do not be detected!")
	}
	for _, mt := range legalTypes {
		matchedModule := defaultFakeModuleMap[mt]
		for _, m := range fakeModules {
			if m.ID() == matchedModule.ID() {
				if !IsMatch(mt, m) {
					t.Fatalf("Inconsistent module type: expected: %T, actual: %T",
						matchedModule, mt)
				}
			} else {
				if IsMatch(mt, m) {
					t.Fatalf("The module type %T is not matched, but do not be detected!",
						mt)
				}
			}
		}
	}
}

func TestTypeLegal(t *testing.T) {
	for _, mt := range legalTypes {
		if !LegalType(mt) {
			t.Fatalf("Illegal predefined module type %q!", mt)
		}
	}
	for _, mt := range illegalTypes {
		if LegalType(mt) {
			t.Fatalf("The module type %q should not be legal!", mt)
		}
	}
}

func TestTypeGet(t *testing.T) {
	for _, mid := range legalMIDs {
		mt, err := GetType(mid)
		if err != nil {
			t.Fatalf("Couldn't get type via MID %q! Error: %s", mid, err)
		}
		expectedType := legalLetterTypeMap[strings.ToUpper(string(mid)[:1])]
		if mt != expectedType {
			t.Fatalf("Inconsistent module type for letter: expected: %d, actual: %d (MID: %s)",
				expectedType, mt, mid)
		}
	}
	for _, illegalMID := range illegalMIDs {
		_, err := GetType(illegalMID)
		if err == nil {
			t.Fatalf("It still can get type from illegal MID %q!", illegalMID)
		}
	}
}

func TestTypeGetLetter(t *testing.T) {
	for letter, mt := range legalLetterTypeMap {
		letter1, ok := type2Letter(mt)
		if !ok {
			t.Fatalf("Couldn't get letter via type %q!", mt)
		}
		if letter1 != letter {
			t.Fatalf("Inconsistent module type etter: expected: %s, actual: %s (type: %s)",
				letter, letter1, mt)
		}
	}
	for _, mt := range illegalTypes {
		_, ok := type2Letter(mt)
		if ok {
			t.Fatalf("It still can get letter from illegal type %q!", mt)
		}
	}
}

func TestTypeToLetter(t *testing.T) {
	for _, mt := range legalTypes {
		letter, ok := type2Letter(mt)
		if !ok {
			t.Fatalf("Couldn't convert module type %q to letter!", mt)
		}
		expectedLetter := legalTypeLetterMap[mt]
		if letter != expectedLetter {
			t.Fatalf("Inconsistent letter for module type: expected: %s, actual: %s (moduleType: %d)",
				expectedLetter, letter, mt)
		}
	}
	for _, mt := range illegalTypes {
		letter, ok := type2Letter(mt)
		if ok {
			t.Fatalf("It still can convert illegal module type %q to letter %q!",
				mt, letter)
		}
	}
}

func TestTypeletterToType(t *testing.T) {
	letters := []string{"D", "A", "P", "M"}
	for _, letter := range letters {
		mt, ok := letter2Type(letter)
		expectedType, legal := legalLetterTypeMap[letter]
		if legal {
			if !ok {
				t.Fatalf("Couldn't convert letter %q to module type!", letter)
			}
			if mt != expectedType {
				t.Fatalf("Inconsistent module type for letter: expected: %s, actual: %s (letter: %s)",
					expectedType, mt, letter)
			}
		} else {
			if ok {
				t.Fatalf("It still can convert illegal letter %q to module type %q!",
					letter, mt)
			}
		}
	}
}
