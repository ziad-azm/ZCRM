// Package repositories holds data-access implementations.
//
// It owns the PostgreSQL connection pool (postgres.go) and the interfaces that
// services depend on. Concrete per-entity repositories are added by the feature
// stories that need them. Nothing in this package imports net/http.
package repositories
