package models

import (
	u "github.com/techjanitor/pram-get/utils"
)

// FavoriteModel holds the parameters from the request and also the key for the cache
type FavoriteModel struct {
	User   uint
	Id     uint
	Result FavoriteType
}

// IndexType is the top level of the JSON response
type FavoriteType struct {
	Starred bool `json:"starred"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *FavoriteModel) Get() (err error) {

	// Initialize response header
	response := FavoriteType{}

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	// see if a user has starred an image
	err = db.QueryRow("select count(*) from favorites where user_id = ? AND image_id = ? LIMIT 1", i.User, i.Id).Scan(&response.Starred)
	if err != nil {
		return
	}

	// This is the data we will serialize
	i.Result = response

	return

}