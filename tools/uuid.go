/**
  @author: $(USER)
  @data:$(DATE)
  @note:
**/
package tools

import uuid "github.com/satori/go.uuid"

func CreateUUID() string {
	return uuid.NewV4().String()
}
