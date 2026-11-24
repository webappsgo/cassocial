pipeline {
    agent none

    triggers {
        cron('0 3 * * *\n0 4 * * *\n0 6 * * 1')
    }

    options {
        disableConcurrentBuilds()
        timeout(time: 90, unit: 'MINUTES')
    }

    environment {
        PROJECT_NAME = 'cassocial'
        PROJECT_ORG = 'casapps'
        INTERNAL_NAME = 'cassocial'
        BINDIR = 'binaries'
        RELDIR = 'releases'
        GO_CACHE = "${JENKINS_HOME}/cache/${PROJECT_ORG}/${PROJECT_NAME}/go-mod"
        GO_BUILD = "${JENKINS_HOME}/cache/${PROJECT_ORG}/${PROJECT_NAME}/go-build"
        REGISTRY = "ghcr.io/${PROJECT_ORG}/${INTERNAL_NAME}"
    }

    stages {
        stage('Setup') {
            agent { label 'amd64' }
            steps {
                script {
                    env.IS_SCHEDULED = currentBuild.getBuildCauses('hudson.triggers.TimerTrigger$TimerTriggerCause').isEmpty() ? 'false' : 'true'
                    env.SCHEDULE_HOUR = sh(script: 'date -u +%H', returnStdout: true).trim()
                    if (env.TAG_NAME) {
                        env.BUILD_TYPE = 'release'
                        env.VERSION = env.TAG_NAME.replaceFirst('^v', '')
                    } else if (env.BRANCH_NAME == 'beta') {
                        env.BUILD_TYPE = 'beta'
                        env.VERSION = sh(script: 'date -u +%Y%m%d%H%M%S', returnStdout: true).trim() + '-beta'
                    } else if (env.BRANCH_NAME == 'main' || env.BRANCH_NAME == 'master') {
                        env.BUILD_TYPE = 'devel'
                        env.VERSION = sh(script: 'git rev-parse --short=7 HEAD', returnStdout: true).trim()
                    } else {
                        env.BUILD_TYPE = 'branch'
                        env.VERSION = sh(script: 'git rev-parse --short=7 HEAD', returnStdout: true).trim()
                    }
                    env.COMMIT_ID = sh(script: 'git rev-parse --short=7 HEAD', returnStdout: true).trim()
                    env.BUILD_EPOCH = sh(script: 'date -u +%s', returnStdout: true).trim()
                    env.BUILD_DATE = sh(script: "date -u -d @${env.BUILD_EPOCH} +'%Y-%m-%dT%H:%M:%SZ'", returnStdout: true).trim()
                    env.OFFICIAL_SITE = sh(script: '[ -f site.txt ] && cat site.txt || echo "${OFFICIAL_SITE:-}"', returnStdout: true).trim()
                    env.LDFLAGS = "-s -w -X 'main.Version=${env.VERSION}' -X 'main.CommitID=${env.COMMIT_ID}' -X 'main.BuildEpoch=${env.BUILD_EPOCH}' -X 'main.OfficialSite=${env.OFFICIAL_SITE}'"
                }
                sh 'mkdir -p "${GO_CACHE}" "${GO_BUILD}" "${BINDIR}" "${RELDIR}"'
            }
        }

        stage('Lint') {
            when {
                expression { env.IS_SCHEDULED != 'true' || env.SCHEDULE_HOUR == '03' }
            }
            agent { label 'amd64' }
            steps {
                sh '''
                    docker run --rm \
                        --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                        -v "${WORKSPACE}:/app" \
                        -v "${GO_CACHE}:/usr/local/share/go/pkg/mod" \
                        -v "${GO_BUILD}:/usr/local/share/go/cache" \
                        -w /app \
                        -e CGO_ENABLED=0 \
                        -e GOFLAGS=-buildvcs=false \
                        casjaysdev/go:latest \
                        sh -c 'go vet ./... && staticcheck ./...'
                '''
            }
        }

        stage('Test') {
            when {
                expression { env.IS_SCHEDULED != 'true' || env.SCHEDULE_HOUR == '03' }
            }
            agent { label 'amd64' }
            steps {
                sh '''
                    docker run --rm \
                        --name "${PROJECT_NAME}-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                        -v "${WORKSPACE}:/app" \
                        -v "${GO_CACHE}:/usr/local/share/go/pkg/mod" \
                        -v "${GO_BUILD}:/usr/local/share/go/cache" \
                        -w /app \
                        -e CGO_ENABLED=0 \
                        -e GOFLAGS=-buildvcs=false \
                        casjaysdev/go:latest \
                        sh -c 'mkdir -p /tmp/casapps && COVDIR=$(mktemp -d /tmp/casapps/cassocial-XXXXXX) && go test -cover -coverprofile="$COVDIR/coverage.out" ./... && PCT=$(go tool cover -func="$COVDIR/coverage.out" | awk '\''/^total:/ {gsub("%", "", $3); print int($3)}'\'') && echo "Coverage: ${PCT}%" && [ "$PCT" -ge 60 ]'
                '''
            }
        }

        stage('Security') {
            parallel {
                stage('Secret Scan') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-secret-scan-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v "${WORKSPACE}:/repo" \
                                -w /repo \
                                trufflesecurity/trufflehog@sha256:9c7a8325d43a4af2960d3f21d61a08981d8a1bf0542e007efa2e47097e386452 \
                                git file:///repo --results=verified --exclude-paths=.github/trufflehog-exclude.txt --fail
                        '''
                    }
                }
                stage('Workflow Policy') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            fail=0
                            for file in .github/workflows/*.yml .gitea/workflows/*.yml .forgejo/workflows/*.yml; do
                                [ -f "$file" ] || continue
                                bad=$(grep -E '^[[:space:]]*-?[[:space:]]*uses:[[:space:]]*[A-Za-z0-9_.-]+/' "$file" | grep -vE '@[0-9a-f]{40}([[:space:]]|$)' || true)
                                if [ -n "$bad" ]; then
                                    echo "ERROR: unpinned action in $file: $bad"
                                    fail=1
                                fi
                                if grep -qE '^[[:space:]]*(-[[:space:]]*)?pull_request_target[[:space:]]*:' "$file"; then
                                    echo "ERROR: forbidden pull_request_target in $file"
                                    fail=1
                                fi
                            done
                            exit "$fail"
                        '''
                    }
                }
                stage('Vulnerability Scan') {
                    when { expression { fileExists('go.sum') } }
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \
                                --name "${PROJECT_NAME}-vuln-scan-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v "${WORKSPACE}:/app" \
                                -v "${GO_CACHE}:/usr/local/share/go/pkg/mod" \
                                -v "${GO_BUILD}:/usr/local/share/go/cache" \
                                -w /app \
                                casjaysdev/go:latest \
                                govulncheck ./src/...
                        '''
                    }
                }
                stage('License Scan') {
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker run --rm \\
                                --name "${PROJECT_NAME}-license-scan-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \\
                                -v "${WORKSPACE}:/app" \\
                                -v "${GO_CACHE}:/usr/local/share/go/pkg/mod" \\
                                -v "${GO_BUILD}:/usr/local/share/go/cache" \\
                                -w /app \\
                                casjaysdev/go:latest \\
                                sh -c 'go-licenses csv ./... > /tmp/cassocial-licenses.csv && if grep -qiE "GPL|AGPL|LGPL" /tmp/cassocial-licenses.csv; then echo "ERROR: copyleft license detected" >&2; exit 1; fi'
                        '''
                    }
                }
                stage('Image Scan') {
                    when { expression { fileExists('docker/Dockerfile') } }
                    agent { label 'amd64' }
                    steps {
                        sh '''
                            docker build -f docker/Dockerfile -t cassocial:ci .
                            docker run --rm \
                                --name "${PROJECT_NAME}-trivy-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                                -v /var/run/docker.sock:/var/run/docker.sock \
                                aquasec/trivy@sha256:be1190afcb28352bfddc4ddeb71470835d16462af68d310f9f4bca710961a41e \
                                image --exit-code 1 --severity CRITICAL,HIGH cassocial:ci
                        '''
                    }
                    post {
                        always {
                            sh 'docker image rm cassocial:ci || true'
                        }
                    }
                }
            }
        }

        stage('Build') {
            when {
                expression { env.IS_SCHEDULED != 'true' || env.SCHEDULE_HOUR == '03' }
            }
            agent { label 'amd64' }
            steps {
                script {
                    def targets = [
                        [goos: 'linux', goarch: 'amd64'],
                        [goos: 'linux', goarch: 'arm64'],
                        [goos: 'darwin', goarch: 'amd64'],
                        [goos: 'darwin', goarch: 'arm64'],
                        [goos: 'windows', goarch: 'amd64'],
                        [goos: 'windows', goarch: 'arm64'],
                        [goos: 'freebsd', goarch: 'amd64'],
                        [goos: 'freebsd', goarch: 'arm64']
                    ]
                    def builds = [:]
                    targets.each { target ->
                        def current = target
                        def name = "${current.goos}-${current.goarch}"
                        builds[name] = {
                            sh """
                                docker run --rm \\
                                    --name \"\${PROJECT_NAME}-${name}-\$(tr -dc 'a-z0-9' </dev/urandom | head -c8)\" \\
                                    -v \"\${WORKSPACE}:/app\" \\
                                    -v \"\${GO_CACHE}:/usr/local/share/go/pkg/mod\" \\
                                    -v \"\${GO_BUILD}:/usr/local/share/go/cache\" \\
                                    -w /app \\
                                    -e CGO_ENABLED=0 \\
                                    -e GOFLAGS=-buildvcs=false \\
                                    -e GOOS=${current.goos} \\
                                    -e GOARCH=${current.goarch} \\
                                    -e PROJECT_NAME=\"\${PROJECT_NAME}\" \\
                                    -e LDFLAGS=\"\${LDFLAGS}\" \\
                                    casjaysdev/go:latest \\
                                    sh -c 'suffix=""; [ "$GOOS" != windows ] || suffix=".exe"; go build -buildvcs=false -trimpath -ldflags "$LDFLAGS" -o "binaries/$PROJECT_NAME-$GOOS-$GOARCH$suffix" ./src && go build -buildvcs=false -trimpath -ldflags "$LDFLAGS" -o "binaries/$PROJECT_NAME-cli-$GOOS-$GOARCH$suffix" ./src/client'
                            """
                        }
                    }
                    parallel builds
                }
                sh '''
                    docker run --rm \
                        --name "${PROJECT_NAME}-sbom-$(tr -dc 'a-z0-9' </dev/urandom | head -c8)" \
                        -v "${WORKSPACE}:/app" \
                        -v "${GO_CACHE}:/usr/local/share/go/pkg/mod" \
                        -v "${GO_BUILD}:/usr/local/share/go/cache" \
                        -w /app \
                        casjaysdev/go:latest \
                        cyclonedx-gomod mod -json -output "${BINDIR}/${PROJECT_NAME}-sbom.cdx.json"
                '''
                stash name: 'built-binaries', includes: 'binaries/**'
            }
        }

        stage('Release Artifacts') {
            when {
                expression {
                    (env.BUILD_TYPE == 'release' || env.BUILD_TYPE == 'beta') ||
                        (env.BUILD_TYPE == 'devel' && (env.IS_SCHEDULED != 'true' || env.SCHEDULE_HOUR == '03'))
                }
            }
            agent { label 'amd64' }
            steps {
                unstash 'built-binaries'
                sh '''
                    mkdir -p "${RELDIR}"
                    cp binaries/* "${RELDIR}/"
                    printf '%s\\n' "${VERSION}" > "${RELDIR}/version.txt"
                    git log -n 50 --format='* %s (%h)' > "${RELDIR}/release-notes.md"
                    tar --exclude='.git' --exclude='.github' --exclude='.gitea' \\
                        --exclude='.forgejo' --exclude='binaries' --exclude='releases' \\
                        --exclude='*.tar.gz' \\
                        -czf "${RELDIR}/${PROJECT_NAME}-${VERSION}-source.tar.gz" .
                    cd "${RELDIR}"
                    find . -maxdepth 1 -type f ! -name sha256sums.txt -print0 | sort -z | xargs -0 sha256sum > sha256sums.txt
                '''
                archiveArtifacts artifacts: 'releases/*', fingerprint: true, onlyIfSuccessful: true
            }
        }

        stage('Docker Standard') {
            when {
                expression { env.CHANGE_ID == null && env.IS_SCHEDULED != 'true' && fileExists('docker/Dockerfile') }
            }
            agent { label 'amd64' }
            steps {
                withCredentials([string(credentialsId: 'github-token', variable: 'REGISTRY_TOKEN')]) {
                    sh '''
                        echo "$REGISTRY_TOKEN" | docker login ghcr.io -u "$PROJECT_ORG" --password-stdin
                        docker buildx create --name cassocial-builder --use 2>/dev/null || docker buildx use cassocial-builder
                        if [ "$BUILD_TYPE" = release ]; then
                            yymm="$(date -u +%y%m)"
                            set -- -t "$REGISTRY:$VERSION" -t "$REGISTRY:latest" -t "$REGISTRY:$yymm" -t "$REGISTRY:$COMMIT_ID"
                        elif [ "$BUILD_TYPE" = beta ]; then
                            set -- -t "$REGISTRY:beta" -t "$REGISTRY:$COMMIT_ID"
                        else
                            set -- -t "$REGISTRY:$COMMIT_ID"
                        fi
                        docker buildx build -f docker/Dockerfile --platform linux/amd64,linux/arm64 \\
                            --build-arg VERSION="$VERSION" \\
                            --build-arg COMMIT_ID="$COMMIT_ID" \\
                            --build-arg BUILD_DATE="$BUILD_DATE" \\
                            --build-arg BUILD_EPOCH="$BUILD_EPOCH" \\
                            --build-arg OFFICIAL_SITE="$OFFICIAL_SITE" \\
                            --label "org.opencontainers.image.vendor=$PROJECT_ORG" \\
                            --label "org.opencontainers.image.authors=$PROJECT_ORG" \\
                            --label "org.opencontainers.image.title=$PROJECT_NAME" \\
                            --label "org.opencontainers.image.licenses=MIT" \\
                            --label "org.opencontainers.image.version=$VERSION" \\
                            --label "org.opencontainers.image.created=$BUILD_DATE" \\
                            --label "org.opencontainers.image.revision=$COMMIT_ID" \\
                            $@ --push .
                    '''
                }
            }
        }

        stage('Docker Devel') {
            when {
                beforeAgent true
                allOf {
                    expression { env.CHANGE_ID == null && fileExists('docker/Dockerfile.dev') }
                    anyOf {
                        expression { env.BRANCH_NAME == 'main' || env.BRANCH_NAME == 'master' }
                        expression { env.IS_SCHEDULED == 'true' && env.SCHEDULE_HOUR == '04' }
                    }
                }
            }
            agent { label 'amd64' }
            steps {
                withCredentials([string(credentialsId: 'github-token', variable: 'REGISTRY_TOKEN')]) {
                    sh '''
                        echo "$REGISTRY_TOKEN" | docker login ghcr.io -u "$PROJECT_ORG" --password-stdin
                        docker buildx create --name cassocial-builder --use 2>/dev/null || docker buildx use cassocial-builder
                        docker buildx build -f docker/Dockerfile.dev --platform linux/amd64,linux/arm64 \\
                            --build-arg VERSION="$COMMIT_ID" \\
                            --build-arg COMMIT_ID="$COMMIT_ID" \\
                            --build-arg BUILD_DATE="$BUILD_DATE" \\
                            --build-arg BUILD_EPOCH="$BUILD_EPOCH" \\
                            --build-arg OFFICIAL_SITE="$OFFICIAL_SITE" \\
                            --label "org.opencontainers.image.vendor=$PROJECT_ORG" \\
                            --label "org.opencontainers.image.authors=$PROJECT_ORG" \\
                            --label "org.opencontainers.image.title=$PROJECT_NAME-devel" \\
                            --label "org.opencontainers.image.licenses=MIT" \\
                            --label "org.opencontainers.image.version=$COMMIT_ID" \\
                            --label "org.opencontainers.image.created=$BUILD_DATE" \\
                            --label "org.opencontainers.image.revision=$COMMIT_ID" \\
                            -t "$REGISTRY:devel" --push .
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
