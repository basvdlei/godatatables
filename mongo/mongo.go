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

// CollectionHandler provides a HTTP handler for a mgo collection.
type CollectionHandler struct {
	Collection *mgo.Collection
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
	dtResponse.Data = ResponseData(q)
	e := json.NewEncoder(w)
	err = e.Encode(&dtResponse)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// ResponseData returns the data for a given query that can be used in a
// Datatables Response.
func ResponseData(q *mgo.Query) (data []types.Row) {
	results := make([]map[string]string, 0)
	q.All(&results)
	data = make([]types.Row, len(results))
	for i, r := range results {
		data[i].Data = r
	}
	return
}

// SortQuery sets the queries sort options based on the Request.
func SortQuery(in *mgo.Query, r types.Request) (out *mgo.Query) {
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
func RangeQuery(in *mgo.Query, r types.Request) (out *mgo.Query) {
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
			}
		} else {
			global[i][c.Data] = bson.RegEx{
				Pattern: regexp.QuoteMeta(r.Search.Value),
			}
		}
		// Column specific search
		if c.Search.Value != "" {
			m := make(bson.M, 1)
			if c.Search.Regex {
				m[c.Data] = bson.RegEx{
					Pattern: c.Search.Value,
				}
			} else {
				m[c.Data] = bson.RegEx{
					Pattern: regexp.QuoteMeta(c.Search.Value),
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
