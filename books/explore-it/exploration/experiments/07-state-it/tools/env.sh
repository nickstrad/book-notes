# E07 isolated stack. Source before any lab/client command.
export COMPOSE_PROJECT_NAME=queuelet-e07 QUEUELET_IMAGE_TAG=e07 QUEUELET_API_PORT=18280 \
  QUEUELET_HOOK_PORT=18281 QUEUELET_PG_PORT=15513 QUEUELET_URL=http://localhost:18280 \
  QUEUELET_HOOK_URL=http://localhost:18281
