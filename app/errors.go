package app

import "errors"

var (
	ErrDatabaseOperation = errors.New("database operation failed")
	ErrCouldNotMakePost  = errors.New("could not send message to telegram channel")
	ErrSummarizerFailed  = errors.New("summarizer failed to generate summary")
	ErrGetContentFailed  = errors.New("failed to get content from article URL")
	ErrFirstFetchFailed  = errors.New("failed to fetch feeds from database on first attempt")
)
