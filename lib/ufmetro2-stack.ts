import * as cdk from 'aws-cdk-lib';
import * as dynamodb from 'aws-cdk-lib/aws-dynamodb';
import * as lambda from 'aws-cdk-lib/aws-lambda';
import * as apigateway from 'aws-cdk-lib/aws-apigateway';
import { Construct } from 'constructs';
import * as path from 'path';

export class Ufmetro2Stack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const table = new dynamodb.Table(this, 'MyTable', {
      tableName: 'items',
      partitionKey: { name: 'id', type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // Define the Go Lambda function.
    const getItemFunction = new lambda.Function(this, 'GetItemFunction', {
      runtime: lambda.Runtime.PROVIDED_AL2, // Use the Go runtime
      handler: 'main',              // The name of the executable file in the ZIP.
      code: lambda.Code.fromAsset(path.join(__dirname, '../lambda/get-item/function.zip')), // Path to the ZIP.
      environment: {
        TABLE_NAME: table.tableName,
      },
    });

    // Grant the Lambda function read access to the DynamoDB table.  For a GET,
    // we only need read access.
    table.grantReadData(getItemFunction);

    // Set up API Gateway.
    const api = new apigateway.RestApi(this, 'MyApi', {
      restApiName: 'My API',
    });

    const itemsResource = api.root.addResource('items');
    const itemResource = itemsResource.addResource('{id}'); // path parameter

    // Integrate the Lambda function with the GET method.
    itemResource.addMethod('GET', new apigateway.LambdaIntegration(getItemFunction));
  }
}
