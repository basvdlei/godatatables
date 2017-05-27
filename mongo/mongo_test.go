package mongo

import (
	"net/http"

	"gopkg.in/mgo.v2"
)

func ExampleCollectionHandler() {
	session, _ := mgo.Dial("mymongohost")
	c := session.DB("mydb").C("mycollection")
	http.Handle("/mycollection", &CollectionHandler{Collection: c})
	http.ListenAndServe(":8080", nil)
}
