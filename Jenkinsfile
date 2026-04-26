pipeline {
    agent any

    environment {
        PROJECTNAME = 'cassocial'
        PROJECTORG = 'casapps'
        REGISTRY = "ghcr.io/${env.PROJECTORG}/${env.PROJECTNAME}"
        CGO_ENABLED = '0'
    }

    stages {
        stage('Build') {
            steps {
                script {
                    sh '''
                        VERSION=$(cat release.txt 2>/dev/null || echo "0.1.0")
                        BUILD_DATE=$(date +"%a %b %d, %Y at %H:%M:%S %Z")
                        COMMIT_ID=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

                        LDFLAGS="-s -w -X 'main.Version=${VERSION}' -X 'main.CommitID=${COMMIT_ID}' -X 'main.BuildDate=${BUILD_DATE}'"

                        mkdir -p binaries

                        # Build all platforms
                        for platform in linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 freebsd/amd64 freebsd/arm64; do
                            OS=${platform%/*}
                            ARCH=${platform#*/}
                            OUTPUT=binaries/${PROJECTNAME}-${OS}-${ARCH}
                            [ "$OS" = "windows" ] && OUTPUT=${OUTPUT}.exe

                            echo "Building ${OS}/${ARCH}..."
                            docker run --rm -v $(pwd):/build -w /build -e CGO_ENABLED=0 golang:alpine \
                                sh -c "GOOS=${OS} GOARCH=${ARCH} go build -ldflags \\"${LDFLAGS}\\" -o ${OUTPUT} ./src/main.go"
                        done
                    '''
                }
            }
        }

        stage('Test') {
            steps {
                sh 'docker run --rm -v $(pwd):/build -w /build golang:alpine go test -v -cover ./...'
            }
        }

        stage('Docker') {
            when {
                anyOf {
                    branch 'main'
                    branch 'master'
                    branch 'beta'
                    tag pattern: 'v*', comparator: 'REGEXP'
                    tag pattern: '*.*.*', comparator: 'REGEXP'
                }
            }
            steps {
                script {
                    sh '''
                        VERSION=$(cat release.txt 2>/dev/null || echo "dev")
                        BUILD_DATE=$(date +"%a %b %d, %Y at %H:%M:%S %Z")
                        COMMIT_ID=$(git rev-parse --short HEAD)

                        docker buildx create --name ${PROJECTNAME}-builder --use 2>/dev/null || \
                            docker buildx use ${PROJECTNAME}-builder

                        docker buildx build \
                            -f docker/Dockerfile \
                            --platform linux/amd64,linux/arm64 \
                            --build-arg VERSION="${VERSION}" \
                            --build-arg BUILD_DATE="${BUILD_DATE}" \
                            --build-arg COMMIT_ID="${COMMIT_ID}" \
                            -t ${REGISTRY}:${VERSION} \
                            -t ${REGISTRY}:latest \
                            --push \
                            .
                    '''
                }
            }
        }
    }

    post {
        always {
            cleanWs()
        }
    }
}
