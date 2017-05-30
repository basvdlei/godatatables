// Package mongo provides Datatables handlers for MongoDB.
package mongo

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/basvdlei/godatatables/types"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Query interface defines the *mgo.Query methods used.
type Query interface {
	All(result interface{}) error
	Count() (n int, err error)
	Limit(n int) Query
	Skip(n int) Query
	Sort(fields ...string) Query
}

// Collection interface contains the *mgo.Collection methods used.
type Collection interface {
	Count() (n int, err error)
	Find(query interface{}) Query
}

// queryWrapper wraps a *mgo.Query into Query interface to allow for mocked
// testing.
type queryWrapper struct {
	q *mgo.Query
}

// All wraps *mgo.Query.All().
func (w *queryWrapper) All(result interface{}) error {
	return w.q.All(result)
}

// Count wraps *mgo.Query.Count().
func (w *queryWrapper) Count() (n int, err error) {
	return w.q.Count()
}

// Limit wraps *mgo.Query.Limit().
func (w *queryWrapper) Limit(n int) Query {
	return &queryWrapper{
		q: w.q.Limit(n),
	}
}

// Skip wraps *mgo.Query.Skip().
func (w *queryWrapper) Skip(n int) Query {
	return &queryWrapper{
		q: w.q.Skip(n),
	}
}

// Sort wraps *mgo.Query.Sort().
func (w *queryWrapper) Sort(fields ...string) Query {
	return &queryWrapper{
		q: w.q.Sort(fields...),
	}
}

// collectionWrapper wraps a *mgo.Collection into Query interface to allow for mocked
// testing.
type collectionWrapper struct {
	c *mgo.Collection
}

// Count wraps *mgo.Collection.Count().
func (cw *collectionWrapper) Count() (n int, err error) {
	return cw.c.Count()
}

// Find wraps *mgo.Collection.Find().
func (cw *collectionWrapper) Find(query interface{}) Query {
	return &queryWrapper{
		q: cw.c.Find(query),
	}
}

// CollectionHandler provides a HTTP handler for a mgo collection.
type CollectionHandler struct {
	Collection Collection
}

// NewCollectionHandler returns a CollectionHandler for the given collection.
func NewCollectionHandler(c *mgo.Collection) *CollectionHandler {
	return &CollectionHandler{
		Collection: &collectionWrapper{c: c},
	}
}

// ServeHTTP implements the http.Handler interface
func (ch *CollectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	dtRequest, err := types.ParseURLValues(r.Form)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var dtResponse types.Response
	dtResponse.Draw = dtRequest.Draw
	f := CreateFilter(dtRequest)
	q := ch.Collection.Find(f)
	dtResponse.RecordsFiltered, err = q.Count()
	if err != nil {
		dtResponse.Error = err.Error()
	}
	dtResponse.RecordsTotal, err = ch.Collection.Count()
	if err != nil {
		dtResponse.Error = err.Error()
	}
	q = SortQuery(q, dtRequest)
	q = RangeQuery(q, dtRequest)
	dtResponse.Data, err = ResponseData(q)
	if err != nil {
		dtResponse.Error = err.Error()
	}
	e := json.NewEncoder(w)
	err = e.Encode(&dtResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ResponseData returns the data for a given query that can be used in a
// Datatables Response.
func ResponseData(q Query) (data []types.Row, err error) {
	var results []map[string]string
	if err = q.All(&results); err != nil {
		return nil, err
	}
	data = make([]types.Row, len(results))
	for i, r := range results {
		data[i].Data = r
	}
	return
}

// SortQuery sets the queries sort options based on the Request.
func SortQuery(in Query, r types.Request) (out Query) {
	sort := make([]string, len(r.Order))
	for i, o := range r.Order {
		prefix := ""
		if o.Dir == types.OrderDescending {
			prefix = "-"
		}
		sort[i] = prefix + r.Columns[o.Column].Data
	}
	out = in.Sort(sort...)
	return
}

// RangeQuery sets range of items to return based on the Datatables Request.
func RangeQuery(in Query, r types.Request) (out Query) {
	out = in.Skip(r.Start)
	out = out.Limit(r.Length)
	return
}

// CreateFilter creates a BSON query from a Datatables Request.
func CreateFilter(r types.Request) bson.M {
	global := make([]bson.M, len(r.Columns))
	column := make([]bson.M, 0, len(r.Columns))
	for i, c := range r.Columns {
		// Global search
		global[i] = make(bson.M, 1)
		if r.Search.Regex {
			global[i][c.Data] = bson.RegEx{
				Pattern: r.Search.Value,
				Options: "i",
			}
		} else {
			global[i][c.Data] = bson.RegEx{
				Pattern: regexp.QuoteMeta(r.Search.Value),
				Options: "i",
			}
		}
		// Column specific search
		if c.Search.Value != "" {
			m := make(bson.M, 1)
			if c.Search.Regex {
				m[c.Data] = bson.RegEx{
					Pattern: c.Search.Value,
					Options: "i",
				}
			} else {
				m[c.Data] = bson.RegEx{
					Pattern: regexp.QuoteMeta(c.Search.Value),
					Options: "i",
				}
			}
			column = append(column, m)
		}
	}
	q := bson.M{"$or": global}
	if len(column) > 0 {
		columnfind := bson.M{"$and": column}
		q = bson.M{"$and": []bson.M{q, columnfind}}
	}
	return q
}
