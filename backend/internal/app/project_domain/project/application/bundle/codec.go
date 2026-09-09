package bundle

import (
	"fmt"

	bundlecodec "judge-backend/internal/app/shared/types/bundle_codec"
)

// Marshal and Unmarshal delegate to the shared bundle rules; see that package
// for why the output is deterministic and the input is strict.

func Marshal(p *Project) ([]byte, error) {
	return bundlecodec.Marshal(p)
}

func Unmarshal(data []byte) (*Project, error) {
	var p Project
	if err := bundlecodec.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("%s 無法解析：%w", ProjectFile, err)
	}
	return &p, nil
}

func MarshalMeta(m *Meta) ([]byte, error) {
	return bundlecodec.Marshal(m)
}
