module rpclambda

go 1.12

require (
	github.com/aws/aws-lambda-go v1.12.0
	github.com/aws/aws-sdk-go v1.21.10
	github.com/davecgh/go-spew v1.1.1
	github.com/tespo/satya/v2 v2.3.1
)

replace github.com/tespo/satya/v2 => ../../satya
