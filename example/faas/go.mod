module github.com/Astenna/Nubes/example/faas

go 1.19

replace github.com/Astenna/Nubes/lib v0.0.0 => ../../lib

require github.com/Astenna/Nubes/lib v0.0.0

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/uuid v1.3.0 // indirect
)

require (
	github.com/aws/aws-lambda-go v1.37.0 // indirect
	github.com/aws/aws-sdk-go v1.44.179 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)