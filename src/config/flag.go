package config

// TriBoolFlag is a tri-state boolean flag.Value implementation: it
// distinguishes "not passed on the command line" from "explicitly passed
// as false". A plain flag.Bool cannot make that distinction, which makes it
// impossible to express `--debug=false` as an override that wins over a
// config-file or environment-variable value of true.
//
// Usage:
//
//	var debugFlag config.TriBoolFlag
//	fs.Var(&debugFlag, "debug", "Enable debug mode")
//	...
//	overrides.Debug = debugFlag.Pointer() // nil if --debug was never passed
type TriBoolFlag struct {
	set bool
	val bool
}

// String returns the current string representation, required by flag.Value.
func (t *TriBoolFlag) String() string {
	if t == nil || !t.set {
		return ""
	}
	if t.val {
		return "true"
	}
	return "false"
}

// Set parses s as a truthy/falsy value and records that the flag was
// explicitly provided. Called by the flag package for both `--debug` (with
// implicit "true") and `--debug=false` (explicit value).
func (t *TriBoolFlag) Set(s string) error {
	val, err := ParseBool(s, false)
	if err != nil {
		return err
	}
	t.val = val
	t.set = true
	return nil
}

// IsBoolFlag tells the flag package this flag may be used without an
// argument (`--debug` is equivalent to `--debug=true`).
func (t *TriBoolFlag) IsBoolFlag() bool {
	return true
}

// IsSet reports whether the flag was explicitly provided on the command line.
func (t *TriBoolFlag) IsSet() bool {
	return t.set
}

// Pointer returns a pointer to the parsed value, or nil if the flag was
// never explicitly provided. Intended for direct use as an Overrides field.
func (t *TriBoolFlag) Pointer() *bool {
	if !t.set {
		return nil
	}
	val := t.val
	return &val
}
