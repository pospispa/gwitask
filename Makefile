
all:
	go build -o gwitask
	docker build -t gwitask .

clean:
	rm gwitask
