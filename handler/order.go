package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/PRS1161/go-micro-service/helpers"
	"github.com/PRS1161/go-micro-service/model"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Order struct {
	Helper *helpers.RedisHelper
}

func (o *Order) GenerateOrder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserId uuid.UUID    `json:"user_id"`
		Items  []model.Item `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()

	order := model.Order{
		Id:        rand.Uint64(),
		UserId:    body.UserId,
		Status:    model.Accpeted,
		Items:     body.Items,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	err := o.Helper.Insert(r.Context(), order)
	if err != nil {
		fmt.Println("SOMETHING WENT WRONG WHILE GENERATE ORDER", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	res, err := json.Marshal(order)
	if err != nil {
		fmt.Println("SOMETHING WENT WRONG WHILE PARSING ORDER", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(res)
	w.WriteHeader(http.StatusCreated)
}

func (o *Order) GetOrders(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	cursor, err := strconv.ParseInt(cursorStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	const size = 50
	res, err := o.Helper.GetAllKeys(r.Context(), helpers.Pagination{Cursor: uint(cursor), Limit: size})
	if err != nil {
		fmt.Println("SOMETHING WENT WRONG WHILE GETTING ORDER LIST", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var Response struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}

	Response.Items = res.Orders
	Response.Next = uint64(res.Cursor)

	data, err := json.Marshal(Response)
	if err != nil {
		fmt.Println("SOMETHING WENT WRONG WHILE PARSING ORDER LIST", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(data)
}

func (o *Order) GetSingleOrder(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	orderId, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	order, err := o.Helper.GetByKey(r.Context(), uint64(orderId))
	if errors.Is(err, errors.New("ORDER DOES NOT EXIST")) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("SOMETHING WENT WRONG WHILE GETTING ORDER ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(order); err != nil {
		fmt.Println("SOMETHING WENT WRONG WHILE PARSING ORDER", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	idParam := chi.URLParam(r, "id")

	orderId, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	theOrder, err := o.Helper.GetByKey(r.Context(), orderId)
	if errors.Is(err, errors.New("ORDER DOES NOT EXIST")) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("SOMETHING WENT WRONG WHILE GETTING ORDER ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	switch body.Status {
	case string(model.Delivered):
		theOrder.Status = model.Delivered
	case string(model.Completed):
		theOrder.Status = model.Completed
	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	theOrder.UpdatedAt = &now

	err = o.Helper.Update(r.Context(), theOrder)
	if err != nil {
		fmt.Println("SOMETHING WENT WRONG WHILE UPDATE ORDER", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(theOrder); err != nil {
		fmt.Println("SOMETHING WENT WRONG WHILE PARSING ORDER", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (o *Order) RemoveOrder(w http.ResponseWriter, r *http.Request) {
	idParam := chi.URLParam(r, "id")

	orderId, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = o.Helper.DeleteByKey(r.Context(), orderId)
	if errors.Is(err, errors.New("ORDER DOES NOT EXIST")) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		fmt.Println("SOMETHING WENT WRONG WHILE GETTING ORDER ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
