package fs

import "io"

type OpenReadFunction func(start, end int) (io.Reader, error)
