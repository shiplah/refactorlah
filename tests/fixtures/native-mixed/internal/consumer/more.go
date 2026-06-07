package consumer

import "example.com/project/internal/oldpkg"

func Worker() oldpkg.OldWorker {
	return oldpkg.BuildWorker()
}
