package consumer

import "example.com/project/internal/oldpkg"

func Build() oldpkg.OldService {
	return oldpkg.OldService{}
}
