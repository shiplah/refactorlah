package oldpkg_test

import "example.com/project/internal/oldpkg"

func TestService() {
	_ = oldpkg.OldService{}
	_ = oldpkg.OldWorker{}
}
