package gradle

import (
	"strings"
	"testing"
)

func TestInitScriptAvoidsVarargDetachedConfiguration(t *testing.T) {
	script := InitScript()
	if strings.Contains(script, "detachedConfiguration(*") {
		t.Fatalf("init script should not use vararg detachedConfiguration; it breaks with large dep sets")
	}
	if !strings.Contains(script, "detachedConfiguration()") {
		t.Fatalf("init script should create detachedConfiguration without varargs")
	}
}
