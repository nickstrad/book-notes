# E08 isolated stack (adapted from inherited 07-state-it/tools/env.sh; ports/names changed).
# Source before any lab/client command. Compose file: books/explore-it/project/deploy/compose.yml in THIS export.
export COMPOSE_PROJECT_NAME=queuelet-e08 QUEUELET_IMAGE_TAG=e08 QUEUELET_API_PORT=18290 \
  QUEUELET_HOOK_PORT=18291 QUEUELET_PG_PORT=15523 QUEUELET_URL=http://localhost:18290 \
  QUEUELET_HOOK_URL=http://localhost:18291
