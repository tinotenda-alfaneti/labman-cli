pipeline {
  agent any

  options {
    timestamps()
  }

  stages {
    stage('Checkout') {
      steps {
        echo 'Fetching source...'
        checkout scm
      }
    }

    stage('Install Go') {
      steps {
        echo 'Installing Go toolchain...'
        sh '''

          GO_VERSION=1.24.0
          ARCH=$(uname -m)
          case "$ARCH" in
            x86_64)  GO_ARCH=amd64 ;;
            aarch64) GO_ARCH=arm64 ;;
            arm64)   GO_ARCH=arm64 ;;
            armv7l)  GO_ARCH=armv6l ;;
            *) echo "Unsupported architecture: $ARCH" && exit 1 ;;
          esac

          curl -sSLo go.tar.gz https://go.dev/dl/go${GO_VERSION}.linux-${GO_ARCH}.tar.gz
          rm -rf "$WORKSPACE/go"
          tar -C "$WORKSPACE" -xzf go.tar.gz
          rm go.tar.gz
        '''
      }
    }

    stage('Unit Tests') {
      steps {
        echo 'Running go test ./...'
        sh '''
          export PATH="$WORKSPACE/go/bin:$PATH"
          go version
          go test ./...
        '''
      }
    }
  }

  post {
    success {
      echo 'Tests passed.'
    }
    failure {
      echo 'Tests failed.'
    }
  }
}
