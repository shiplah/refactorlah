package twig

type PathRoot struct {
	Path      string
	Namespace string
}

type PathConfiguration struct {
	Roots []PathRoot
}
