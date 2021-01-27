package scan

type InfrastructureNotInSync struct{}

func (i InfrastructureNotInSync) Error() string {
	return "Infrastructure is not in sync"
}
