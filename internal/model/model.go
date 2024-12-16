package model

import "time"

type Wallet struct {
    ID        string    `json:"id"`        
    Balance   int       `json:"balance"`   
    CreatedAt time.Time `json:"created_at"` 
    UpdatedAt time.Time `json:"updated_at"` 
}
