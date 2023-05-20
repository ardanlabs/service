package handlers

/*
	This code can be used to execute the handlers as AWS Lambda functions. It
	takes the lambda request and converts it to a HTTP request and then converts
	the	HTTP response to a lamdba response.

	Instead of having the service run ListenAndSeve, you use this code instead.
	AWS starts the service to handle the request, sends the request,
	and after the response, shuts it down. This allows you to manage all your
	lambda functions as handlers all in one place.


	// -------------------------------------------------------------------------
	// Inside main.go

	import (
		ddlambda "github.com/DataDog/datadog-lambda-go"
		"github.com/aws/aws-lambda-go/lambda"
	)

	log.Infow("startup", "status", "execute lambda request")
	defer log.Infow("shutdown", "status", "shutdown complete")

	l := handlers.Lambda{
		apiMux: apiMux,
	}

	lambda.StartWithOptions(
		ddlambda.WrapFunction(l.ServeLambda, nil),
		lambda.WithEnableSIGTERM(func() {
			log.Infow("shutdown", "status", "function container shutting down...")
		}),
	)

	// -------------------------------------------------------------------------
	// Keep this here in lambda.go

	import (
		"github.com/aws/aws-lambda-go/events"
		lambdaproxy "github.com/awslabs/aws-lambda-go-api-proxy/core"
	)

	// Lambda represents the lambda router.
	type Lambda struct {
		apiMux *web.App
	}

	// ServeLambda routes the lambda request to an internal endpoint.
	func (l *Lambda) ServeLambda(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		var ra lambdaproxy.RequestAccessorV2

		r, err := ra.EventToRequestWithContext(ctx, request)
		if err != nil {
			return events.APIGatewayV2HTTPResponse{}, fmt.Errorf("event to request: %w", err)
		}

		w := lambdaproxy.NewProxyResponseWriterV2()
		l.apiMux.ServeHTTP(w, r)

		return w.GetProxyResponse()
	}
*/
