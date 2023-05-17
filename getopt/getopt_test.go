package getopt

import(
	"testing"
)

//Check whether bare flags passed are recognized
func TestLongFlag(t *testing.T) {
	a := NewFlag('a', "about", "topic")
	b := NewFlag('b', "before", "date prior")
	c := NewFlag('c', "config", "Configuration file")
	argv := []string { "--about", "--before" }
	ParseArgv(argv)

	if !a.Passed {
		t.Fatal("'--about' passed but not recognized")
	}
	if !b.Passed {
		t.Fatal("'--before' passed but not recognized")
	}
	if c.Passed {
		t.Fatal("'--config' not passed but flagged")
	}
	ParseArgv(argv)
}

//Check whether long opt works with equals sign
func TestLongEqualOptArg(t *testing.T) {
	f := NewOptArg('f', "file", "file to read")
	argv := []string { "--file=hello.txt" }
	ParseArgv(argv)

	if f.Opt != "hello.txt" {
		t.Fatalf("Expected 'hello.txt', got %s\n", f.Opt)
	}
}

//Check whether long opt works with separate argument
func TestLongSepOptArg(t *testing.T) {
	f := NewOptArg('f', "file", "file to read")
	argv := []string { "--file", "hello.txt" }
	ParseArgv(argv)
	if f.Opt != "hello.txt" {
		t.Fatalf("Expected 'hello.txt', got %s\n", f.Opt)
	}
}

//Check whether flags can be set with boolean strings
func TestLongFlagEqualBool(t *testing.T) {
	f := NewFlag('f', "force", "force action")
	g := NewFlag('g', "global", "global change")
	g.Passed = true
	argv := []string { "--force=True", "--global=False" }
	err := ParseArgv(argv)
	if err != nil {
		t.Fatal("Failed to parse boolean")
	}
	if !f.Passed {
		t.Fatal("Passed f, not recognized")
	}
	if g.Passed {
		t.Fatal("Negated g, still true")
	}
}

//Check that invalid strings cause an error
func TestLongFlagBoolError(t *testing.T) {
	_ = NewFlag('f', "force", "force")
	argv := []string {"--force=Fase" }
	err := ParseArgv(argv)
	if err == nil {
		t.Fatal("Did not recognize mis-spelled false")
	}
}

//Check that short flags work, including negating them
func TestShortSeparate(t *testing.T) {
	a := NewFlag('a', "about", "topic")
	b := NewFlag('b', "before", "date prior")
	b.Passed = true
	c := NewFlag('c', "config", "Configuration file")
	argv := []string{ "-a", "+b" }
	ParseArgv(argv)
	if !a.Passed {
		t.Fatal("'a' should be true as it was passed")
	}

	if b.Passed {
		t.Fatal("'b' should be false, as it was passed +b")
	}

	if c.Passed {
		t.Fatal("'c' should be false as it was not passed")
	}
}

//Check that groups of short flags are parsed
func TestShortTogether_pair(t *testing.T) {
	a := NewFlag('a', "about", "topic")
	b := NewFlag('b', "before", "date prior")
	c := NewFlag('c', "config", "Configuration file")

	argv := []string { "-ab" }
	ParseArgv(argv)
	if !a.Passed {
		t.Fatalf("'a' was passed, should be true")
	}

	if !b.Passed {
		t.Fatal("'b' was passed, should be true")
	}
	
	if c.Passed {
		t.Fatal("'c' was not passed, should be false")
	}
}

//Check if longer clump works
func TestShortTogether_triple(t *testing.T) {
	a := NewFlag('a', "about", "topic")
	b := NewFlag('b', "before", "date prior")
	c := NewFlag('c', "config", "Configuration file")

	argv := []string { "-abc" }
	ParseArgv(argv)
	if !a.Passed {
		t.Fatalf("'a' was passed, should be true")
	}

	if !b.Passed {
		t.Fatal("'b' was passed, should be true")
	}
	
	if !c.Passed {
		t.Fatal("'c' was passed, should be true")
	}
}

//Check if clump of negated flags works
func TestShortClumpNegate(t *testing.T) {
	a := NewFlag('a', "about", "topic")
	b := NewFlag('b', "before", "date prior")
	c := NewFlag('c', "config", "Configuration file")
	a.Passed = true
	b.Passed = true
	c.Passed = true
	argv := []string { "+abc" }
	ParseArgv(argv)
	if a.Passed {
		t.Fatal("A should be negated")
	}
	if b.Passed {
		t.Fatal("B should be negated")
	}
	if c.Passed {
		t.Fatal("C should be negated")
	}
}

//Check that '-' calls the assigned function
func TestStdin(t *testing.T) {
	read := false
	StdinHandler = func() error {
		read = true
		return nil
	}
	argv := []string{ "-" }
	ParseArgv(argv)
	if !read {
		t.Fatal("'-' was passed, but function not called")
	}
}

func TestOptCountShort(t *testing.T) {
	v := NewOptCount('v', "verbose", "Verbosity of the program")
	argv := []string { "-vvv" }
	ParseArgv(argv)
	if v.Count != 3 {
		t.Fatalf("Expected verbosity of 3, got %d\n", v.Count)
	}
}

func TestOptCountLong(t *testing.T) {
	v := NewOptCount('v', "verbose", "Verbosity of the program")
	argv := []string { "--verbose", "--verbose", "--verbose" }
	ParseArgv(argv)
	if v.Count != 3 {
		t.Fatalf("Expected verbosity of 3, got %d\n", v.Count)
	}
}

func TestOptCountEquals(t *testing.T) {
	v := NewOptCount('v', "verbose", "Verbosity of the program")
	argv := []string { "--verbose=3" }
	ParseArgv(argv)
	if v.Count != 3 {
		t.Fatalf("Expected verbosity of 3, got %d\n", v.Count)
	}
}

//Test that opt that takes arg works when connected as short opt
func TestShortOptArgConn(t *testing.T) {
	f := NewOptArg('f', "file", "file to process")
	argv := []string { "-fhello.txt" }
	ParseArgv(argv)
	if f.Opt != "hello.txt" {
		t.Fatalf("Expected hello.txt, got %s", f.Opt)
	}
}

//Test that arguments passed between options get added to rest
func TestRest(t *testing.T) {
	f := NewFlag('f', "file", "file to process")
	argv := []string { "hello", "--file", "world" }
	Rest = make([]string, initialCapacity)
	ParseArgv(argv)
	if !f.Passed {
		t.Fatal("Expected f, not passed")
	}
	if len(Rest) != 2 {
		t.Fatalf("Expected 2 in Rest, got %d", len(Rest))
	}
	if Rest[0] != "hello" {
		t.Fatalf("Expected 'hello', got %s", Rest[0])
	}
	if Rest[1] != "world" {
		t.Fatalf("Expected 'world', got %s", Rest[1])
	}
}

//Test the -- passes everything to Rest
func TestAllRest(t *testing.T) {
	f := NewFlag('f', "file", "file to process")
	argv := []string { "hello", "--", "--file", "-f", "world" }
	Rest = make([]string, initialCapacity)
	ParseArgv(argv)
	if f.Passed {
		t.Fatal("Expected not f, got true")
	}
	if Rest[0] != "hello" {
		t.Fatalf("Expected 'hello', got %s", Rest[0])
	}
	if Rest[1] != "--file" {
		t.Fatalf("Expected '--file', got %s", Rest[1])
	}
}
