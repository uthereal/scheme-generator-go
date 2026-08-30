package example_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uthereal/scheme-generator-go/example/generated"
	"github.com/uthereal/scheme-generator-go/internal/generator"
	"github.com/uthereal/scheme-runtime-go/pkg/contract"
)

// Test_Integration_DynamicGeneration verifies dynamic generation from DDL.
func Test_Integration_DynamicGeneration(
	t *testing.T,
) {
	err := generator.Run(
		os.DirFS("./migrations"),
		"generated",
		"./generated",
	)
	require.NoError(t, err, "Dynamic code generation failed")
}

// Test_Integration_RealPostgres_KitchenSink executes full integration checks.
func Test_Integration_RealPostgres_KitchenSink(
	t *testing.T,
) {
	ctx := context.Background()
	db, cleanup, err := pgContainer.CreateDB(ctx, t)
	require.NoError(t, err)
	defer cleanup()

	roleQuery := generated.NewAuthRoleQuery(db)
	permQuery := generated.NewAuthPermissionQuery(db)
	userQuery := generated.NewAuthUserQuery(db)
	profQuery := generated.NewAuthProfileQuery(db)
	postQuery := generated.NewContentPostQuery(db)
	tagQuery := generated.NewContentTagQuery(db)
	commentQuery := generated.NewContentCommentQuery(db)
	parentQuery := generated.NewTenantParentQuery(db)
	childQuery := generated.NewTenantChildQuery(db)
	deviceQuery := generated.NewAuthDeviceQuery(db)

	var insertedRole generated.AuthRole
	var insertedPerm generated.AuthPermission
	var insertedUser generated.AuthUser
	var insertedPost generated.ContentPost
	var insertedTag generated.ContentTag
	var insertedParent generated.TenantParent
	var insertedChild generated.TenantChild

	t.Run("Insert Auth roles, permissions, and M:N link", func(
		t *testing.T,
	) {
		r, errRole := roleQuery.InsertReturning(
			ctx,
			generated.AuthRoleMutator{
				Name: contract.Set[string]{
					IsSet: true,
					Value: "Administrator",
				},
			},
		)
		require.NoError(t, errRole)
		assert.Greater(t, *r.ID, int64(0))
		insertedRole = r

		p, errPerm := permQuery.InsertReturning(
			ctx,
			generated.AuthPermissionMutator{
				Code: contract.Set[string]{
					IsSet: true,
					Value: "admin.all",
				},
			},
		)
		require.NoError(t, errPerm)
		assert.Greater(t, *p.ID, int64(0))
		insertedPerm = p

		_, errLink := db.Exec(
			ctx,
			`INSERT INTO auth.role_permissions (role_id, permission_id)
			VALUES ($1, $2);`,
			*r.ID,
			*p.ID,
		)
		require.NoError(t, errLink)
	})

	t.Run("Insert User, Profile (1:1), and Content (Cross-Schema)", func(
		t *testing.T,
	) {
		prefsJSON := []byte(`{"dark_mode":true}`)
		metaJSON := []byte(`{"region":"us"}`)
		u, errUser := userQuery.InsertReturning(
			ctx,
			generated.AuthUserMutator{
				Email: contract.Set[string]{
					IsSet: true,
					Value: "author@example.com",
				},
				Age: contract.Set[*int32]{
					IsSet: true,
					Value: pointerTo[int32](30),
				},
				Tags: contract.Set[[]string]{
					IsSet: true,
					Value: []string{"developer", "writer"},
				},
				Preferences: contract.Set[*[]byte]{
					IsSet: true,
					Value: &prefsJSON,
				},
				Metadata: contract.Set[[]byte]{
					IsSet: true,
					Value: metaJSON,
				},
				Status: contract.Set[string]{
					IsSet: true,
					Value: "active",
				},
				RoleID: contract.Set[*int64]{
					IsSet: true,
					Value: insertedRole.ID,
				},
			},
		)
		require.NoError(t, errUser)
		assert.Greater(t, *u.ID, int64(0))
		insertedUser = u

		bio := "Full-stack developer"
		prof, errProf := profQuery.InsertReturning(
			ctx,
			generated.AuthProfileMutator{
				UserID: contract.Set[int64]{
					IsSet: true,
					Value: *u.ID,
				},
				Biography: contract.Set[*string]{
					IsSet: true,
					Value: &bio,
				},
				IsPublic: contract.Set[*bool]{
					IsSet: true,
					Value: pointerTo[bool](true),
				},
			},
		)
		require.NoError(t, errProf)
		assert.Equal(t, *u.ID, prof.UserID)

		post, errPost := postQuery.InsertReturning(
			ctx,
			generated.ContentPostMutator{
				UserID: contract.Set[int64]{
					IsSet: true,
					Value: *u.ID,
				},
				Title: contract.Set[string]{
					IsSet: true,
					Value: "Cross-Schema DDL in Postgres",
				},
				Content: contract.Set[string]{
					IsSet: true,
					Value: "PostgreSQL multi-schema relationships.",
				},
			},
		)
		require.NoError(t, errPost)
		insertedPost = post

		tag, errTag := tagQuery.InsertReturning(
			ctx,
			generated.ContentTagMutator{
				Slug: contract.Set[string]{
					IsSet: true,
					Value: "postgresql",
				},
			},
		)
		require.NoError(t, errTag)
		insertedTag = tag

		_, errPostTag := db.Exec(
			ctx,
			`INSERT INTO content.post_tags (post_id, tag_id)
			VALUES ($1, $2);`,
			*post.ID,
			*tag.ID,
		)
		require.NoError(t, errPostTag)

		_, errComment := commentQuery.InsertReturning(
			ctx,
			generated.ContentCommentMutator{
				PostID: contract.Set[int64]{
					IsSet: true,
					Value: *post.ID,
				},
				UserID: contract.Set[int64]{
					IsSet: true,
					Value: *u.ID,
				},
				Text: contract.Set[string]{
					IsSet: true,
					Value: "Great article!",
				},
			},
		)
		require.NoError(t, errComment)
	})

	t.Run("Insert Composite Key Parent and Child", func(
		t *testing.T,
	) {
		pTenantID := int32(100)
		pID := int32(1)
		pName := "Engineering Division"
		p, errParent := parentQuery.InsertReturning(
			ctx,
			generated.TenantParentMutator{
				TenantID: contract.Set[*int32]{
					IsSet: true,
					Value: &pTenantID,
				},
				ID: contract.Set[*int32]{
					IsSet: true,
					Value: &pID,
				},
				Name: contract.Set[*string]{
					IsSet: true,
					Value: &pName,
				},
			},
		)
		require.NoError(t, errParent)
		insertedParent = p

		cID := int32(999)
		c, errChild := childQuery.InsertReturning(
			ctx,
			generated.TenantChildMutator{
				ID: contract.Set[*int32]{
					IsSet: true,
					Value: &cID,
				},
				TenantID: contract.Set[*int32]{
					IsSet: true,
					Value: &pTenantID,
				},
				ParentID: contract.Set[*int32]{
					IsSet: true,
					Value: &pID,
				},
			},
		)
		require.NoError(t, errChild)
		insertedChild = c
	})

	t.Run("Eager Load User -> Posts (Cross-Schema HasMany)", func(
		t *testing.T,
	) {
		users, errQuery := userQuery.
			With(generated.Schema.Auth.AuthUser.Posts).
			Where(generated.Schema.Auth.AuthUser.ID.Eq(*insertedUser.ID)).
			Get(ctx)
		require.NoError(t, errQuery)
		require.Len(t, users, 1)
		require.Len(t, users[0].Posts, 1)
		assert.Equal(t, *insertedPost.ID, *users[0].Posts[0].ID)
	})

	t.Run("Eager Load Post -> User (Cross-Schema BelongsTo)", func(
		t *testing.T,
	) {
		posts, errQuery := postQuery.
			With(generated.Schema.Content.ContentPost.AuthUser).
			Where(generated.Schema.Content.ContentPost.ID.Eq(*insertedPost.ID)).
			Get(ctx)
		require.NoError(t, errQuery)
		require.Len(t, posts, 1)
		require.NotNil(t, posts[0].AuthUser)
		assert.Equal(t, insertedUser.Email, posts[0].AuthUser.Email)
	})

	t.Run("Eager Load Post -> Tags (M:N BelongsToMany)", func(
		t *testing.T,
	) {
		posts, errQuery := postQuery.
			With(generated.Schema.Content.ContentPost.Tags).
			Where(generated.Schema.Content.ContentPost.ID.Eq(*insertedPost.ID)).
			Get(ctx)
		require.NoError(t, errQuery)
		require.Len(t, posts, 1)
		require.Len(t, posts[0].Tags, 1)
		assert.Equal(t, insertedTag.Slug, posts[0].Tags[0].Slug)
	})

	t.Run("Eager Load Role -> Permissions (M:N BelongsToMany)", func(
		t *testing.T,
	) {
		roles, errQuery := roleQuery.
			With(generated.Schema.Auth.AuthRole.Permissions).
			Where(generated.Schema.Auth.AuthRole.ID.Eq(*insertedRole.ID)).
			Get(ctx)
		require.NoError(t, errQuery)
		require.Len(t, roles, 1)
		require.Len(t, roles[0].Permissions, 1)
		assert.Equal(
			t,
			insertedPerm.Code,
			roles[0].Permissions[0].Code,
		)
	})

	t.Run("Eager Load Composite Parent -> Children", func(
		t *testing.T,
	) {
		parents, errQuery := parentQuery.
			With(generated.Schema.Tenant.TenantParent.Childrens).
			Where(
				generated.Schema.Tenant.TenantParent.ID.Eq(
					*insertedParent.ID,
				),
			).
			Get(ctx)
		require.NoError(t, errQuery)
		require.Len(t, parents, 1)
		require.Len(t, parents[0].Childrens, 1)
		assert.Equal(t, *insertedChild.ID, *parents[0].Childrens[0].ID)
	})

	t.Run("Insert and Query Auth Device with UUID columns", func(
		t *testing.T,
	) {
		deviceID := uuid.New()
		sessionToken := uuid.New()

		insertedDev, errDev := deviceQuery.InsertReturning(
			ctx,
			generated.AuthDeviceMutator{
				ID: contract.Set[*uuid.UUID]{
					IsSet: true,
					Value: &deviceID,
				},
				UserID: contract.Set[*int32]{
					IsSet: true,
					Value: pointerTo(int32(*insertedUser.ID)),
				},
				SessionToken: contract.Set[*uuid.UUID]{
					IsSet: true,
					Value: &sessionToken,
				},
			},
		)
		require.NoError(t, errDev)
		require.NotNil(t, insertedDev.ID)
		assert.Equal(t, deviceID, *insertedDev.ID)
		require.NotNil(t, insertedDev.SessionToken)
		assert.Equal(t, sessionToken, *insertedDev.SessionToken)

		devices, errQuery := deviceQuery.
			Where(
				generated.Schema.Auth.AuthDevice.ID.Eq(
					deviceID.String(),
				),
			).
			Get(ctx)
		require.NoError(t, errQuery)
		require.Len(t, devices, 1)
		assert.Equal(t, deviceID, *devices[0].ID)
		assert.Equal(t, sessionToken, *devices[0].SessionToken)
	})
}
