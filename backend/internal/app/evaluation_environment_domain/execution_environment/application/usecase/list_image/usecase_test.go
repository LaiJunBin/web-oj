package listimage_test

import (
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/service/query/image_query/models/imageinfo"
	listimage "judge-backend/internal/app/evaluation_environment_domain/execution_environment/application/usecase/list_image"
	"judge-backend/internal/app/evaluation_environment_domain/execution_environment/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type ListImageUsecaseMocks struct {
	service *mocks.ImageQueryService
}

func NewListImageMocks(t *testing.T) *ListImageUsecaseMocks {
	return &ListImageUsecaseMocks{
		service: mocks.NewImageQueryService(t),
	}
}

func NewListImageUsecase(mocks *ListImageUsecaseMocks) listimage.Usecase {
	return listimage.NewUsecase(mocks.service)
}

func (mocks *ListImageUsecaseMocks) AssertExpectations(t *testing.T) {
	mocks.service.AssertExpectations(t)
}

func TestListImageUsecase(t *testing.T) {
	t.Run("have images", func(t *testing.T) {
		mockImageInfos := []imageinfo.ImageInfo{
			{
				Tag:         "nginx:latest",
				Description: "description1",
				IsReady:     true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				Tag:         "nginx:1.19",
				Description: "description2",
				IsReady:     false,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		}

		mocks := NewListImageMocks(t)
		mocks.service.On("GetImages").Return(mockImageInfos, nil).Once()
		input := &listimage.Input{}
		uc := NewListImageUsecase(mocks)

		output, err := uc.Execute(input)

		if err != nil {
			t.Fatalf("Execute should not return an error, but got %v", err)
		}

		assert.Equal(t, mockImageInfos, output.Images)
	})

	t.Run("no images", func(t *testing.T) {
		mocks := NewListImageMocks(t)
		input := &listimage.Input{}
		uc := NewListImageUsecase(mocks)
		mocks.service.On("GetImages").Return([]imageinfo.ImageInfo{}, nil).Once()

		output, err := uc.Execute(input)

		if err != nil {
			t.Fatalf("Execute should not return an error, but got %v", err)
		}

		if len(output.Images) != 0 {
			t.Fatalf("Expected 0 images, got %d", len(output.Images))
		}
	})

	t.Run("error", func(t *testing.T) {
		mocks := NewListImageMocks(t)
		input := &listimage.Input{}
		uc := NewListImageUsecase(mocks)
		mocks.service.On("GetImages").Return(nil, assert.AnError).Once()

		output, err := uc.Execute(input)

		if err == nil {
			t.Fatalf("Execute should return an error, but got nil")
		}

		if output != nil {
			t.Fatalf("Expected output to be nil, got %v", output)
		}
	})
}
