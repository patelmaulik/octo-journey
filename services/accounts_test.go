package service

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
	"patelmaulik.com/maulik/v1/dbclient"
	"patelmaulik.com/maulik/v1/models"
)

var Sf = fmt.Sprintf

// TestWrongPath - testing 404
func TestWrongPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := dbclient.NewMockIDatabaseRepository(ctrl)

	s := testSetup(mockRepo)

	req := httptest.NewRequest("GET", Sf("%v", "/invalid/1002"), nil)
	res := httptest.NewRecorder()

	s.r.ServeHTTP(res, req)

	assert.Equal(t, 404, res.Code)
}

func TestGetAccountPath(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	//defer gock.Off()

	// gomock.Any()
	mockRepo := dbclient.NewMockIDatabaseRepository(ctrl)
	mockRepo.EXPECT().
		QueryAccount("1002").
		Return(models.Account{Id: "1002", Name: "Person_1002"}, nil).
		Times(1)

	s := testSetup(mockRepo)

	var serverURL = "http://localhost:7001"

	gock.New(serverURL).
		Get("/account/1002").
		Reply(200).
		BodyString(`{"ID":"1002", "Name":"Person_1002", "ServedBy":"localhost"}`)

	req := httptest.NewRequest("GET", Sf("%v", "/account/1002"), nil)
	req.Header.Set("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ")
	//ctx := req.Context() 	//req = req.WithContext(ctx)

	res := httptest.NewRecorder()

	s.r.ServeHTTP(res, req)

	assert.Equal(t, 200, res.Code)

	account := models.Account{}
	json.Unmarshal(res.Body.Bytes(), &account)

	assert.Equal(t, account.Id, "1002")
	assert.Equal(t, account.Name, "Person_1002")
}

func testSetup(d dbclient.IDatabaseRepository) *Server {
	s := NewServer()
	h := AccountHandler{}

	h.DBClient = d
	s.SetupRoutes(&h)

	return s
}
