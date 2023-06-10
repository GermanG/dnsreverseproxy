env GOOS=linux GOARCH=amd64 go build -o dnsreverseproxy ./main.go
env GOOS=linux GOARCH=arm go build -o dnsreverseproxy.arm ./main.go
