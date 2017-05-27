// Package types provides the request and response types for Datatable calls.
package types

// OrderDirection specifies column ordering direction.
type OrderDirection string

const (
	// OrderAscending indicates ascending ordering.
	OrderAscending OrderDirection = "asc"
	// OrderDescending indicates descending ordering.
	OrderDescending OrderDirection = "desc"
)

// Response contains the outgoing table data.
type Response struct {
	// The draw counter that this object is a response to - from the draw
	// parameter sent as part of the data request. Note that it is strongly
	// recommended for security reasons that you cast this parameter to an
	// integer, rather than simply echoing back to the client what it sent
	// in the draw parameter, in order to prevent Cross Site Scripting
	// (XSS) attacks.
	Draw int `json:"draw"`
	// Total records, before filtering (i.e. the total number of records in
	// the database)
	RecordsTotal int `json:"recordsTotal"`
	// Total records, after filtering (i.e. the total number of records
	// after filtering has been applied - not just the number of records
	// being returned for this page of data).
	RecordsFiltered int `json:"recordsFiltered"`
	// The data to be displayed in the table. This is an array of data
	// source objects, one for each row, which will be used by DataTables.
	// Note that this parameter's name can be changed using the ajax
	// option's dataSrc property.
	Data []Row `json:"data"`
	// Optional: If an error occurs during the running of the server-side
	// processing script, you can inform the user of this error by passing
	// back the error message to be displayed using this parameter. Do not
	// include if there is no error.
	Error string `json:"error,omitempty"`
}

// Row contains the data columns.
type Row struct {
	// Column data.
	Data map[string]string `json:"-"`

	// Optional: Set the ID property of the tr node to this value
	RowID string `json:"DT_RowId,omitempty"`
	// Optional: Add this class to the tr node
	RowClass string `json:"DT_RowClass,omitempty"`
	// Optional: Add the data contained in the object to the row using the
	// jQuery data() method to set the data, which can also then be used
	// for later retrieval (for example on a click event).
	RowData map[string]string `json:"DT_RowData,omitempty"`
	// Optional: Add the data contained in the object to the row tr node as
	// attributes. The object keys are used as the attribute keys and the
	// values as the corresponding attribute values. This is performed
	// using using the jQuery param() method. Please note that this option
	// requires DataTables 1.10.5 or newer.
	RowAttr map[string]string `json:"DT_RowAttr,omitempty"`
}

// Request is the incoming Datatables request.
type Request struct {
	// Draw counter. This is used by DataTables to ensure that the Ajax
	// returns from server-side processing requests are drawn in sequence
	// by DataTables (Ajax requests are asynchronous and thus can return
	// out of sequence). This is used as part of the draw return parameter
	// (see below).
	Draw int `json:"draw"`
	// Paging first record indicator. This is the start point in the
	// current data set (0 index based - i.e. 0 is the first record).
	Start int `json:"start"`
	// Number of records that the table can display in the current draw. It
	// is expected that the number of records returned will be equal to
	// this number, unless the server has fewer records to return. Note
	// that this can be -1 to indicate that all records should be returned
	// (although that negates any benefits of server-side processing!)
	Length int `json:"length"`
	// Global search value. To be applied to all columns which have
	// searchable as true.
	Search Search `json:"search"`
	// Ordering direction for the columns.
	Order []Order `json:"order"`
	// Columns requests as specified in the column.data source options.
	Columns []Column `json:"columns"`
}

// Search contains the (regex) value to search for in a specific column.
type Search struct {
	// Search value to apply to this specific column.
	Value string `json:"value"`
	// Flag to indicate if the search term for this column should be
	// treated as regular expression (true) or not (false). As with global
	// search, normally server-side processing scripts will not perform
	// regular expression searching for performance reasons on large data
	// sets, but it is technically possible and at the discretion of your
	// script.
	Regex bool `json:"regex"`
}

// Order contains ordering information.
type Order struct {
	// Column to which ordering should be applied. This is an index
	// reference to the columns array of information that is also
	// submitted to the server.
	Column int `json:"column"`
	// Ordering direction for this column. It will be asc or desc to
	// indicate ascending ordering or descending ordering, respectively.
	Dir OrderDirection `json:"dir"`
}

// Column contains the requested column data.
type Column struct {
	// Column's data source, as defined by columns.data.
	Data string `json:"data"`
	// Column's name, as defined by columns.name.
	Name string `json:"name"`
	// Flag to indicate if this column is searchable (true) or not (false).
	// This is controlled by columns.searchable.
	Searchable bool `json:"searchable"`
	// Flag to indicate if this column is orderable (true) or not (false).
	// This is controlled by columns.orderable.
	Orderable bool `json:"orderable"`
	// Search to apply to this specific column.
	Search Search `json:"search"`
}
