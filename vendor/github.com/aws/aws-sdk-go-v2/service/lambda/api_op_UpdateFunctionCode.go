// Code generated by smithy-go-codegen DO NOT EDIT.

package lambda

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Updates a Lambda function's code. If code signing is enabled for the function,
// the code package must be signed by a trusted publisher. For more information,
// see [Configuring code signing for Lambda].
//
// If the function's package type is Image , then you must specify the code package
// in ImageUri as the URI of a [container image] in the Amazon ECR registry.
//
// If the function's package type is Zip , then you must specify the deployment
// package as a [.zip file archive]. Enter the Amazon S3 bucket and key of the code .zip file
// location. You can also provide the function code inline using the ZipFile field.
//
// The code in the deployment package must be compatible with the target
// instruction set architecture of the function ( x86-64 or arm64 ).
//
// The function's code is locked when you publish a version. You can't modify the
// code of a published version, only the unpublished version.
//
// For a function defined as a container image, Lambda resolves the image tag to
// an image digest. In Amazon ECR, if you update the image tag to a new image,
// Lambda does not automatically update the function.
//
// [.zip file archive]: https://docs.aws.amazon.com/lambda/latest/dg/gettingstarted-package.html#gettingstarted-package-zip
// [Configuring code signing for Lambda]: https://docs.aws.amazon.com/lambda/latest/dg/configuration-codesigning.html
// [container image]: https://docs.aws.amazon.com/lambda/latest/dg/lambda-images.html
func (c *Client) UpdateFunctionCode(ctx context.Context, params *UpdateFunctionCodeInput, optFns ...func(*Options)) (*UpdateFunctionCodeOutput, error) {
	if params == nil {
		params = &UpdateFunctionCodeInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "UpdateFunctionCode", params, optFns, c.addOperationUpdateFunctionCodeMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*UpdateFunctionCodeOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type UpdateFunctionCodeInput struct {

	// The name or ARN of the Lambda function.
	//
	// Name formats
	//
	//   - Function name – my-function .
	//
	//   - Function ARN – arn:aws:lambda:us-west-2:123456789012:function:my-function .
	//
	//   - Partial ARN – 123456789012:function:my-function .
	//
	// The length constraint applies only to the full ARN. If you specify only the
	// function name, it is limited to 64 characters in length.
	//
	// This member is required.
	FunctionName *string

	// The instruction set architecture that the function supports. Enter a string
	// array with one of the valid values (arm64 or x86_64). The default value is
	// x86_64 .
	Architectures []types.Architecture

	// Set to true to validate the request parameters and access permissions without
	// modifying the function code.
	DryRun bool

	// URI of a container image in the Amazon ECR registry. Do not use for a function
	// defined with a .zip file archive.
	ImageUri *string

	// Set to true to publish a new version of the function after updating the code.
	// This has the same effect as calling PublishVersionseparately.
	Publish bool

	// Update the function only if the revision ID matches the ID that's specified.
	// Use this option to avoid modifying a function that has changed since you last
	// read it.
	RevisionId *string

	// An Amazon S3 bucket in the same Amazon Web Services Region as your function.
	// The bucket can be in a different Amazon Web Services account. Use only with a
	// function defined with a .zip file archive deployment package.
	S3Bucket *string

	// The Amazon S3 key of the deployment package. Use only with a function defined
	// with a .zip file archive deployment package.
	S3Key *string

	// For versioned objects, the version of the deployment package object to use.
	S3ObjectVersion *string

	// The ARN of the Key Management Service (KMS) customer managed key that's used to
	// encrypt your function's .zip deployment package. If you don't provide a customer
	// managed key, Lambda uses an Amazon Web Services managed key.
	SourceKMSKeyArn *string

	// The base64-encoded contents of the deployment package. Amazon Web Services SDK
	// and CLI clients handle the encoding for you. Use only with a function defined
	// with a .zip file archive deployment package.
	ZipFile []byte

	noSmithyDocumentSerde
}

// Details about a function's configuration.
type UpdateFunctionCodeOutput struct {

	// The instruction set architecture that the function supports. Architecture is a
	// string array with one of the valid values. The default architecture value is
	// x86_64 .
	Architectures []types.Architecture

	// The SHA256 hash of the function's deployment package.
	CodeSha256 *string

	// The size of the function's deployment package, in bytes.
	CodeSize int64

	// The function's dead letter queue.
	DeadLetterConfig *types.DeadLetterConfig

	// The function's description.
	Description *string

	// The function's [environment variables]. Omitted from CloudTrail logs.
	//
	// [environment variables]: https://docs.aws.amazon.com/lambda/latest/dg/configuration-envvars.html
	Environment *types.EnvironmentResponse

	// The size of the function's /tmp directory in MB. The default value is 512, but
	// can be any whole number between 512 and 10,240 MB. For more information, see [Configuring ephemeral storage (console)].
	//
	// [Configuring ephemeral storage (console)]: https://docs.aws.amazon.com/lambda/latest/dg/configuration-function-common.html#configuration-ephemeral-storage
	EphemeralStorage *types.EphemeralStorage

	// Connection settings for an [Amazon EFS file system].
	//
	// [Amazon EFS file system]: https://docs.aws.amazon.com/lambda/latest/dg/configuration-filesystem.html
	FileSystemConfigs []types.FileSystemConfig

	// The function's Amazon Resource Name (ARN).
	FunctionArn *string

	// The name of the function.
	FunctionName *string

	// The function that Lambda calls to begin running your function.
	Handler *string

	// The function's image configuration values.
	ImageConfigResponse *types.ImageConfigResponse

	// The ARN of the Key Management Service (KMS) customer managed key that's used to
	// encrypt the following resources:
	//
	//   - The function's [environment variables].
	//
	//   - The function's [Lambda SnapStart]snapshots.
	//
	//   - When used with SourceKMSKeyArn , the unzipped version of the .zip deployment
	//   package that's used for function invocations. For more information, see [Specifying a customer managed key for Lambda].
	//
	//   - The optimized version of the container image that's used for function
	//   invocations. Note that this is not the same key that's used to protect your
	//   container image in the Amazon Elastic Container Registry (Amazon ECR). For more
	//   information, see [Function lifecycle].
	//
	// If you don't provide a customer managed key, Lambda uses an [Amazon Web Services owned key] or an [Amazon Web Services managed key].
	//
	// [Amazon Web Services owned key]: https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#aws-owned-cmk
	// [Specifying a customer managed key for Lambda]: https://docs.aws.amazon.com/lambda/latest/dg/encrypt-zip-package.html#enable-zip-custom-encryption
	// [Lambda SnapStart]: https://docs.aws.amazon.com/lambda/latest/dg/snapstart-security.html
	// [environment variables]: https://docs.aws.amazon.com/lambda/latest/dg/configuration-envvars.html#configuration-envvars-encryption
	// [Function lifecycle]: https://docs.aws.amazon.com/lambda/latest/dg/images-create.html#images-lifecycle
	// [Amazon Web Services managed key]: https://docs.aws.amazon.com/kms/latest/developerguide/concepts.html#aws-managed-cmk
	KMSKeyArn *string

	// The date and time that the function was last updated, in [ISO-8601 format]
	// (YYYY-MM-DDThh:mm:ss.sTZD).
	//
	// [ISO-8601 format]: https://www.w3.org/TR/NOTE-datetime
	LastModified *string

	// The status of the last update that was performed on the function. This is first
	// set to Successful after function creation completes.
	LastUpdateStatus types.LastUpdateStatus

	// The reason for the last update that was performed on the function.
	LastUpdateStatusReason *string

	// The reason code for the last update that was performed on the function.
	LastUpdateStatusReasonCode types.LastUpdateStatusReasonCode

	// The function's [layers].
	//
	// [layers]: https://docs.aws.amazon.com/lambda/latest/dg/configuration-layers.html
	Layers []types.Layer

	// The function's Amazon CloudWatch Logs configuration settings.
	LoggingConfig *types.LoggingConfig

	// For Lambda@Edge functions, the ARN of the main function.
	MasterArn *string

	// The amount of memory available to the function at runtime.
	MemorySize *int32

	// The type of deployment package. Set to Image for container image and set Zip
	// for .zip file archive.
	PackageType types.PackageType

	// The latest updated revision of the function or alias.
	RevisionId *string

	// The function's execution role.
	Role *string

	// The identifier of the function's [runtime]. Runtime is required if the deployment
	// package is a .zip file archive. Specifying a runtime results in an error if
	// you're deploying a function using a container image.
	//
	// The following list includes deprecated runtimes. Lambda blocks creating new
	// functions and updating existing functions shortly after each runtime is
	// deprecated. For more information, see [Runtime use after deprecation].
	//
	// For a list of all currently supported runtimes, see [Supported runtimes].
	//
	// [Runtime use after deprecation]: https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html#runtime-deprecation-levels
	// [runtime]: https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html
	// [Supported runtimes]: https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html#runtimes-supported
	Runtime types.Runtime

	// The ARN of the runtime and any errors that occured.
	RuntimeVersionConfig *types.RuntimeVersionConfig

	// The ARN of the signing job.
	SigningJobArn *string

	// The ARN of the signing profile version.
	SigningProfileVersionArn *string

	// Set ApplyOn to PublishedVersions to create a snapshot of the initialized
	// execution environment when you publish a function version. For more information,
	// see [Improving startup performance with Lambda SnapStart].
	//
	// [Improving startup performance with Lambda SnapStart]: https://docs.aws.amazon.com/lambda/latest/dg/snapstart.html
	SnapStart *types.SnapStartResponse

	// The current state of the function. When the state is Inactive , you can
	// reactivate the function by invoking it.
	State types.State

	// The reason for the function's current state.
	StateReason *string

	// The reason code for the function's current state. When the code is Creating ,
	// you can't invoke or modify the function.
	StateReasonCode types.StateReasonCode

	// The amount of time in seconds that Lambda allows a function to run before
	// stopping it.
	Timeout *int32

	// The function's X-Ray tracing configuration.
	TracingConfig *types.TracingConfigResponse

	// The version of the Lambda function.
	Version *string

	// The function's networking configuration.
	VpcConfig *types.VpcConfigResponse

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationUpdateFunctionCodeMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsRestjson1_serializeOpUpdateFunctionCode{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpUpdateFunctionCode{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "UpdateFunctionCode"); err != nil {
		return fmt.Errorf("add protocol finalizers: %v", err)
	}

	if err = addlegacyEndpointContextSetter(stack, options); err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = addClientRequestID(stack); err != nil {
		return err
	}
	if err = addComputeContentLength(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = addComputePayloadSHA256(stack); err != nil {
		return err
	}
	if err = addRetry(stack, options); err != nil {
		return err
	}
	if err = addRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = addRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addSpanRetryLoop(stack, options); err != nil {
		return err
	}
	if err = addClientUserAgent(stack, options); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addSetLegacyContextSigningOptionsMiddleware(stack); err != nil {
		return err
	}
	if err = addTimeOffsetBuild(stack, c); err != nil {
		return err
	}
	if err = addUserAgentRetryMode(stack, options); err != nil {
		return err
	}
	if err = addOpUpdateFunctionCodeValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opUpdateFunctionCode(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRecursionDetection(stack); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	if err = addDisableHTTPSMiddleware(stack, options); err != nil {
		return err
	}
	if err = addSpanInitializeStart(stack); err != nil {
		return err
	}
	if err = addSpanInitializeEnd(stack); err != nil {
		return err
	}
	if err = addSpanBuildRequestStart(stack); err != nil {
		return err
	}
	if err = addSpanBuildRequestEnd(stack); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opUpdateFunctionCode(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "UpdateFunctionCode",
	}
}