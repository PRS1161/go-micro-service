package helpers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/PRS1161/go-micro-service/model"
	"github.com/redis/go-redis/v9"
)

type RedisHelper struct {
	Client *redis.Client
}

func OrderKey(id uint64) string {
	return fmt.Sprintf("ORDER ID: %d", id)
}

func (r *RedisHelper) Insert(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("FAILED TO ENCODE ORDER %w", err)
	}

	key := OrderKey(order.Id)

	txn := r.Client.Pipeline()

	res := txn.SetNX(ctx, key, string(data), 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("FAILED TO INSERT KEY %w", err)
	}

	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("FAILED TO INSERT ORDER IN SET %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("FAILED TO EXECUTE PIPELINE %w", err)
	}

	return nil
}

func (r *RedisHelper) GetByKey(ctx context.Context, id uint64) (model.Order, error) {

	key := OrderKey(id)

	val, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, errors.New("ORDER DOES NOT EXIST")
	} else if err != nil {
		return model.Order{}, fmt.Errorf("FAILED TO GET KEY %w", err)
	}
	var order model.Order
	err = json.Unmarshal([]byte(val), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("FAILED TO DECODE ORDER %w", err)
	}

	return order, nil
}

func (r *RedisHelper) DeleteByKey(ctx context.Context, id uint64) error {
	key := OrderKey(id)

	txn := r.Client.Pipeline()

	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return errors.New("ORDER DOES NOT EXIST")
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("FAILED TO DELETE KEY %w", err)
	}

	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("FAILED TO REMOVE ORDER IN SET %w", err)
	}

	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("FAILED TO EXECUTE PIPELINE %w", err)
	}

	return nil
}

func (r *RedisHelper) Update(ctx context.Context, order model.Order) error {
	data, err := json.Marshal(order)

	if err != nil {
		return fmt.Errorf("FAILED TO ENCODE ORDER %w", err)
	}

	key := OrderKey(order.Id)

	err = r.Client.SetXX(ctx, key, string(data), 0).Err()
	if errors.Is(err, redis.Nil) {
		return errors.New("ORDER DOES NOT EXIST")
	} else if err != nil {
		return fmt.Errorf("FAILED TO UPDATE KEY %w", err)
	}
	return nil
}

func (r *RedisHelper) GetAllKeys(ctx context.Context, page Pagination) (Result, error) {

	res := r.Client.SScan(ctx, "orders", uint64(page.Cursor), "*", int64(page.Limit))

	keys, cursor, err := res.Result()
	if err != nil {
		return Result{}, fmt.Errorf("FAILED TO GET ALL KEYS %w", err)
	}

	if len(keys) == 0 {
		return Result{
			Orders: []model.Order{},
			Cursor: uint(cursor),
			Limit:  page.Limit,
		}, nil
	}

	xs, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return Result{}, fmt.Errorf("FAILED TO GET ALL ORDERS %w", err)
	}

	orders := make([]model.Order, len(xs))

	for i, x := range xs {
		x := x.(string)
		var order model.Order
		err = json.Unmarshal([]byte(x), &order)
		if err != nil {
			return Result{}, fmt.Errorf("FAILED TO DECODE ORDER %w", err)
		}

		orders[i] = order
	}

	return Result{
		Orders: orders,
		Cursor: uint(cursor),
		Limit:  page.Limit,
	}, nil
}

type Pagination struct {
	Limit  uint
	Cursor uint
}

type Result struct {
	Orders []model.Order
	Limit  uint
	Cursor uint
}
