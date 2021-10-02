pipeline {
    agent any
    stages {
        stage('Pre') {
            steps {
                sh 'env'
                sh '''
                mkdir -p /var/local/cache/jenkins/$JOB_NAME/go-build
                mkdir -p /var/local/cache/jenkins/$JOB_NAME/go-pkg
                mkdir -p /var/local/cache/jenkins/$JOB_NAME/golangci-lint
                '''
            }
        }
        stage('Build') {
            steps {
                sh '''
                docker run --rm \
                    -w /var/app/ \
                    -v /var/local/cache/jenkins/$JOB_NAME/go-build:/root/.cache/go-build \
                    -v /var/local/cache/jenkins/$JOB_NAME/go-pkg:/go/pkg \
                    -v $WORKSPACE:/var/app/ \
                    -e GOPROXY="https://goproxy.cn,direct" \
                    -e GOSUMDB=off \
                    -e GO111MODULE=on \
                    -e CGO_ENABLED=0 \
                    golang:1.15-alpine \
                    go build -ldflags='-s' -v
                '''
            }
        }
        stage('Lint') {
            steps {
                sh '''
                docker run --rm \
                    -w /var/app/ \
                    -v /var/local/cache/jenkins/$JOB_NAME/go-build:/root/.cache/go-build \
                    -v /var/local/cache/jenkins/$JOB_NAME/golangci-lint:/root/.cache/golangci-lint \
                    -v $WORKSPACE:/var/app/ \
                    -e GOPROXY="https://goproxy.cn,direct" \
                    -e GOSUMDB=off \
                    -e GO111MODULE=on \
                    -e CGO_ENABLED=0 \
                    golangci/golangci-lint:v1.42-alpine \
                    golangci-lint run -v
                '''
            }
        }
        stage('Test') {
            steps {
                sh '''
                docker run --rm \
                    -w /var/app/ \
                    -v /var/local/cache/jenkins/$JOB_NAME/go-build:/root/.cache/go-build \
                    -v /var/local/cache/jenkins/$JOB_NAME/go-pkg:/go/pkg \
                    -v $WORKSPACE:/var/app/ \
                    -e GOPROXY="https://goproxy.cn,direct" \
                    -e GOSUMDB=off \
                    -e GO111MODULE=on \
                    -e CGO_ENABLED=0 \
                    golang:1.15-alpine \
                    go test -count=1 -covermode=atomic -coverprofile=.testCoverage.txt -timeout=2m -v . ./message
                '''
            }
        }
    }
}
