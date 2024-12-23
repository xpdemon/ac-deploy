
# Xpdemon-Deploy

**Xpdemon-Deploy** is a powerful CLI tool designed to streamline your Docker workflow by allowing you to build Docker images on a more powerful machine and push them to a remote Docker registry. This tool leverages Docker contexts and registries to manage and deploy your Docker Compose applications efficiently.

## Table of Contents

- [Features](#features)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Managing Docker Contexts](#managing-docker-contexts)
    - [Add a Docker Context](#add-a-docker-context)
    - [List Docker Contexts](#list-docker-contexts)
  - [Managing Docker Registries](#managing-docker-registries)
    - [Add a Docker Registry](#add-a-docker-registry)
    - [Login to a Docker Registry](#login-to-a-docker-registry)
  - [Running the Deployment Flow](#running-the-deployment-flow)
- [Example Workflow](#example-workflow)
- [License](#license)

## Features

- **Build on Powerful Machines**: Leverage more capable machines to build your Docker images efficiently.
- **Push to Remote Registries**: Seamlessly push your built images to any Docker registry of your choice.
- **Manage Docker Contexts**: Easily add, list, and manage multiple Docker contexts to switch between different environments.
- **Handle Multiple Registries**: Add and authenticate with multiple Docker registries to cater to diverse deployment needs.
- **Automated Deployment Flow**: Execute a complete deployment flow including building, tagging, pushing, and deploying Docker Compose applications.
- **Configuration Management**: Save and load configurations for Docker contexts and registries to streamline your workflow.

## Prerequisites

- **Go**: Ensure you have Go installed on your machine. You can download it from [here](https://golang.org/dl/).
- **Docker**: Docker must be installed and accessible in your system's PATH. Install Docker from [here](https://docs.docker.com/get-docker/).
- **Docker Compose**: Make sure Docker Compose is installed. Refer to the [Docker Compose installation guide](https://docs.docker.com/compose/install/) if needed.

## Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/xpdemon/ac-deploy.git
   cd ac-deploy
   ```

2. **Build the CLI Tool**

   ```bash
   go build -o xpdemon-deploy main.go
   ```

3. **Move the Executable to Your PATH**

   ```bash
   sudo mv xpdemon-deploy /usr/local/bin/
   ```

   Now, you can use `xpdemon-deploy` from anywhere in your terminal.

## Configuration

Xpdemon-Deploy stores its configuration in `~/.xpdemon-deploy/config.json`. This file manages your Docker contexts and registries.

### Initial Setup

Upon the first run, if the configuration file does not exist, Xpdemon-Deploy will initialize an empty configuration. You can start adding Docker contexts and registries using the provided commands.

## Usage

Xpdemon-Deploy provides several commands to manage Docker contexts, registries, and execute the deployment flow.

### Managing Docker Contexts

#### Add a Docker Context

To add a new Docker context or register an existing local context:

```bash
xpdemon-deploy add-context
```

You will be prompted to choose whether to create a new context or register an existing one. Follow the interactive prompts to complete the process.

#### List Docker Contexts

To list all Docker contexts present in the configuration and on your machine:

```bash
xpdemon-deploy list-contexts
```

This command displays two sections:
- **Docker Contexts in the Application Config**: Contexts registered within Xpdemon-Deploy.
- **Docker Contexts Detected on the Machine**: All Docker contexts available on your local machine.

### Managing Docker Registries

#### Add a Docker Registry

To add a new Docker registry to your configuration:

```bash
xpdemon-deploy add-registry
```

You will be prompted to enter the URL or hostname of the Docker registry (e.g., `docker.io/myuser`).

#### Login to a Docker Registry

To authenticate with an existing Docker registry:

```bash
xpdemon-deploy login-registry
```

Select the registry you wish to log in to from the list of available registries. You will be prompted to enter your username and password/token.

### Running the Deployment Flow

The `run-flow` command executes the complete deployment process, including selecting contexts, building images, pushing to registries, and deploying your Docker Compose applications.

```bash
xpdemon-deploy run-flow
```

**Steps Involved:**

1. **Select Build and Deploy Contexts**: Choose the Docker contexts for building and deploying.
2. **Select Registry (Optional)**: Optionally select a Docker registry to push your images.
3. **Specify docker-compose.yml Path**: Provide the path to your `docker-compose.yml` file.
4. **Parse and Tag Images**: Detect images in the `docker-compose.yml`, apply tags and prefixes if desired.
5. **Prune Docker Environment (Optional)**: Optionally clean up unused Docker images and builder cache.
6. **Build Images**: Build the Docker images using the specified context.
7. **Push Images (Optional)**: Push the built images to the selected Docker registry.
8. **Deploy Images**: Deploy the Docker Compose application in no-build mode.
9. **Cleanup (Optional)**: Optionally delete temporary `docker-compose` files created during the process.

## Example Workflow

Here's an example of how you might use Xpdemon-Deploy in a typical workflow:

1. **Add a Docker Context**

   ```bash
   xpdemon-deploy add-context
   ```

   Choose to create a new context, provide the context name, description, and Docker host.

2. **Add a Docker Registry**

   ```bash
   xpdemon-deploy add-registry
   ```

   Enter your Docker registry URL (e.g., `docker.io/myuser`).

3. **Login to the Docker Registry**

   ```bash
   xpdemon-deploy login-registry
   ```

   Select your registry and provide your credentials.

4. **Run the Deployment Flow**

   ```bash
   xpdemon-deploy run-flow
   ```

   Follow the interactive prompts to build, tag, push, and deploy your Docker Compose application.

## License

This project is licensed under the [MIT License](LICENSE).

---
