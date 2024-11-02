package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/thats-insane/awt-test1/internal/data"
	"github.com/thats-insane/awt-test1/internal/validator"
)

var ErrRecordNotFound = errors.New("record not found")

func (a *appDependencies) createProductHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		Name          string  `json:"name"`
		Description   string  `json:"description"`
		Category      string  `json:"category"`
		Price         float64 `json:"price"`
		AverageRating float64 `json:"average_rating"`
		ImageURL      string  `json:"image_url"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	product := &data.Product{
		Name:          incomingData.Name,
		Description:   incomingData.Description,
		Category:      incomingData.Category,
		Price:         incomingData.Price,
		AverageRating: incomingData.AverageRating,
		ImageURL:      incomingData.ImageURL,
	}

	v := validator.New()
	data.ValidateProduct(v, product, 1)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.productModel.Insert(product)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/product/%d", product.ID))

	data := envelope{
		"product": product,
	}

	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}

	fmt.Fprintf(w, "%+v\n", incomingData)
}

func (a *appDependencies) displayProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	product, err := a.productModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"product": product,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}
}

func (a *appDependencies) updateProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	product, err := a.productModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrResponse(w, r, err)
		}
		return
	}

	var incomingData struct {
		Name          *string  `json:"name"`
		Description   *string  `json:"description"`
		Category      *string  `json:"category"`
		Price         *float64 `json:"price"`
		AverageRating *float64 `json:"average_rating"`
		ImageURL      *string  `json:"image_url"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if incomingData.Name != nil {
		product.Name = *incomingData.Name
	}
	if incomingData.Description != nil {
		product.Description = *incomingData.Description
	}
	if incomingData.Category != nil {
		product.Category = *incomingData.Category
	}
	if incomingData.Price != nil {
		product.Price = *incomingData.Price
	}
	if incomingData.AverageRating != nil {
		product.AverageRating = *incomingData.AverageRating
	}
	if incomingData.ImageURL != nil {
		product.ImageURL = *incomingData.ImageURL
	}

	v := validator.New()

	data.ValidateProduct(v, product, 1)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.productModel.Update(product)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}

	data := envelope{
		"product": product,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}
}

func (a *appDependencies) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.productModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"message": "product successfully deleted",
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrResponse(w, r, err)
	}
}

func (a *appDependencies) listProductsHandler(w http.ResponseWriter, r *http.Request) {
	var queryParamsData struct {
		Name          string
		Description   string
		Category      string
		Price         string
		AverageRating string
		ImageURL      string
		data.Filters
	}

	queryParams := r.URL.Query()
	queryParamsData.Name = a.getSingleQueryParam(queryParams, "name", "")
	queryParamsData.Description = a.getSingleQueryParam(queryParams, "description", "")
	queryParamsData.Category = a.getSingleQueryParam(queryParams, "category", "")
	queryParamsData.Price = a.getSingleQueryParam(queryParams, "price", "")
	queryParamsData.AverageRating = a.getSingleQueryParam(queryParams, "average_rating", "")
	queryParamsData.ImageURL = a.getSingleQueryParam(queryParams, "image_url", "")
	v := validator.New()
	queryParamsData.Filters.Page = a.getSingleIntParam(queryParams, "page", 1, v)
	queryParamsData.Filters.PageSize = a.getSingleIntParam(queryParams, "page_size", 10, v)
	queryParamsData.Filters.Sort = a.getSingleQueryParam(queryParams, "sort", "id")
	queryParamsData.Filters.SortSafeList = []string{"id", "name", "-id", "-name"}

	data.ValidateFilters(v, queryParamsData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	product, metadata, err := a.productModel.GetAll(queryParamsData.Name, queryParamsData.Description, queryParamsData.Category, queryParamsData.Price, queryParamsData.AverageRating, queryParamsData.ImageURL, queryParamsData.Filters)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}

	data := envelope{
		"product":   product,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrResponse(w, r, err)
	}
}
