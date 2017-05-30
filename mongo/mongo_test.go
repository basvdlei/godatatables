package mongo

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/basvdlei/godatatables/types"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type RequestTestCase struct {
	Request      types.Request
	SortColumns  []string
	Result       []map[string]string
	ResponseData []types.Row
	Filter       bson.M
}

var RequestTests = []RequestTestCase{
	{
		Request: types.Request{
			Draw:   1,
			Start:  5,
			Length: 10,
			Order:  nil,
			Search: types.Search{
				Value: "test",
				Regex: false,
			},
			Columns: []types.Column{
				{
					Data:       "foo",
					Name:       "",
					Orderable:  true,
					Searchable: true,
					Search: types.Search{
						Value: "",
						Regex: false,
					},
				},
				{
					Data:       "bar",
					Name:       "",
					Orderable:  false,
					Searchable: false,
					Search:     types.Search{},
				},
			},
		},
		SortColumns: []string{},
		Result: []map[string]string{
			{
				"foo": "1",
				"bar": "2",
			},
			{
				"foo": "3",
				"bar": "4",
			},
		},
		ResponseData: []types.Row{
			{
				Data: map[string]string{
					"foo": "1",
					"bar": "2",
				},
			},
			{
				Data: map[string]string{
					"foo": "3",
					"bar": "4",
				},
			},
		},
		Filter: bson.M{
			"$or": []bson.M{
				{
					"foo": bson.RegEx{
						Pattern: "test",
						Options: "i",
					},
				},
				{
					"bar": bson.RegEx{
						Pattern: "test",
						Options: "i",
					},
				},
			},
		},
	},
	{
		Request: types.Request{
			Draw:   10,
			Start:  25,
			Length: 100,
			Order: []types.Order{
				{
					Column: 1,
					Dir:    types.OrderDescending,
				},
			},
			Search: types.Search{
				Value: "^test$",
				Regex: true,
			},
			Columns: []types.Column{
				{
					Data:       "foo",
					Name:       "",
					Orderable:  false,
					Searchable: true,
					Search: types.Search{
						Value: "test",
						Regex: false,
					},
				},
				{
					Data:       "bar",
					Name:       "",
					Orderable:  true,
					Searchable: true,
					Search: types.Search{
						Value: "^test$",
						Regex: true,
					},
				},
			},
		},
		SortColumns: []string{"-bar"},
		Result: []map[string]string{
			{
				"foo": "1",
				"bar": "2",
			},
			{
				"foo": "3",
				"bar": "4",
			},
		},
		ResponseData: []types.Row{
			{
				Data: map[string]string{
					"foo": "1",
					"bar": "2",
				},
			},
			{
				Data: map[string]string{
					"foo": "3",
					"bar": "4",
				},
			},
		},
		Filter: bson.M{
			"$and": []bson.M{
				{
					"$or": []bson.M{
						{
							"foo": bson.RegEx{
								Pattern: "^test$",
								Options: "i",
							},
						},
						{
							"bar": bson.RegEx{
								Pattern: "^test$",
								Options: "i",
							},
						},
					},
				},
				{
					"$and": []bson.M{
						{
							"foo": bson.RegEx{
								Pattern: "test",
								Options: "i",
							},
						},
						{
							"bar": bson.RegEx{
								Pattern: "^test$",
								Options: "i",
							},
						},
					},
				},
			},
		},
	},
}

type QueryMock struct {
	Result      []map[string]string
	CountCalled bool
	LimitValue  int
	SkipValue   int
	SortValue   []string
}

func (q *QueryMock) All(result interface{}) error {
	if v, ok := result.(*[]map[string]string); ok {
		*v = append(*v, q.Result...)
		return nil
	}
	return errors.New("unknown type")
}
func (q *QueryMock) Count() (n int, err error) {
	q.CountCalled = true
	return
}
func (q *QueryMock) Limit(n int) Query {
	q.LimitValue = n
	return q
}
func (q *QueryMock) Skip(n int) Query {
	q.SkipValue = n
	return q
}
func (q *QueryMock) Sort(fields ...string) Query {
	q.SortValue = fields
	return q
}

type CollectionMock struct {
	count int
	err   error
	query *QueryMock
}

func (c *CollectionMock) Count() (n int, err error) {
	return c.count, c.err
}
func (c *CollectionMock) Find(query interface{}) Query {
	return c.query
}

func TestCollectionHandlerServeHTTP(t *testing.T) {
	for i, c := range RequestTests {
		var totalRecords = 100
		ch := &CollectionHandler{
			Collection: &CollectionMock{
				count: totalRecords,
				err:   nil,
				query: &QueryMock{
					Result: c.Result,
				},
			},
		}
		req := &http.Request{
			Method: "GET",
			URL:    &url.URL{Path: "/"},
			Form: url.Values{
				"draw": []string{strconv.Itoa(c.Request.Draw)},
			},
		}
		w := httptest.NewRecorder()
		ch.ServeHTTP(w, req)
		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Errorf("case %d: unexpected statuscode, want %d, got %d",
				i, http.StatusOK, resp.StatusCode)
		}
		dec := json.NewDecoder(resp.Body)
		var dtResponse types.Response
		err := dec.Decode(&dtResponse)
		if err != nil {
			t.Errorf("case %d: could not marshal response: %v", i, err)
		}
		if dtResponse.Error != "" {
			t.Errorf("case %d: unexpected error returned. want %v, got %v",
				i, "", dtResponse.Error)
		}
		if dtResponse.Draw != c.Request.Draw {
			t.Errorf("case %d: draw value does not match. want %d, got %d",
				i, c.Request.Draw, dtResponse.Draw)
		}
		if dtResponse.RecordsTotal != totalRecords {
			t.Errorf("case %d: totalRecords does not match. want %d, got %d",
				i, totalRecords, dtResponse.RecordsTotal)
		}
		if !reflect.DeepEqual(dtResponse.Data, c.ResponseData) {
			t.Errorf("case %d: data does not match. want %v, got %v",
				i, c.ResponseData, dtResponse.Data)
		}
	}
}

func TestResponseData(t *testing.T) {
	for i, c := range RequestTests {
		q := &QueryMock{
			Result: c.Result,
		}
		data, err := ResponseData(q)
		if err != nil {
			t.Errorf("case %d: error %v", i, err)
		}
		if !reflect.DeepEqual(data, c.ResponseData) {
			t.Errorf("case %d: data does not match, want %+v, got %+v",
				i, c.ResponseData, data)
		}
	}
}

func TestSortQuery(t *testing.T) {
	for i, c := range RequestTests {
		q := SortQuery(&QueryMock{}, c.Request)
		if v, ok := q.(*QueryMock); ok {
			if len(v.SortValue) != len(c.SortColumns) {
				t.Errorf("case %d: sort columns count does not match, want %d, got %d",
					i, len(c.SortColumns), len(v.SortValue))
			}
			for i, s := range c.SortColumns {
				if v.SortValue[i] != s {
					t.Errorf("case %d: sortcolumn does not match, want %s, got %s",
						i, v.SortValue[i], s)
				}
			}
		} else {
			t.Errorf("bad query type")
		}
	}
}

func TestRangeQuery(t *testing.T) {
	for i, c := range RequestTests {
		q := RangeQuery(&QueryMock{}, c.Request)
		if v, ok := q.(*QueryMock); ok {
			if v.LimitValue != c.Request.Length {
				t.Errorf("case %d: limit does not match, want %d, got %d",
					i, c.Request.Length, v.LimitValue)
			}
			if v.SkipValue != c.Request.Start {
				t.Errorf("case %d: skip does not match, want %d, got %d",
					i, c.Request.Start, v.SkipValue)
			}
		} else {
			t.Errorf("bad query type")
		}
	}
}

func TestCreateFilter(t *testing.T) {
	for i, c := range RequestTests {
		f := CreateFilter(c.Request)
		if !reflect.DeepEqual(f, c.Filter) {
			t.Errorf("case %d: filter not match, want %+v, got %+v",
				i, c.Filter, f)
		}
	}
}

func ExampleCollectionHandler() {
	session, _ := mgo.Dial("mymongohost")
	c := session.DB("mydb").C("mycollection")
	http.Handle("/mycollection", NewCollectionHandler(c))
	http.ListenAndServe(":8080", nil)
}
