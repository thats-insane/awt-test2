package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/thats-insane/awt-test1/internal/data"
	"github.com/thats-insane/awt-test1/internal/validator"
)

func (a *appDependencies) createReviewHandler(w http.ResponseWriter, r *http.Request) {
	var incomingData struct {
		ProductID    *int64  `json:"product_id"`
		Author       *string `json:"author"`
		Rating       *int64  `json:"rating"`
		HelpfulCount *int32  `json:"helpful_count"`
	}

	err := a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if incomingData.ProductID == nil {
		a.badRequestResponse(w, r, errors.New("product id is required"))
		return
	}

	exists, err := a.productModel.Exists(*incomingData.ProductID)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}
	if !exists {
		a.notFoundResponse(w, r)
		return
	}

	if incomingData.HelpfulCount == nil {
		incomingData.HelpfulCount = new(int32)
	}

	review := &data.Review{
		ProductID:    *incomingData.ProductID,
		Author:       *incomingData.Author,
		Rating:       *incomingData.Rating,
		HelpfulCount: *incomingData.HelpfulCount,
	}

	v := validator.New()
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.reviewModel.Insert(review)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/reviews/%d", review.ID))

	data := envelope{
		"Review": review,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}

	fmt.Fprintf(w, "%+v\n", incomingData)
}

func (a *appDependencies) displayReviewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	review, err := a.reviewModel.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"Review": review,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}

}

func (a *appDependencies) updateReviewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	review, err := a.reviewModel.Get(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrResponse(w, r, err)
		}
		return
	}

	var incomingData struct {
		Author       *string `json:"author"`
		Rating       *int64  `json:"rating"`
		HelpfulCount *int32  `json:"helpful_count"`
	}

	err = a.readJSON(w, r, &incomingData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	if incomingData.Author != nil {
		review.Author = *incomingData.Author
	}
	if incomingData.Rating != nil {
		review.Rating = *incomingData.Rating
	}
	if incomingData.HelpfulCount != nil {
		review.HelpfulCount = *incomingData.HelpfulCount
	}

	v := validator.New()
	data.ValidateReview(v, review)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = a.reviewModel.Update(review)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}

	data := envelope{
		"review": review,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrResponse(w, r, err)
	}
}

func (a *appDependencies) deleteReviewHandler(w http.ResponseWriter, r *http.Request) {
	id, err := a.readIDParam(r)
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	err = a.reviewModel.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrResponse(w, r, err)
		}
		return
	}

	data := envelope{
		"message": "review successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrResponse(w, r, err)
	}
}

func (a *appDependencies) listReviewsHandler(w http.ResponseWriter, r *http.Request) {
	var queryParamsData struct {
		Author       string
		Rating       string
		HelpfulCount string
		data.Filters
	}

	queryParams := r.URL.Query()

	queryParamsData.Author = a.getSingleQueryParam(queryParams, "author", "")
	queryParamsData.Rating = a.getSingleQueryParam(queryParams, "rating", "")
	queryParamsData.HelpfulCount = a.getSingleQueryParam(queryParams, "helpful_count", "")
	v := validator.New()
	queryParamsData.Filters.Page = a.getSingleIntParam(queryParams, "page", 1, v)
	queryParamsData.Filters.PageSize = a.getSingleIntParam(queryParams, "page_size", 10, v)
	queryParamsData.Filters.Sort = a.getSingleQueryParam(queryParams, "sort", "id")
	queryParamsData.Filters.SortSafeList = []string{"id", "author", "-id", "-author"}

	data.ValidateFilters(v, queryParamsData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	reviews, metadata, err := a.reviewModel.GetAll(queryParamsData.Author, queryParamsData.Rating, queryParamsData.HelpfulCount, queryParamsData.Filters)
	if err != nil {
		a.serverErrResponse(w, r, err)
		return
	}

	data := envelope{
		"reviews":   reviews,
		"@metadata": metadata,
	}

	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrResponse(w, r, err)
	}
}
