target: start

tailwind-build:
	npx tailwindcss -i ./styles.css -o ./public/index.css --minify

tailwind-watch:
	npx tailwindcss -i ./styles.css -o ./public/index.css --watch

start: 
	go run main.go

build: tailwind-build
	go build -o ./build main.go
