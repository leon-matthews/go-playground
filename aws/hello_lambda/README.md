
# AWS Lambda



## Install AWS SAM CLI tool

Provides Docker-based environment to test AWS Lambda code. Per Amazon, SAM is:

    CLI tool to build, test, debug, and deploy Serverless applications using AWS SAM

Download latest ZIP file for you system, verify the checksum in the release notes,
and trust Amazon implicity with your root account.

https://github.com/aws/aws-sam-cli/releases

    $ sha256sum aws-sam-cli-linux-x86_64.zip
    $ unzip aws-sam-cli-linux-x86_64.zip -d sam-installer
    $ cd sam-installer/
    $ sudo ./install 
    $ sam --version
    SAM CLI, version 1.134.0

## AWS Lambda Runtimes for Go

The runtime is the environment in which the code for our AWS Lambda runs.
Because Go generates native code, we use an OS-only runtime. There was a 
Go-specific runtime, `go1.x`, but it was [deprecated by Amazon in 2024](https://aws.amazon.com/blogs/compute/migrating-aws-lambda-functions-from-the-go1-x-runtime-to-the-custom-runtime-on-amazon-linux-2/).

As of Feburary 2025, the recomended runtime is `provided.al2023`, which is
supported until June 2029.
