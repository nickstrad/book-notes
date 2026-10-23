# Source this for every E09 lab/client command (isolated e09 stack).
export COMPOSE_PROJECT_NAME=queuelet-e09
export QUEUELET_IMAGE_TAG=e09
export QUEUELET_API_PORT=18300
export QUEUELET_HOOK_PORT=18301
export QUEUELET_PG_PORT=15533
export QUEUELET_URL=http://localhost:18300
export QUEUELET_HOOK_URL=http://localhost:18301
export QUEUELET_DATABASE_URL=postgres://queuelet:queuelet@localhost:15533/queuelet?sslmode=disable
