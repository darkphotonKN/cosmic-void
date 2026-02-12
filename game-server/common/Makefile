gen:
	@protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/common/common.proto \
		api/proto/example/example.proto \
		api/proto/auth/auth.proto \
		api/proto/game/game.proto \
		api/proto/stats/stats.proto \
		api/proto/items/items.proto \
		api/proto/payment/payment.proto \
		api/proto/notification/notification.proto \
		api/proto/events/game.proto \
		api/proto/events/auth.proto \
		api/proto/events/item.proto \
