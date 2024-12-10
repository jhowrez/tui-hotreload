
upgrade-deps:
	go get -u
	go mod tidy

generate-options:
	go run github.com/jhowrez/go-options-generator@latest \
		-in options.yaml \
		-out_go pkg/options/options.gen.go \
		-out_md OPTIONS.md \
		-pkg options
