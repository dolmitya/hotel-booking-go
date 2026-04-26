package room

type Request struct {
	Floor    int    `json:"floor" binding:"min=0" example:"3"`
	Number   string `json:"number" binding:"required" example:"305"`
	Capacity int    `json:"capacity" binding:"min=1" example:"2"`
}
