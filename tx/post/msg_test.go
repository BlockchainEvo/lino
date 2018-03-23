package post

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	acc "github.com/lino-network/lino/tx/account"
	"github.com/stretchr/testify/assert"
)

func testDonationValidate(t *testing.T, donateMsg DonateMsg, expectError sdk.Error) {
	result := donateMsg.ValidateBasic()
	assert.Equal(t, result, expectError)
}

func testLikeValidate(t *testing.T, likeMsg LikeMsg, expectError sdk.Error) {
	result := likeMsg.ValidateBasic()
	assert.Equal(t, result, expectError)
}

func TestCreatePostMsg(t *testing.T) {
	author := acc.AccountKey("TestAuthor")
	// test valid post
	post := PostInfo{
		PostID:       "TestPostID",
		Title:        string(make([]byte, 50)),
		Content:      string(make([]byte, 1000)),
		Author:       author,
		ParentAuthor: "",
		ParentPostID: "",
		SourceAuthor: "",
		SourcePostID: "",
	}
	createMsg := NewCreatePostMsg(post)
	result := createMsg.ValidateBasic()
	assert.Nil(t, result)

	// test missing post id
	post.PostID = ""

	createMsg = NewCreatePostMsg(post)
	result = createMsg.ValidateBasic()
	assert.Equal(t, result, ErrPostCreateNoPostID())

	post.Author = ""
	post.PostID = "testPost"
	createMsg = NewCreatePostMsg(post)
	result = createMsg.ValidateBasic()
	assert.Equal(t, result, ErrPostCreateNoAuthor())

	// test exceeding max title length
	post.Author = author
	post.Title = string(make([]byte, 51))
	createMsg = NewCreatePostMsg(post)
	result = createMsg.ValidateBasic()
	assert.Equal(t, result, ErrPostTitleExceedMaxLength())

	// test exceeding max content length
	post.Title = string(make([]byte, 50))
	post.Content = string(make([]byte, 1001))
	createMsg = NewCreatePostMsg(post)
	result = createMsg.ValidateBasic()
	assert.Equal(t, result, ErrPostContentExceedMaxLength())
}

func TestLikeMsg(t *testing.T) {
	cases := []struct {
		likeMsg     LikeMsg
		expectError sdk.Error
	}{
		{NewLikeMsg(acc.AccountKey("test"), 10000, acc.AccountKey("author"), "postID"), nil},
		{NewLikeMsg(acc.AccountKey("test"), -10000, acc.AccountKey("author"), "postID"), nil},
		{NewLikeMsg(acc.AccountKey("test"), 10001, acc.AccountKey("author"), "postID"), ErrPostLikeWeightOverflow(10001)},
		{NewLikeMsg(acc.AccountKey("test"), -10001, acc.AccountKey("author"), "postID"), ErrPostLikeWeightOverflow(-10001)},
		{NewLikeMsg(acc.AccountKey(""), 10000, acc.AccountKey("author"), "postID"), ErrPostLikeNoUsername()},
		{NewLikeMsg(acc.AccountKey("test"), 10000, acc.AccountKey(""), "postID"), ErrPostLikeInvalidTarget()},
		{NewLikeMsg(acc.AccountKey("test"), 10000, acc.AccountKey("author"), ""), ErrPostLikeInvalidTarget()},
		{NewLikeMsg(acc.AccountKey("test"), 10000, acc.AccountKey(""), ""), ErrPostLikeInvalidTarget()},
	}

	for _, cs := range cases {
		testLikeValidate(t, cs.likeMsg, cs.expectError)
	}
}

func TestDonationMsg(t *testing.T) {
	cases := []struct {
		donateMsg   DonateMsg
		expectError sdk.Error
	}{
		{NewDonateMsg(acc.AccountKey("test"), newAmount(1), acc.AccountKey("author"), "postID"), nil},
		{NewDonateMsg(acc.AccountKey(""), newAmount(1), acc.AccountKey("author"), "postID"), ErrPostDonateNoUsername()},
		{NewDonateMsg(acc.AccountKey("test"), newAmount(0), acc.AccountKey("author"), "postID"), sdk.ErrInvalidCoins("0lino")},
		{NewDonateMsg(acc.AccountKey("test"), newAmount(-1), acc.AccountKey("author"), "postID"), sdk.ErrInvalidCoins("-1lino")},
		{NewDonateMsg(acc.AccountKey("test"), newAmount(1), acc.AccountKey("author"), ""), ErrPostDonateInvalidTarget()},
		{NewDonateMsg(acc.AccountKey("test"), newAmount(1), acc.AccountKey(""), "postID"), ErrPostDonateInvalidTarget()},
		{NewDonateMsg(acc.AccountKey("test"), newAmount(1), acc.AccountKey(""), ""), ErrPostDonateInvalidTarget()},
	}

	for _, cs := range cases {
		testDonationValidate(t, cs.donateMsg, cs.expectError)
	}
}