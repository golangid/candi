build-activity:
	cd packages/activity-service
	go build -o activity

build-order:
	cd packages/order-service
	go build packages/order-service -o order