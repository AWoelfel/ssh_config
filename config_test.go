package ssh_config

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func loadFile(t *testing.T, filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

var files = []string{"testdata/config1", "testdata/config2"}

func TestDecode(t *testing.T) {
	for _, filename := range files {
		data := loadFile(t, filename)
		cfg, err := Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatal(err)
		}
		out := cfg.String()
		if out != string(data) {
			t.Errorf("out != data: out: %q\ndata: %q", out, string(data))
		}
	}
}

func testConfigFinder(filename string) func() string {
	return func() string { return filename }
}

func nullConfigFinder() string {
	return ""
}

func TestGet(t *testing.T) {
	us := &UserSettings{
		userConfigFinder: testConfigFinder("testdata/config1"),
	}

	val := us.Get("wap", "User")
	if val != "root" {
		t.Errorf("expected to find User root, got %q", val)
	}
}

func TestGetWildcard(t *testing.T) {
	us := &UserSettings{
		userConfigFinder: testConfigFinder("testdata/config3"),
	}

	val := us.Get("bastion.stage.i.us.example.net", "Port")
	if val != "22" {
		t.Errorf("expected to find Port 22, got %q", val)
	}

	val = us.Get("bastion.net", "Port")
	if val != "25" {
		t.Errorf("expected to find Port 24, got %q", val)
	}

	val = us.Get("10.2.3.4", "Port")
	if val != "23" {
		t.Errorf("expected to find Port 23, got %q", val)
	}
	val = us.Get("101.2.3.4", "Port")
	if val != "25" {
		t.Errorf("expected to find Port 24, got %q", val)
	}
	val = us.Get("20.20.20.4", "Port")
	if val != "24" {
		t.Errorf("expected to find Port 24, got %q", val)
	}
	val = us.Get("20.20.20.20", "Port")
	if val != "25" {
		t.Errorf("expected to find Port 25, got %q", val)
	}
}

func TestGetExtraSpaces(t *testing.T) {
	us := &UserSettings{
		userConfigFinder: testConfigFinder("testdata/extraspace"),
	}

	val := us.Get("test.test", "Port")
	if val != "1234" {
		t.Errorf("expected to find Port 1234, got %q", val)
	}
}

func TestGetCaseInsensitive(t *testing.T) {
	us := &UserSettings{
		userConfigFinder: testConfigFinder("testdata/config1"),
	}

	val := us.Get("wap", "uSER")
	if val != "root" {
		t.Errorf("expected to find User root, got %q", val)
	}
}

func TestGetEmpty(t *testing.T) {
	us := &UserSettings{
		userConfigFinder:   nullConfigFinder,
		systemConfigFinder: nullConfigFinder,
	}
	val, err := us.GetStrict("wap", "User")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if val != "" {
		t.Errorf("expected to get empty string, got %q", val)
	}
}

func TestGetEqsign(t *testing.T) {
	us := &UserSettings{
		userConfigFinder: testConfigFinder("testdata/eqsign"),
	}

	val := us.Get("test.test", "Port")
	if val != "1234" {
		t.Errorf("expected to find Port 1234, got %q", val)
	}
	val = us.Get("test.test", "Port2")
	if val != "5678" {
		t.Errorf("expected to find Port2 5678, got %q", val)
	}
}

func TestInclude(t *testing.T) {
	t.Skip("can't currently handle Include directives")
	us := &UserSettings{
		userConfigFinder: testConfigFinder("testdata/include"),
	}
	val := us.Get("test.test", "Compression")
	if val != "yes" {
		t.Errorf("expected to find Compression=yes in included file, got %q", val)
	}
}

var matchTests = []struct {
	in    []string
	alias string
	want  bool
}{
	{[]string{"*"}, "any.test", true},
	{[]string{"a", "b", "*", "c"}, "any.test", true},
	{[]string{"a", "b", "c"}, "any.test", false},
	{[]string{"any.test"}, "any1test", false},
	{[]string{"192.168.0.?"}, "192.168.0.1", true},
	{[]string{"192.168.0.?"}, "192.168.0.10", false},
	{[]string{"*.co.uk"}, "bbc.co.uk", true},
	{[]string{"*.co.uk"}, "subdomain.bbc.co.uk", true},
	{[]string{"*.*.co.uk"}, "bbc.co.uk", false},
	{[]string{"*.*.co.uk"}, "subdomain.bbc.co.uk", true},
}

func TestMatches(t *testing.T) {
	for _, tt := range matchTests {
		patterns := make([]*Pattern, len(tt.in), len(tt.in))
		for i := range tt.in {
			pat, err := NewPattern(tt.in[i])
			if err != nil {
				t.Fatalf("error compiling pattern %s: %v", tt.in[i], err)
			}
			patterns[i] = pat
		}
		host := &Host{
			Patterns: patterns,
		}
		got := host.Matches(tt.alias)
		if got != tt.want {
			t.Errorf("host(%q).Matches(%q): got %v, want %v", tt.in, tt.alias, got, tt.want)
		}
	}
}