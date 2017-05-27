package types

import (
	"encoding/json"
	"net/url"
	"reflect"
	"testing"
)

type unmarshalRespTestCase struct {
	Name   string
	Input  string
	Output Response
}

var unmarshalRespTests = []unmarshalRespTestCase{
	{
		Name: "array-rowdata",
		Input: `{
  "draw": 9,
  "recordsTotal": 57,
  "recordsFiltered": 2,
  "data": [
    [
      "Airi",
      "Satou"
    ],
    [
      "Dai",
      "Rios"
    ]
  ]
}`,
		Output: Response{
			Draw:            9,
			RecordsTotal:    57,
			RecordsFiltered: 2,
			Data: []Row{
				{
					Data: map[string]string{
						"0": "Airi",
						"1": "Satou",
					},
				},
				{
					Data: map[string]string{
						"0": "Dai",
						"1": "Rios",
					},
				},
			},
		},
	},
	{
		Name: "object-rowdata",
		Input: `{
  "draw": 1,
  "recordsTotal": 57,
  "recordsFiltered": 2,
  "data": [
    {
      "first_name": "Airi",
      "last_name": "Satou"
    },
    {
      "first_name": "Angelica",
      "last_name": "Ramos"
    }
  ]
}`,
		Output: Response{
			Draw:            1,
			RecordsTotal:    57,
			RecordsFiltered: 2,
			Data: []Row{
				{
					Data: map[string]string{
						"first_name": "Airi",
						"last_name":  "Satou",
					},
				},
				{
					Data: map[string]string{
						"first_name": "Angelica",
						"last_name":  "Ramos",
					},
				},
			},
		},
	},
	{
		Name: "custom-id-class",
		Input: `{
  "draw": 1,
  "recordsTotal": 57,
  "recordsFiltered": 2,
  "data": [
    {
      "DT_RowId": "row_1",
      "DT_RowClass": "rowclass",
      "first_name": "Airi",
      "last_name": "Satou"
    },
    {
      "DT_RowId": "row_2",
      "DT_RowClass": "rowclassspecial",
      "first_name": "Angelica",
      "last_name": "Ramos"
    }
  ]
}`,
		Output: Response{
			Draw:            1,
			RecordsTotal:    57,
			RecordsFiltered: 2,
			Data: []Row{
				{
					RowID:    "row_1",
					RowClass: "rowclass",
					Data: map[string]string{
						"first_name": "Airi",
						"last_name":  "Satou",
					},
				},
				{
					RowID:    "row_2",
					RowClass: "rowclassspecial",
					Data: map[string]string{
						"first_name": "Angelica",
						"last_name":  "Ramos",
					},
				},
			},
		},
	},
}

func TestUnmarshalResponse(t *testing.T) {
	for _, v := range unmarshalRespTests {
		var r Response
		err := json.Unmarshal([]byte(v.Input), &r)
		if err != nil {
			t.Errorf("case %s: error %v", v.Name, err)
		}
		if !reflect.DeepEqual(r, v.Output) {
			t.Errorf("case %s: want: %+v, got %+v",
				v.Name, v.Output, r)
		}
	}
}

type marshalRespTestCase struct {
	Name   string
	Input  Response
	Output string
}

var marshalRespTests = []marshalRespTestCase{
	{
		Name: "object-data",
		Input: Response{
			Draw:            5,
			RecordsTotal:    2,
			RecordsFiltered: 2,
			Data: []Row{
				{
					Data: map[string]string{
						"name": "Foo",
						"age":  "16",
					},
				},
				{
					Data: map[string]string{
						"name": "Bar",
						"age":  "32",
					},
				},
			},
		},
		Output: `{"draw":5,"recordsTotal":2,"recordsFiltered":2,"data":[{"name":"Foo","age":"16"},{"name":"Bar","age":"32"}]}`,
	},
	{
		Name: "custom-id-class",
		Input: Response{
			Draw:            5,
			RecordsTotal:    2,
			RecordsFiltered: 2,
			Data: []Row{
				{
					RowID:    "row_1",
					RowClass: "odd",
					Data: map[string]string{
						"name": "Foo",
						"age":  "16",
					},
				},
				{
					RowID:    "row_2",
					RowClass: "even",
					Data: map[string]string{
						"name": "Bar",
						"age":  "32",
					},
				},
			},
		},
		Output: `{"draw":5,"recordsTotal":2,"recordsFiltered":2,"data":[{"DT_RowId":"row_1","DT_RowClass":"odd","name":"Foo","age":"16"},{"DT_RowId":"row_2","DT_RowClass":"even","name":"Bar","age":"32"}]}`,
	},
}

func TestMarshalResponse(t *testing.T) {
	for _, v := range marshalRespTests {
		out, err := json.Marshal(v.Input)
		if err != nil {
			t.Errorf("case %s: error %v", v.Name, err)
		}
		var r Response
		err = json.Unmarshal(out, &r)
		if err != nil {
			t.Errorf("case %s: could not unmarshal, marshaled response: %v",
				v.Name, err)
		}
		if !reflect.DeepEqual(r, v.Input) {
			t.Errorf("case %s: want: %+v, got %+v",
				v.Name, v.Input, r)
		}
	}
}

type unmarshalReqTestCase struct {
	Name   string
	Input  string
	Output Request
}

var unmarshalReqTests = []unmarshalReqTestCase{
	{
		Name: "simple-request",
		Input: `{
  "draw": 1,
  "columns": [
    {
      "data": "1",
      "name": "",
      "searchable": true,
      "orderable": true,
      "search": {
        "value": "",
        "regex": false
      }
    },
    {
      "data": "2",
      "name": "",
      "searchable": true,
      "orderable": true,
      "search": {
        "value": "",
        "regex": false
      }
    }
  ],
  "order": [
    {
      "column": 0,
      "dir": "asc"
    }
  ],
  "start": 0,
  "length": 10,
  "search": {
    "value": "test"
  }
}`,
		Output: Request{
			Draw: 1,
			Columns: []Column{
				{
					Data:       "1",
					Name:       "",
					Searchable: true,
					Orderable:  true,
					Search: Search{
						Value: "",
						Regex: false,
					},
				},
				{
					Data:       "2",
					Name:       "",
					Searchable: true,
					Orderable:  true,
					Search: Search{
						Value: "",
						Regex: false,
					},
				},
			},
			Order: []Order{
				{
					Column: 0,
					Dir:    OrderAscending,
				},
			},
			Start:  0,
			Length: 10,
			Search: Search{
				Value: "test",
				Regex: false,
			},
		},
	},
}

func TestUnmarshalRequest(t *testing.T) {
	for _, v := range unmarshalReqTests {
		var r Request
		err := json.Unmarshal([]byte(v.Input), &r)
		if err != nil {
			t.Errorf("case %s: error %v", v.Name, err)
		}
		if !reflect.DeepEqual(r, v.Output) {
			t.Errorf("case %s: want: %+v, got %+v",
				v.Name, v.Output, r)
		}
	}
}

type marshalReqTestCase struct {
	Name   string
	Input  Request
	Output string
}

var marshalReqTests = []marshalReqTestCase{
	{
		Name: "simple-request",
		Input: Request{
			Draw: 10,
			Columns: []Column{
				{
					Data:       "1",
					Name:       "",
					Searchable: true,
					Orderable:  true,
					Search: Search{
						Value: "^bla$",
						Regex: true,
					},
				},
				{
					Data:       "2",
					Name:       "",
					Searchable: true,
					Orderable:  true,
					Search: Search{
						Value: "",
						Regex: false,
					},
				},
			},
			Order: []Order{
				{
					Column: 0,
					Dir:    OrderAscending,
				},
			},
			Start:  0,
			Length: 10,
			Search: Search{
				Value: "test",
				Regex: false,
			},
		},
		Output: `{"draw":10,"columns":[{"data":"1","name":"","searchable":true,"orderable":true,"search":{"value":"^bla$","regex":true}},{"data":"2","name":"","searchable":true,"orderable":true,"search":{"value":"","regex":false}}],"order":[{"column":0,"dir":"asc"}],"start":0,"length":10,"search":{"value":"test","regex":false}
}`,
	},
}

func TestMarshalRequest(t *testing.T) {
	for _, v := range marshalReqTests {
		out, err := json.Marshal(v.Input)
		if err != nil {
			t.Errorf("case %s: error %v", v.Name, err)
		}
		var r Request
		err = json.Unmarshal(out, &r)
		if err != nil {
			t.Errorf("case %s: could not unmarshal, marshaled response: %v",
				v.Name, err)
		}
		if !reflect.DeepEqual(r, v.Input) {
			t.Errorf("case %s: want: %+v, got %+v",
				v.Name, v.Input, r)
		}
	}
}

type decTestCase struct {
	Name   string
	Input  url.Values
	Output Request
}

var decTests = []decTestCase{
	{
		Name: "1",
		Input: url.Values{
			"draw":                      []string{"2"},
			"columns[0][data]":          []string{"0"},
			"columns[0][name]":          []string{""},
			"columns[0][searchable]":    []string{"true"},
			"columns[0][orderable]":     []string{"true"},
			"columns[0][search][value]": []string{"^bla$"},
			"columns[0][search][regex]": []string{"true"},
			"columns[1][data]":          []string{"1"},
			"columns[1][name]":          []string{""},
			"columns[1][searchable]":    []string{"false"},
			"columns[1][orderable]":     []string{"false"},
			"columns[1][search][value]": []string{""},
			"columns[1][search][regex]": []string{"false"},
			"order[0][column]":          []string{"0"},
			"order[0][dir]":             []string{"asc"},
			"start":                     []string{"0"},
			"length":                    []string{"10"},
			"search[value]":             []string{"t"},
			"search[regex]":             []string{"false"},
			"_":                         []string{"1495876460828"},
		},
		Output: Request{
			Draw: 2,
			Columns: []Column{
				{
					Data:       "0",
					Name:       "",
					Searchable: true,
					Orderable:  true,
					Search: Search{
						Value: "^bla$",
						Regex: true,
					},
				},
				{
					Data:       "1",
					Name:       "",
					Searchable: false,
					Orderable:  false,
					Search: Search{
						Value: "",
						Regex: false,
					},
				},
			},
			Order: []Order{
				{
					Column: 0,
					Dir:    OrderAscending,
				},
			},
			Start:  0,
			Length: 10,
			Search: Search{
				Value: "t",
				Regex: false,
			},
		},
	},
}

func TestParseURLValues(t *testing.T) {
	r, err := ParseURLValues(decTests[0].Input)
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(r, decTests[0].Output) {
		t.Errorf("case %s: want %+v, got %+v\n",
			decTests[0].Name, decTests[0].Output, r)
	}

}
