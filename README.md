# Flutter APK Builder CLI using Bubble Tea

This project provider a command-linen interface (CLI) build with Golang and Bubble Tea, allowing you to easily build and upload APKs for different flavors and build modes.

## Features

- Build APKs for specific flavors (e.g., dev, raf, wellcare).
- Choose between debug and release build modes.
- View real-time command output and progress.
- Automatically compress and upload generated APKs to a transfer service.

## Getting Started

1. Clone the repository

```bash
git clone https://github.com/your-username/flutter-apk-builder.git
```

2. Build the app

```bash
go mod tidy
go build
```

3. Run the app

```bash
./apk_builder_v3
```

4. Follow the on-screen instructions to select your desired flavor and build mode.

5. The app will automatically build, zip, and upload your APK.

## Requirements

- Go programming language installed.
- Flutter SDK installed.

## Usage

The app presents you with two interactive lists:

- _Flavors_: Choose the flavor for which you want to build the APK.
- _Build Modes_: Select either `debug` or `release` mode.
  Once you've chosen both options, the build process starts. The app will display the output of the build commands in real-time, along with the progress. After successful completion, it will notify you of the upload link.

## Additional Notes:

- You can replace the transfer service used for uploading with your preferred platform.
- The app currently assumes certain directory structures for your Flutter project. Make sure your project layout matches the expectations.
- Feel free to adopt and extend this codebase to fit your specific needs and workflows.
