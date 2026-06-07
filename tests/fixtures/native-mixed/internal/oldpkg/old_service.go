package oldpkg

type OldService struct{}

func (service OldService) Build(worker OldWorker) OldWorker {
	return OldWorker{}
}
