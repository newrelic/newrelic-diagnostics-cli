package fixtures

type BadTaskFile struct {
}

func (p BadTaskFile) Dependencies() []string {
	return []string{
		"I/Am/Dependency1",
		"I/Am/Dependency2",
	}
}
