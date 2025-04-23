package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Define a struct to hold the item data.  Make sure the field
// names start with a capital letter, otherwise the json Marshall
// will not populate them.
type Item struct {
	ID   string `json:"id"`
	Data string `json:"data"`
}

var ddbClient *dynamodb.Client
var tableName string

func init() {
	// Load the table name from the environment variable.
	tableName = os.Getenv("TABLE_NAME")
	if tableName == "" {
		log.Fatal("TABLE_NAME environment variable not set")
	}

	// Initialize the DynamoDB client.
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}
	ddbClient = dynamodb.NewFromConfig(cfg)
}

func HandleRequest(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the item ID from the path parameters.
	id := event.PathParameters["id"]
	if id == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       "Missing 'id' path parameter",
		}, nil
	}

	// Prepare the input for the DynamoDB GetItem operation.
	key := map[string]types.AttributeValue{
		"id": &types.AttributeValueMemberS{Value: id},
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key:       key,
	}

	// Get the item from the DynamoDB table.
	output, err := ddbClient.GetItem(ctx, input)
	if err != nil {
		log.Printf("failed to get item: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Internal server error",
		}, nil
	}

	// Check if the item was found.
	if output.Item == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 404,
			Body:       fmt.Sprintf("Item with id '%s' not found", id),
		}, nil
	}

	// Unmarshal the DynamoDB item into the Item struct.
	var item Item
	err = attributevalue.UnmarshalMap(output.Item, &item)
	if err != nil {
		log.Printf("failed to unmarshal item: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error unmarshalling DynamoDB data",
		}, nil
	}

	// Marshal the item to JSON for the response.
	response, err := json.Marshal(item)
	if err != nil {
		log.Printf("failed to marshal JSON: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       "Error marshalling JSON",
		}, nil
	}

	// Return the response.
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(response),
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}, nil
}

func main() {
	lambda.Start(HandleRequest)
}
