package models

import (
	"database/sql"

	"github.com/techjanitor/pram-get/config"
	e "github.com/techjanitor/pram-get/errors"
	u "github.com/techjanitor/pram-get/utils"
)

// FavoritesModel holds the parameters from the request and also the key for the cache
type FavoritesModel struct {
	User   uint
	Page   uint
	Result FavoritesType
}

// IndexType is the top level of the JSON response
type FavoritesType struct {
	Body u.PagedResponse `json:"favorites"`
}

// Header for tag page
type FavoritesHeader struct {
	Images []FavoritesImage `json:"images,omitempty"`
}

// Image struct for tag page
type FavoritesImage struct {
	Id          uint    `json:"id"`
	Thumb       *string `json:"thumbnail"`
	ThumbHeight *uint   `json:"tn_height"`
	ThumbWidth  *uint   `json:"tn_width"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *FavoritesModel) Get() (err error) {

	// Initialize response header
	response := FavoritesType{}

	// Initialize struct for pagination
	paged := u.PagedResponse{}
	// Set current page to parameter
	paged.CurrentPage = i.Page
	// Set tags per index page to config setting
	paged.PerPage = config.Settings.Limits.PostsPerPage

	// Initialize struct for tag info
	tagheader := FavoritesHeader{}

	// Get Database handle
	db, err := u.GetDb()
	if err != nil {
		return
	}

	// Get total tag count
	// Get total tag count and put it in pagination struct
	err = db.QueryRow(`select count(*) from favorites where user_id = ?`, i.User).Scan(&paged.Total)
	if err != nil {
		return
	}

	// Calculate Limit and total Pages
	paged.Get()

	// Return 404 if page requested is larger than actual pages
	if i.Page > paged.Pages {
		return e.ErrNotFound
	}

	// Set perpage and limit to total and 0 if page num is 0
	if i.Page == 0 {
		paged.PerPage = paged.Total
		paged.Limit = 0
	}

	rows, err := db.Query(`SELECT images.image_id,image_thumbnail,image_tn_height,image_tn_width 
	FROM favorites
	LEFT JOIN images on favorites.image_id = images.image_id
	WHERE favorites.user_id = ? LIMIT ?,?`, i.User, paged.Limit, paged.PerPage)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		// Initialize posts struct
		image := FavoritesImage{}
		// Scan rows and place column into struct
		err := rows.Scan(&image.Id, &image.Thumb, &image.ThumbHeight, &image.ThumbWidth)
		if err != nil {
			return err
		}
		// Append rows to info struct
		tagheader.Images = append(tagheader.Images, image)
	}
	err = rows.Err()
	if err != nil {
		return
	}

	// Add tags slice to items interface
	paged.Items = tagheader

	// Add pagedresponse to the response struct
	response.Body = paged

	// This is the data we will serialize
	i.Result = response

	return

}
