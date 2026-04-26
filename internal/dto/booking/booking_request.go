package booking

type Request struct {
	GuestIDs  []string `json:"guestIds" binding:"required,min=1,dive,uuid" example:"3fa85f64-5717-4562-b3fc-2c963f66afa6"`
	RoomID    string   `json:"roomId" binding:"required,uuid" example:"9a6c1f90-4d3b-4e7c-8e8a-1f23a1e7a123"`
	StartTime string   `json:"startTime" binding:"required" example:"2026-03-02T10:00:00Z"`
	EndTime   string   `json:"endTime" binding:"required" example:"2026-03-04T11:00:00Z"`
}
