package post

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/store"
	sdk "github.com/cosmos/cosmos-sdk/types"
	acc "github.com/lino-network/lino/tx/account"
	"github.com/stretchr/testify/assert"
	abci "github.com/tendermint/abci/types"
	dbm "github.com/tendermint/tmlibs/db"
)

// Construct some global addrs and txs for tests.
var (
	TestKVStoreKey = sdk.NewKVStoreKey("post")
)

func newPostManager() PostManager {
	return NewPostMananger(TestKVStoreKey)
}

func getContext() sdk.Context {
	db := dbm.NewMemDB()
	ms := store.NewCommitMultiStore(db)
	ms.MountStoreWithDB(TestKVStoreKey, sdk.StoreTypeIAVL, db)
	ms.LoadLatestVersion()

	return sdk.NewContext(ms, abci.Header{}, false, nil)
}

func TestCreatePost(t *testing.T) {
	pm := newPostManager()
	ctx := getContext()

	// test valid post
	post := types.Post{
		PostID:       "TestPostID",
		Title:        string(make([]byte, 50)),
		Content:      string(make([]byte, 1000)),
		Author:       types.AccountKey("test"),
		ParentAuthor: "",
		ParentPostID: "",
		SourceAuthor: "",
		SourcePostID: "",
	}
	postKey := types.GetPostKey(post.Author, post.PostID)
	err := pm.CreatePost(ctx, &post)
	assert.Nil(t, err)

	postPtr, err := pm.GetPost(ctx, postKey)
	assert.Nil(t, err)
	assert.Equal(t, post, *postPtr, "post should be equal")

	postMeta := types.PostMeta{
		LastUpdate:   0,
		LastActivity: 0,
		AllowReplies: true,
	}

	postMetaPtr, err := pm.GetPostMeta(ctx, postKey)
	assert.Nil(t, err)
	assert.Equal(t, postMeta, *postMetaPtr, "Post meta should be equal")

	postLikes := types.PostLikes{Likes: []types.Like{}}
	postLikesPtr, err := pm.GetPostLikes(ctx, postKey)
	assert.Nil(t, err)
	assert.Equal(t, postLikes, *postLikesPtr, "Post like list should be equal")

	postComments := types.PostComments{Comments: []types.PostKey{}}
	postCommentsPtr, err := pm.GetPostComments(ctx, postKey)
	assert.Nil(t, err)
	assert.Equal(t, postComments, *postCommentsPtr, "Post comments should be equal")

	postViews := types.PostViews{Views: []types.View{}}
	postViewsPtr, err := pm.GetPostViews(ctx, postKey)
	assert.Nil(t, err)
	assert.Equal(t, postViews, *postViewsPtr, "Post views should be equal")

	postDonations := types.PostDonations{Donations: []types.Donation{}, Reward: sdk.Coins{}}
	postDonationsPtr, err := pm.GetPostDonations(ctx, postKey)
	assert.Nil(t, err)
	assert.Equal(t, postDonations, *postDonationsPtr, "Post donations should be equal")

	// test recreate post
	err = pm.CreatePost(ctx, &post)
	assert.Equal(t, err, ErrPostExist())

	// test comment post
	post.ParentAuthor = post.Author
	post.ParentPostID = post.PostID
	post.PostID = "commentPost"

	err = pm.CreatePost(ctx, &post)
	assert.Nil(t, err)
	postComments = types.PostComments{Comments: []types.PostKey{types.GetPostKey(post.Author, post.PostID)}}
	postCommentsPtr, err = pm.GetPostComments(ctx, postKey)
	assert.Nil(t, err)
	assert.Equal(t, postComments, *postCommentsPtr, "Post comments should be equal")

	// test invalid comment post
	post.ParentPostID = "invalid"
	post.PostID = "invalid"

	err = pm.CreatePost(ctx, &post)
	assert.Equal(t, err, ErrPostCommentsNotFound(types.GetPostKey(post.ParentAuthor, post.ParentPostID)))
}

func TestPost(t *testing.T) {
	pm := newPostManager()
	ctx := getContext()

	postInfo := types.PostInfo{
		PostID:       "Test Post",
		Title:        "Test Post",
		Content:      "Test Post",
		Author:       types.AccountKey("author"),
		ParentAuthor: "",
		ParentPostID: "",
		SourceAuthor: "",
		SourcePostID: "",
	}
	err := pm.SetPostInfo(ctx, &postInfo)
	assert.Nil(t, err)

	resultPtr, err := pm.GetPostInfo(ctx, GetPostKey(postInfo.Author, postInfo.PostID))
	assert.Nil(t, err)
	assert.Equal(t, postInfo, *resultPtr, "postInfo should be equal")
}

func TestPostMeta(t *testing.T) {
	pm := newPostManager()
	ctx := getContext()

	postMeta := PostMeta{
		AllowReplies: false,
	}
	err := pm.SetPostMeta(ctx, PostKey("test"), &postMeta)
	assert.Nil(t, err)

	resultPtr, err := pm.GetPostMeta(ctx, PostKey("test"))
	assert.Nil(t, err)
	assert.Equal(t, postMeta, *resultPtr, "Post meta should be equal")
}

func TestPostLikes(t *testing.T) {
	pm := newPostManager()
	ctx := getContext()

	postLikes := PostLikes{Likes: []Like{}}
	err := pm.SetPostLikes(ctx, PostKey("test"), &postLikes)
	assert.Nil(t, err)

	resultPtr, err := pm.GetPostLikes(ctx, PostKey("test"))
	assert.Nil(t, err)
	assert.Equal(t, postLikes, *resultPtr, "Post like list should be equal")
}

func TestPostComments(t *testing.T) {
	pm := newPostManager()
	ctx := getContext()

	postComments := PostComments{Comments: []PostKey{}}
	err := pm.SetPostComments(ctx, PostKey("test"), &postComments)
	assert.Nil(t, err)

	resultPtr, err := pm.GetPostComments(ctx, PostKey("test"))
	assert.Nil(t, err)
	assert.Equal(t, postComments, *resultPtr, "Post comments should be equal")
}

func TestPostViews(t *testing.T) {
	pm := newPostManager()
	ctx := getContext()

	postViews := PostViews{Views: []View{}}
	err := pm.SetPostViews(ctx, PostKey("test"), &postViews)
	assert.Nil(t, err)

	resultPtr, err := pm.GetPostViews(ctx, PostKey("test"))
	assert.Nil(t, err)
	assert.Equal(t, postViews, *resultPtr, "Post views should be equal")
}

func TestPostDonate(t *testing.T) {
	pm := newPostManager()
	ctx := getContext()

	postDonations := PostDonations{Donations: []Donation{}, Reward: sdk.Coins{}}
	err := pm.SetPostDonations(ctx, PostKey("test"), &postDonations)
	assert.Nil(t, err)

	resultPtr, err := pm.GetPostDonations(ctx, PostKey("test"))
	assert.Nil(t, err)
	assert.Equal(t, postDonations, *resultPtr, "Post donations should be equal")
}
