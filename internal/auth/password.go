package auth

import (
 "fmt"
 "time"
 "xkeen-panel/internal/models"
 "golang.org/x/crypto/bcrypt"
)

// CreatePasswordUser creates the first account under one lock, without TOTP.
func (um *UserManager) CreatePasswordUser(username, password string) error {
 um.mu.Lock(); defer um.mu.Unlock()
 if um.user != nil { return fmt.Errorf("account already exists") }
 hash,err:=bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost); if err!=nil { return err }
 secret,err:=generateRandomKey(32); if err!=nil { return err }
 um.user=&models.User{Username:username,PasswordHash:string(hash),JWTSecret:secret,CreatedAt:time.Now()}
 if err:=um.persistLocked(); err!=nil { um.user=nil; return err }
 um.pendingSetup=nil
 return nil
}
