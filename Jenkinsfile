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

    stage('Unit Tests') {
      steps {
        echo 'Running go test ./...'
        sh '''
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
