gen:
	@protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/common/common.proto \
		api/proto/example/example.proto \
		api/proto/auth/auth.proto \
		api/proto/item/item.proto \
		api/proto/notification/notification.proto \
		api/proto/build/build.proto \
		api/proto/skill/skill.proto \
		api/proto/tag/tag.proto \
		api/proto/class/class.proto \
		api/proto/rating/rating.proto \
