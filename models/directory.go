package models

import (
	"database/sql"
	"time"

	"github.com/eirka/eirka-libs/config"
	"github.com/eirka/eirka-libs/db"
	e "github.com/eirka/eirka-libs/errors"

	u "github.com/eirka/eirka-get/utils"
)

// DirectoryModel holds the parameters from the request and also the key for the cache
type DirectoryModel struct {
	Ib     uint
	Page   uint
	Result DirectoryType
}

// DirectoryType is the top level of the JSON response
type DirectoryType struct {
	Body u.PagedResponse `json:"directory"`
}

// Directory holds the thread entries for the directory page
type Directory struct {
	ID     uint      `json:"id"`
	Title  string    `json:"title"`
	Closed bool      `json:"closed"`
	Sticky bool      `json:"sticky"`
	Posts  uint      `json:"postcount"`
	Pages  uint      `json:"pages"`
	Last   time.Time `json:"last_post"`
	Images uint      `json:"images"`
}

// Get will gather the information from the database and return it as JSON serialized data
func (i *DirectoryModel) Get() (err error) {

	if i.Ib == 0 || i.Page == 0 {
		return e.ErrNotFound
	}

	// Initialize response header
	response := DirectoryType{}

	// Initialize struct for pagination
	paged := u.PagedResponse{}
	// Set current page to parameter
	paged.CurrentPage = i.Page
	// Set threads per index page to config setting
	paged.PerPage = config.Settings.Limits.PostsPerPage

	// Get Database handle
	dbase, err := db.GetDb()
	if err != nil {
		return
	}

	// Get total tag count and put it in pagination struct
	err = dbase.QueryRow(`SELECT count(thread_id) FROM threads
    WHERE ib_id = ? AND thread_deleted != 1`, i.Ib).Scan(&paged.Total)
	if err != nil {
		return
	}

	// Calculate Limit and total Pages
	paged.Get()

	// Return 404 if page requested is larger than actual pages
	if i.Page > paged.Pages {
		return e.ErrNotFound
	}

	rows, err := dbase.Query(`SELECT threads.thread_id,thread_title,thread_closed,thread_sticky,count(posts.post_id),count(image_id),
    (select max(post_time) from posts where thread_id=threads.thread_id AND post_deleted != 1) as thread_last_post
    FROM threads
    LEFT JOIN posts on threads.thread_id = posts.thread_id
    LEFT JOIN images on images.post_id = posts.post_id
    WHERE ib_id = ? AND thread_deleted != 1 AND post_deleted != 1
    GROUP BY threads.thread_id
    ORDER BY thread_sticky = 1 DESC, thread_last_post DESC LIMIT ?,?`, i.Ib, paged.Limit, paged.PerPage)
	if err != nil {
		return err
	}
	defer rows.Close()

	threads := []Directory{}
	for rows.Next() {
		thread := Directory{}
		var lastPost sql.NullTime
		err = rows.Scan(&thread.ID, &thread.Title, &thread.Closed, &thread.Sticky, &thread.Posts, &thread.Images, &lastPost)
		if err != nil {
			rows.Close() // Explicitly close rows before returning
			return err
		}

		thread.Last = lastPost.Time

		// Get the number of pages in the thread
		postpages := u.PagedResponse{}
		postpages.Total = thread.Posts
		postpages.CurrentPage = 1
		postpages.PerPage = config.Settings.Limits.PostsPerPage
		postpages.Get()

		// set pages
		thread.Pages = postpages.Pages

		threads = append(threads, thread)
	}
	if err = rows.Err(); err != nil {
		return err
	}

	// Check if no threads were found
	if len(threads) == 0 {
		return e.ErrNotFound
	}

	// Add threads slice to items interface
	paged.Items = threads

	// Add pagedresponse to the response struct
	response.Body = paged

	// This is the data we will serialize
	i.Result = response

	return

}
