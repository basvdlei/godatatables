package types

import (
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var (
	// columnRegexp is the column urlvalue regexp (1=id 2=field 3=subfields)
	columnRegexp = regexp.MustCompile(`(?U)^columns\[([0-9]+)\]\[(.+)\](.*)$`)
	// searchRegexp is the search urlvalue regexp (1=field)
	searchRegexp = regexp.MustCompile(`(?U)^search\[(.+)\]$`)
	// orderRegexp is the order urlvalue regexp (1=id 2=field)
	orderRegexp = regexp.MustCompile(`(?U)^order\[([0-9]+)\]\[(.+)\]$`)

	// ErrNotEnoughFields is returned when the urlvalues does not contain
	// enough fields to parse.
	ErrNotEnoughFields = errors.New("not enough fields")
)

// UnmarshalJSON implements the json.Unmarshaler interface.
func (r *Row) UnmarshalJSON(in []byte) error {
	// Try to parse rowdata as an array first
	var rowData []string
	err := json.Unmarshal(in, &rowData)
	if err == nil {
		r.Data = make(map[string]string, len(rowData))
		for i, v := range rowData {
			r.Data[strconv.Itoa(i)] = v
		}
		return nil
	}
	// Otherwise assume it's an object
	type rowCopy struct {
		RowID    string            `json:"DT_RowId,omitempty"`
		RowClass string            `json:"DT_RowClass,omitempty"`
		RowData  map[string]string `json:"DT_RowData,omitempty"`
		RowAttr  map[string]string `json:"DT_RowAttr,omitempty"`
	}
	var c rowCopy
	err = json.Unmarshal(in, &c)
	if err != nil {
		return err
	}
	r.RowID = c.RowID
	r.RowClass = c.RowClass
	r.RowData = c.RowData
	r.RowAttr = c.RowAttr

	var data = make(map[string]string)
	err = json.Unmarshal(in, &data)
	if err != nil {
		return err
	}
	for _, v := range []string{"DT_RowId", "DT_RowClass", "DT_RowData", "DT_RowAttr"} {
		delete(data, v)
	}
	r.Data = data
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (r Row) MarshalJSON() ([]byte, error) {
	c := make(map[string]interface{})
	for k, v := range r.Data {
		c[k] = v
	}
	if r.RowID != "" {
		c["DT_RowId"] = r.RowID
	}
	if r.RowClass != "" {
		c["DT_RowClass"] = r.RowClass
	}
	if r.RowData != nil && len(r.RowData) > 0 {
		c["DT_RowData"] = r.RowData
	}
	if r.RowAttr != nil && len(r.RowAttr) > 0 {
		c["DT_RowAttr"] = r.RowAttr
	}
	return json.Marshal(&c)
}

// ParseURLValues parses http request url.Values into a Request.
func ParseURLValues(u url.Values) (r Request, err error) {
	for k, v := range u {
		if len(v) < 1 {
			continue
		}
		switch true {
		case k == "draw":
			r.Draw, err = strconv.Atoi(v[0])
		case k == "start":
			r.Start, err = strconv.Atoi(v[0])
		case k == "length":
			r.Length, err = strconv.Atoi(v[0])
		case strings.HasPrefix(k, "search"):
			r.Search, err = parseSearch(r.Search, k, v[0])
		case strings.HasPrefix(k, "order"):
			r.Order, err = parseOrder(r.Order, k, v[0])
		case strings.HasPrefix(k, "column"):
			r.Columns, err = parseColumn(r.Columns, k, v[0])
		}
		if err != nil {
			return
		}
	}
	return
}

// parseOrder parses the order urlvalue fields.
// eg `order[0][...]`
func parseOrder(o []Order, k, v string) (out []Order, err error) {
	m := orderRegexp.FindStringSubmatch(k)
	if len(m) < 3 {
		return o, ErrNotEnoughFields
	}
	id, err := strconv.Atoi(m[1])
	if err != nil {
		return nil, err
	}
	if id+1 > len(o) {
		out = make([]Order, id+1)
		copy(out, o)
	} else {
		out = o
	}
	switch m[2] {
	case "column":
		out[id].Column, err = strconv.Atoi(v)
	case "dir":
		if v == "asc" {
			out[id].Dir = OrderAscending
		} else if v == "desc" {
			out[id].Dir = OrderDescending
		}
	}
	return
}

// parseSearch parses the search urlvalue fields.
// eg `search[i][...]`
func parseSearch(s Search, k, v string) (out Search, err error) {
	m := searchRegexp.FindStringSubmatch(k)
	if len(m) < 2 {
		return s, ErrNotEnoughFields
	}
	out = s
	switch m[1] {
	case "value":
		out.Value = v
	case "regex":
		if v == "true" {
			out.Regex = true
		} else {
			out.Regex = false
		}

	}
	return
}

// parseColumn parses the column urlvalue fields.
// eg `cloumns[i][...]
func parseColumn(in []Column, k, v string) (out []Column, err error) {
	m := columnRegexp.FindStringSubmatch(k)
	if len(m) < 2 {
		return in, ErrNotEnoughFields
	}
	id, err := strconv.Atoi(m[1])
	if err != nil {
		return in, err
	}
	if id+1 > len(in) {
		out = make([]Column, id+1)
		copy(out, in)
	} else {
		out = in
	}

	switch m[2] {
	case "data":
		out[id].Data = v
	case "name":
		out[id].Name = v
	case "searchable":
		if v == "true" {
			out[id].Searchable = true
		} else {
			out[id].Searchable = false
		}
	case "orderable":
		if v == "true" {
			out[id].Orderable = true
		} else {
			out[id].Orderable = false
		}
	case "search":
		if len(m) > 3 {
			out[id].Search, err = parseSearch(out[id].Search, "search"+m[3], v)
		}
	}
	return
}
