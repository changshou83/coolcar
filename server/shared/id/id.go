package id

// AccountID defines account id Object.
type AccountID string

func (a AccountID) String() string {
	return string(a)
}

// TripID defines trip id Object.
type TripID string

func (t TripID) String() string {
	return string(t)
}

// CarID defines car id Object.
type CarID string

func (c CarID) String() string {
	return string(c)
}

// BlobID defines blob id Object.
type BlobID string

func (b BlobID) String() string {
	return string(b)
}
