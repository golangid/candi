package candihelper

import (
	"context"
	"math"
)

// StreamAllBatch helper func for stream data
func StreamAllBatch[T any, F FilterStreamer](ctx context.Context, totalData int, filter F, fetchAllFunc func(context.Context, F) ([]T, error), handleFunc func(idx int, data *T) error) error {
	totalPages := int(math.Ceil(float64(totalData) / float64(filter.GetLimit())))
	for filter.GetPage() <= totalPages {
		list, err := fetchAllFunc(ctx, filter)
		if err != nil {
			return err
		}
		for i, data := range list {
			offset := (filter.GetPage() - 1) * filter.GetLimit()
			if err := handleFunc(offset+i, &data); err != nil {
				return err
			}
		}
		filter.IncrPage()
	}
	return nil
}

// StreamAllBatchDynamic helper func for stream data with dynamic source changes
func StreamAllBatchDynamic[T any, F FilterStreamer](ctx context.Context, filter F, fetchAllFunc func(context.Context, F) ([]T, error), handleFunc func(idx int, data *T) error) error {
	for {
		list, err := fetchAllFunc(ctx, filter)
		if err != nil {
			return err
		}
		if len(list) == 0 {
			return nil
		}
		for i, data := range list {
			offset := (filter.GetPage() - 1) * filter.GetLimit()
			if err := handleFunc(offset+i, &data); err != nil {
				return err
			}
		}
		filter.IncrPage()
	}
}
