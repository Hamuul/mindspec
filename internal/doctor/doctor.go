package doctor

// Status represents the result of a single health check.
type Status int

const (
	OK      Status = iota
	Missing        // expected artifact is absent
	Error          // something is wrong and needs action
	Warn           // advisory, not a failure
)

// Check represents a single health check result.
type Check struct {
	Name    string
	Status  Status
	Message string
}

// Report holds the results of all doctor checks.
type Report struct {
	Checks []Check
}

// HasFailures returns true if any check has Error or Missing status.
func (r *Report) HasFailures() bool {
	for _, c := range r.Checks {
		if c.Status == Error || c.Status == Missing {
			return true
		}
	}
	return false
}

// Run executes all doctor checks against the given project root.
func Run(root string) *Report {
	r := &Report{}
	checkDocs(r, root)
	checkBeads(r, root)
	return r
}
