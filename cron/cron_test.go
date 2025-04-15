package cron

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestExecFunc(t *testing.T) {
	count := 0
	err := AddFunc("aa", time.Second, func() {
		t.Log("aa", time.Now())
		count++
		if count > 3 {
			go func() {
				RemoveFunc("aa")
				t.Log("exit")
			}()
		}
	})
	assert.Equal(t, nil, err)
	time.Sleep(100 * time.Second)

}

func TestRemoveFunc(t *testing.T) {
	RemoveFunc("aa")
}
