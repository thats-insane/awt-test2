package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *appDependencies) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(a.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(a.notAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthCheckHandler)

	router.HandlerFunc(http.MethodPost, "/v1/product", a.createProductHandler)
	router.HandlerFunc(http.MethodGet, "/v1/products/:id", a.displayProductHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/products/:id", a.updateProductHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/products/:id", a.deleteProductHandler)
	router.HandlerFunc(http.MethodGet, "/v1/products", a.listProductsHandler)

	router.HandlerFunc(http.MethodPost, "/v1/review", a.createReviewHandler)
	router.HandlerFunc(http.MethodGet, "/v1/review/:id", a.displayReviewHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/review/:id", a.updateReviewHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/review/:id", a.deleteReviewHandler)
	router.HandlerFunc(http.MethodGet, "/v1/reviews", a.listReviewsHandler)

	return a.recoverPanic(router)
}
