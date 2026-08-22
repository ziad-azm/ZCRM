// Package repositories holds data-access implementations.
//
// It is intentionally empty until ZCRM-5 (Database & Migrations), which adds
// the pgxpool connection and the first concrete repository. Services depend on
// interfaces declared here; nothing in this package imports net/http.
package repositories
