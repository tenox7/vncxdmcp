local:
	docker buildx build --platform linux/arm64 -t tenox7/vncxdmcp:latest --load .

push:
	docker buildx build --platform linux/amd64,linux/arm64 -t tenox7/vncxdmcp:latest --push .

clean:
	docker rmi -f tenox7/vncxdmcp:latest
	docker builder prune -a -f
	docker buildx prune -a -f
