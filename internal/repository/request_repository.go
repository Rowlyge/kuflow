package repository

import (
	"context"

	"github.com/Rowlyge/kuflow/internal/model"
)

type RequestRepository interface {
	Create(ctx context.Context, request *model.Request) error
}
