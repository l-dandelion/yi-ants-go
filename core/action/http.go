package action

import "net/http"

type HttpServer interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Welcome(w http.ResponseWriter, r *http.Request)
}

type HttpServerCluster interface {
	Cluster(w http.ResponseWriter, r *http.Request)
}

type HttpServerCrawler interface {
	Spiders(w http.ResponseWriter, r *http.Request)
	Crawl(w http.ResponseWriter, r *http.Request)
}

type HttpServerAnts interface {
	HttpServer
	HttpServerCluster
	HttpServerCrawler
}
